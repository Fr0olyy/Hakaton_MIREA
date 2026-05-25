from __future__ import annotations

from typing import Any


def build_collection_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    modality = str(adapter_output.get("modality", "unknown"))

    if modality == "image_classification":
        return _build_image_collection_tasks(adapter_output)

    if modality == "tabular_classification":
        return _build_tabular_collection_tasks(adapter_output)

    if modality == "image_detection_yolo":
        return _build_yolo_collection_tasks(adapter_output)

    return _build_default_collection_tasks(adapter_output)


def _build_image_collection_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    class_action_plan = adapter_output.get("class_action_plan", [])
    tasks = []

    for item in class_action_plan:
        action = str(item.get("recommended_action", "")).strip()
        target_class = str(item.get("target", "")).strip()
        priority = str(item.get("priority", "medium")).strip()
        risk = str(item.get("risk", "medium")).strip()

        if not target_class:
            continue

        if action == "collect_more_real_data_and_generate_synthetic":
            tasks.append(
                {
                    "task_type": "collect_real_data",
                    "target_class": target_class,
                    "target_count": int(item.get("target_count", 30) or 30),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get(
                        "reason",
                        "Class is underrepresented and needs more real samples.",
                    ),
                    "expected_impact": "Improve class balance and reduce rare-class weakness.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "ML Engineer",
                }
            )

        elif action == "collect_more_real_data":
            tasks.append(
                {
                    "task_type": "collect_real_data",
                    "target_class": target_class,
                    "target_count": int(item.get("target_count", 20) or 20),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get(
                        "reason",
                        "Class needs more real examples.",
                    ),
                    "expected_impact": "Improve class coverage and model generalization.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "ML Engineer",
                }
            )

        elif action == "review_labels":
            tasks.append(
                {
                    "task_type": "review_real_labels",
                    "target_class": target_class,
                    "target_count": int(item.get("label_error_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get(
                        "reason",
                        "Class has many suspected label errors.",
                    ),
                    "expected_impact": "Improve label quality and reduce training noise.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Annotator",
                }
            )

        elif action == "review_bad_quality_samples":
            tasks.append(
                {
                    "task_type": "review_image_quality",
                    "target_class": target_class,
                    "target_count": int(item.get("bad_quality_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get(
                        "reason",
                        "Class contains low-quality images.",
                    ),
                    "expected_impact": "Remove noisy samples and improve dataset reliability.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Annotator",
                }
            )

        elif action == "remove_or_merge_duplicates":
            tasks.append(
                {
                    "task_type": "review_duplicates",
                    "target_class": target_class,
                    "target_count": int(item.get("duplicate_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get(
                        "reason",
                        "Class contains duplicate images.",
                    ),
                    "expected_impact": "Improve dataset diversity and reduce redundant samples.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Data Analyst",
                }
            )

    return tasks


def _build_tabular_collection_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    class_action_plan = adapter_output.get("class_action_plan", [])
    tasks = []

    for item in class_action_plan:
        action = str(item.get("recommended_action", "")).strip()
        target = str(item.get("target", "")).strip()
        priority = str(item.get("priority", "medium")).strip()
        risk = str(item.get("risk", "medium")).strip()

        if action == "fix_missing_values":
            tasks.append(
                {
                    "task_type": "fix_tabular_missing_values",
                    "target": target,
                    "target_count": None,
                    "priority": priority,
                    "status": "open",
                    "reason": item.get("reason", "Rows contain missing values."),
                    "expected_impact": "Improve feature completeness and reduce training instability.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Data Analyst",
                }
            )

        elif action == "review_outliers":
            tasks.append(
                {
                    "task_type": "review_tabular_outliers",
                    "target": target,
                    "target_count": int(item.get("affected_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get("reason", "Dataset contains numeric outliers."),
                    "expected_impact": "Reduce noise and detect possible data entry errors.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Data Analyst",
                }
            )

        elif action == "review_or_drop_high_cardinality_columns":
            tasks.append(
                {
                    "task_type": "review_high_cardinality_features",
                    "target": target,
                    "target_count": int(item.get("affected_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get("reason", "Some columns may behave like IDs."),
                    "expected_impact": "Reduce overfitting risk and improve feature quality.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Data Analyst",
                }
            )

    return tasks


def _build_yolo_collection_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    class_action_plan = adapter_output.get("class_action_plan", [])
    tasks = []

    for item in class_action_plan:
        action = str(item.get("recommended_action", "")).strip()
        target = str(item.get("target", "")).strip()
        priority = str(item.get("priority", "medium")).strip()
        risk = str(item.get("risk", "medium")).strip()

        if action == "fix_invalid_bboxes":
            tasks.append(
                {
                    "task_type": "fix_detection_annotations",
                    "target": target,
                    "target_count": int(item.get("affected_files_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get("reason", "Some bounding boxes are invalid."),
                    "expected_impact": "Improve detection annotation quality and reduce training errors.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Annotator",
                }
            )

        elif action == "review_tiny_boxes":
            tasks.append(
                {
                    "task_type": "review_tiny_detection_boxes",
                    "target": target,
                    "target_count": int(item.get("affected_files_count", 0) or 0),
                    "priority": priority,
                    "status": "open",
                    "reason": item.get("reason", "Some boxes are too small."),
                    "expected_impact": "Reduce noisy annotations and improve detector stability.",
                    "risk": risk,
                    "source_action": action,
                    "owner_role": "Domain Expert",
                }
            )

    return tasks


def _build_default_collection_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    review_queue_count = len(adapter_output.get("review_queue", []))

    if review_queue_count == 0:
        return []

    return [
        {
            "task_type": "review_problematic_objects",
            "target": "dataset",
            "target_count": review_queue_count,
            "priority": "medium",
            "status": "open",
            "reason": "Dataset contains objects that require human review.",
            "expected_impact": "Improve dataset quality before training.",
            "risk": "medium",
            "source_action": "review_problematic_objects",
            "owner_role": "Annotator",
        }
    ]