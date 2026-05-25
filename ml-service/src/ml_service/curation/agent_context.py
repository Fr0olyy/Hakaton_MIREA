from __future__ import annotations

from pathlib import Path
from typing import Any

import json


def build_level2_agent_context(
    adapter_output: dict[str, Any],
    output_dir: str | Path,
) -> dict[str, Any]:
    output_dir = Path(output_dir)

    dataset_metrics = adapter_output.get("dataset_metrics", {})
    object_metrics = adapter_output.get("object_metrics", [])
    review_queue = adapter_output.get("review_queue", [])
    recommendations = adapter_output.get("recommendations", [])
    roadmap = adapter_output.get("roadmap", [])
    class_action_plan = adapter_output.get("class_action_plan", [])
    collection_tasks = adapter_output.get("collection_tasks", [])
    synthetic_tasks = adapter_output.get("synthetic_tasks", [])
    exports = adapter_output.get("exports", {})

    context = {
        "project_summary": {
            "modality": adapter_output.get("modality"),
            "task_type": adapter_output.get("task_type"),
            "status": adapter_output.get("status"),
            "objects_count": len(object_metrics),
            "review_queue_count": len(review_queue),
            "recommendations_count": len(recommendations),
            "class_action_plan_count": len(class_action_plan),
            "collection_tasks_count": len(collection_tasks),
            "synthetic_tasks_count": len(synthetic_tasks),
        },
        "dataset_metrics": dataset_metrics,
        "top_problems": _extract_top_problems(adapter_output),
        "review_queue_summary": _build_review_queue_summary(review_queue),
        "recommended_actions_summary": _build_actions_summary(object_metrics),
        "class_action_plan": class_action_plan,
        "collection_tasks": collection_tasks,
        "synthetic_tasks": synthetic_tasks,
        "recommendations": recommendations,
        "roadmap": roadmap,
        "dataset_strategies": _load_dataset_strategy_summary(output_dir),
        "exports": exports,
        "agent_rules": {
            "do_not_invent_facts": True,
            "use_only_provided_context": True,
            "if_data_missing_say_not_enough_data": True,
            "do_not_recalculate_ml_metrics": True,
        },
        "suggested_questions": [
            "What are the main dataset problems?",
            "Which objects should be reviewed first?",
            "Which classes need more real data?",
            "Where is synthetic data useful?",
            "Which dataset_v2 strategy is safest?",
        ],
    }

    return context


def save_level2_agent_context(
    adapter_output: dict[str, Any],
    output_dir: str | Path,
) -> str:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    context = build_level2_agent_context(adapter_output, output_dir)
    path = output_dir / "agent_context.json"

    with path.open("w", encoding="utf-8") as file:
        json.dump(context, file, ensure_ascii=False, indent=2)

    return str(path)


def _extract_top_problems(adapter_output: dict[str, Any]) -> list[str]:
    dataset_metrics = adapter_output.get("dataset_metrics", {})
    modality = adapter_output.get("modality")

    problems = []

    if modality == "image_classification":
        problems.extend(dataset_metrics.get("main_problems", []))

        if dataset_metrics.get("suspected_label_errors_count", 0) > 0:
            problems.append("suspected_label_errors")

        if dataset_metrics.get("hard_examples_count", 0) > 0:
            problems.append("hard_examples")

        if dataset_metrics.get("duplicates_count", 0) > 0:
            problems.append("duplicates")

        if dataset_metrics.get("bad_quality_count", 0) > 0:
            problems.append("bad_quality")

    elif modality == "tabular_classification":
        if dataset_metrics.get("rows_with_any_nan_fraction", 0) > 0:
            problems.append("missing_values")

        if dataset_metrics.get("outliers_total_count", 0) > 0:
            problems.append("numeric_outliers")

        if dataset_metrics.get("high_cardinality_columns_count", 0) > 0:
            problems.append("high_cardinality_columns")

    elif modality == "image_detection_yolo":
        if dataset_metrics.get("out_of_bounds_files_count", 0) > 0:
            problems.append("invalid_bboxes")

        if dataset_metrics.get("tiny_boxes_files_count", 0) > 0:
            problems.append("tiny_boxes")

    return sorted(set(problems))


def _build_review_queue_summary(review_queue: list[dict[str, Any]]) -> dict[str, Any]:
    status_counts: dict[str, int] = {}
    risk_counts: dict[str, int] = {}
    action_counts: dict[str, int] = {}

    for item in review_queue:
        status = str(item.get("status", "unknown")).strip()
        risk = str(item.get("action_risk", "unknown")).strip()
        action = str(item.get("recommended_action", "unknown")).strip()

        status_counts[status] = status_counts.get(status, 0) + 1
        risk_counts[risk] = risk_counts.get(risk, 0) + 1
        action_counts[action] = action_counts.get(action, 0) + 1

    return {
        "total": len(review_queue),
        "status_counts": status_counts,
        "risk_counts": risk_counts,
        "recommended_action_counts": action_counts,
        "top_items": review_queue[:10],
    }


def _build_actions_summary(object_metrics: list[dict[str, Any]]) -> dict[str, int]:
    counts: dict[str, int] = {}

    for item in object_metrics:
        action = str(item.get("recommended_action", "unknown")).strip()
        counts[action] = counts.get(action, 0) + 1

    return counts


def _load_dataset_strategy_summary(output_dir: Path) -> Any:
    path = output_dir / "dataset_strategy_summary.json"

    if not path.exists():
        return {}

    try:
        with path.open("r", encoding="utf-8") as file:
            data = json.load(file)

        if isinstance(data, list) and data:
            return data[0]

        return data

    except Exception:
        return {}