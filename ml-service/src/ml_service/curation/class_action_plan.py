from __future__ import annotations

from collections import Counter, defaultdict
from typing import Any


def build_class_action_plan(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    modality = str(adapter_output.get("modality", "unknown"))

    if modality == "image_classification":
        return _build_image_class_action_plan(adapter_output)

    if modality == "tabular_classification":
        return _build_tabular_action_plan(adapter_output)

    if modality == "image_detection_yolo":
        return _build_yolo_action_plan(adapter_output)

    return _build_default_action_plan(adapter_output)


def _build_image_class_action_plan(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    dataset_metrics = adapter_output.get("dataset_metrics", {})
    object_metrics = adapter_output.get("object_metrics", [])
    review_queue = adapter_output.get("review_queue", [])

    class_distribution = dataset_metrics.get("class_distribution", {})

    if not class_distribution:
        class_distribution = _count_by_field(object_metrics, "label")

    if not class_distribution:
        return []

    max_count = max(class_distribution.values())
    total_objects = sum(class_distribution.values())
    avg_count = total_objects / max(len(class_distribution), 1)

    review_by_class = _group_count_by_field(review_queue, "label")
    label_error_by_class = _group_count_by_status(object_metrics, "label", "suspected_label_error")
    hard_by_class = _group_count_by_status(object_metrics, "label", "hard_example")
    bad_quality_by_class = _group_count_by_status(object_metrics, "label", "bad_quality")
    duplicate_by_class = _group_count_by_status(object_metrics, "label", "duplicate")

    actions_by_class = _group_actions_by_class(object_metrics)

    plan = []

    for class_name, count in sorted(class_distribution.items()):
        class_name = str(class_name).strip()

        class_deficit_score = 1.0 - (count / max_count) if max_count > 0 else 0.0
        count_ratio_to_avg = count / avg_count if avg_count > 0 else 1.0

        review_count = review_by_class.get(class_name, 0)
        label_error_count = label_error_by_class.get(class_name, 0)
        hard_count = hard_by_class.get(class_name, 0)
        bad_quality_count = bad_quality_by_class.get(class_name, 0)
        duplicate_count = duplicate_by_class.get(class_name, 0)

        class_actions = actions_by_class.get(class_name, Counter())

        recommended_action, priority, reason, risk = _decide_image_class_action(
            count=count,
            count_ratio_to_avg=count_ratio_to_avg,
            class_deficit_score=class_deficit_score,
            review_count=review_count,
            label_error_count=label_error_count,
            hard_count=hard_count,
            bad_quality_count=bad_quality_count,
            duplicate_count=duplicate_count,
            class_actions=class_actions,
        )

        target_count = _estimate_target_count(
            count=count,
            avg_count=avg_count,
            recommended_action=recommended_action,
        )

        plan.append(
            {
                "target": class_name,
                "target_type": "class",
                "objects_count": int(count),
                "target_count": int(target_count),
                "class_deficit_score": round(float(class_deficit_score), 4),
                "count_ratio_to_avg": round(float(count_ratio_to_avg), 4),
                "review_objects_count": int(review_count),
                "label_error_count": int(label_error_count),
                "hard_examples_count": int(hard_count),
                "bad_quality_count": int(bad_quality_count),
                "duplicate_count": int(duplicate_count),
                "recommended_action": recommended_action,
                "priority": priority,
                "reason": reason,
                "risk": risk,
                "action_stats": dict(class_actions),
            }
        )

    priority_order = {"high": 0, "medium": 1, "low": 2}
    plan.sort(
        key=lambda item: (
            priority_order.get(item["priority"], 9),
            -item["class_deficit_score"],
            -item["review_objects_count"],
        )
    )

    return plan


def _decide_image_class_action(
    count: int,
    count_ratio_to_avg: float,
    class_deficit_score: float,
    review_count: int,
    label_error_count: int,
    hard_count: int,
    bad_quality_count: int,
    duplicate_count: int,
    class_actions: Counter,
) -> tuple[str, str, str, str]:
    if label_error_count >= max(5, count * 0.08):
        return (
            "review_labels",
            "high",
            "Class has many suspected label errors and needs annotation review.",
            "high",
        )

    if class_deficit_score >= 0.45 or count_ratio_to_avg < 0.65:
        return (
            "collect_more_real_data_and_generate_synthetic",
            "high",
            "Class is strongly underrepresented. Collect real samples and prepare synthetic candidates.",
            "medium",
        )

    if class_deficit_score >= 0.25 or count_ratio_to_avg < 0.8:
        return (
            "collect_more_real_data",
            "medium",
            "Class is moderately underrepresented and needs more real examples.",
            "low",
        )

    if hard_count >= max(5, count * 0.08):
        return (
            "add_hard_examples_to_next_train",
            "medium",
            "Class contains many hard examples that may be useful for the next training iteration.",
            "medium",
        )

    if bad_quality_count >= max(3, count * 0.05):
        return (
            "review_bad_quality_samples",
            "medium",
            "Class contains low-quality samples that should be reviewed before training.",
            "medium",
        )

    if duplicate_count >= max(3, count * 0.05):
        return (
            "remove_or_merge_duplicates",
            "medium",
            "Class contains duplicates that may reduce dataset diversity.",
            "low",
        )

    if class_actions.get("augment_existing_data", 0) > 0:
        return (
            "augment_existing_data",
            "medium",
            "Some objects in this class are good candidates for augmentation.",
            "low",
        )

    if review_count > 0:
        return (
            "review_selected_objects",
            "low",
            "Class has a small number of objects that should be reviewed.",
            "low",
        )

    return (
        "enough_data",
        "low",
        "Class looks balanced and has no critical issues.",
        "low",
    )


def _estimate_target_count(
    count: int,
    avg_count: float,
    recommended_action: str,
) -> int:
    if recommended_action == "collect_more_real_data_and_generate_synthetic":
        return max(10, int(avg_count - count))

    if recommended_action == "collect_more_real_data":
        return max(5, int((avg_count - count) * 0.75))

    if recommended_action == "augment_existing_data":
        return max(5, int(count * 0.15))

    return 0


def _build_tabular_action_plan(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    metrics = adapter_output.get("dataset_metrics", {})

    plan = []

    if metrics.get("rows_with_any_nan_fraction", 0.0) > 0:
        plan.append(
            {
                "target": "missing_values",
                "target_type": "tabular_issue",
                "recommended_action": "fix_missing_values",
                "priority": "medium",
                "reason": "Some rows contain missing values.",
                "risk": "medium",
                "affected_fraction": metrics.get("rows_with_any_nan_fraction", 0.0),
            }
        )

    if metrics.get("outliers_total_count", 0) > 0:
        plan.append(
            {
                "target": "numeric_outliers",
                "target_type": "tabular_issue",
                "recommended_action": "review_outliers",
                "priority": "medium",
                "reason": "Some numeric values look like statistical outliers.",
                "risk": "medium",
                "affected_count": metrics.get("outliers_total_count", 0),
            }
        )

    if metrics.get("high_cardinality_columns_count", 0) > 0:
        plan.append(
            {
                "target": "high_cardinality_columns",
                "target_type": "tabular_issue",
                "recommended_action": "review_or_drop_high_cardinality_columns",
                "priority": "medium",
                "reason": "Some columns look like IDs or have too many unique values.",
                "risk": "medium",
                "affected_count": metrics.get("high_cardinality_columns_count", 0),
            }
        )

    if not plan:
        plan.append(
            {
                "target": "tabular_dataset",
                "target_type": "dataset",
                "recommended_action": "ready_for_training",
                "priority": "low",
                "reason": "No critical tabular issues found.",
                "risk": "low",
            }
        )

    return plan


def _build_yolo_action_plan(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    metrics = adapter_output.get("dataset_metrics", {})
    plan = []

    if metrics.get("out_of_bounds_files_count", 0) > 0:
        plan.append(
            {
                "target": "invalid_bboxes",
                "target_type": "detection_annotation_issue",
                "recommended_action": "fix_invalid_bboxes",
                "priority": "high",
                "reason": "Some YOLO bounding boxes are outside normalized 0..1 range.",
                "risk": "high",
                "affected_files_count": metrics.get("out_of_bounds_files_count", 0),
            }
        )

    if metrics.get("tiny_boxes_files_count", 0) > 0:
        plan.append(
            {
                "target": "tiny_boxes",
                "target_type": "detection_annotation_issue",
                "recommended_action": "review_tiny_boxes",
                "priority": "medium",
                "reason": "Some YOLO boxes are too small and may add training noise.",
                "risk": "medium",
                "affected_files_count": metrics.get("tiny_boxes_files_count", 0),
            }
        )

    if not plan:
        plan.append(
            {
                "target": "yolo_annotations",
                "target_type": "dataset",
                "recommended_action": "ready_for_training",
                "priority": "low",
                "reason": "No critical YOLO annotation issues found.",
                "risk": "low",
            }
        )

    return plan


def _build_default_action_plan(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    review_count = len(adapter_output.get("review_queue", []))

    if review_count > 0:
        return [
            {
                "target": "dataset",
                "target_type": "dataset",
                "recommended_action": "review_problematic_objects",
                "priority": "medium",
                "reason": "Dataset contains objects that require review.",
                "risk": "medium",
                "review_objects_count": review_count,
            }
        ]

    return [
        {
            "target": "dataset",
            "target_type": "dataset",
            "recommended_action": "ready_for_training",
            "priority": "low",
            "reason": "No critical issues found.",
            "risk": "low",
        }
    ]


def _count_by_field(items: list[dict[str, Any]], field: str) -> dict[str, int]:
    counter = Counter()

    for item in items:
        value = str(item.get(field, "")).strip()
        if value:
            counter[value] += 1

    return dict(counter)


def _group_count_by_field(items: list[dict[str, Any]], field: str) -> dict[str, int]:
    return _count_by_field(items, field)


def _group_count_by_status(
    items: list[dict[str, Any]],
    group_field: str,
    status_value: str,
) -> dict[str, int]:
    counter = Counter()

    for item in items:
        group_value = str(item.get(group_field, "")).strip()
        status = str(item.get("status", "")).strip()

        if group_value and status == status_value:
            counter[group_value] += 1

    return dict(counter)


def _group_actions_by_class(items: list[dict[str, Any]]) -> dict[str, Counter]:
    result = defaultdict(Counter)

    for item in items:
        class_name = str(item.get("label", "")).strip()
        action = str(item.get("recommended_action", "")).strip()

        if class_name and action:
            result[class_name][action] += 1

    return result