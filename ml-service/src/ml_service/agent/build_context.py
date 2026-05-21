from pathlib import Path
import json

import pandas as pd


def build_agent_context(
    dataset_report: dict,
    roadmap: list[dict],
    recommendations: list[dict],
    review_queue: pd.DataFrame,
) -> dict:
    top_review_objects = _build_top_review_objects(review_queue)

    agent_context = {
        "dataset_summary": {
            "total_objects": dataset_report.get("total_objects", 0),
            "dataset_v2_objects": dataset_report.get("dataset_v2_objects", 0),
            "classes_count": dataset_report.get("classes_count", 0),
            "dataset_readiness_score": dataset_report.get("dataset_readiness_score", 0.0),
            "duplicates_count": dataset_report.get("duplicates_count", 0),
            "bad_quality_count": dataset_report.get("bad_quality_count", 0),
            "suspected_label_errors_count": dataset_report.get("suspected_label_errors_count", 0),
            "hard_examples_count": dataset_report.get("hard_examples_count", 0),
            "rare_class_candidates_count": dataset_report.get("rare_class_candidates_count", 0),
            "review_queue_count": dataset_report.get("review_queue_count", 0),
        },
        "class_distribution": dataset_report.get("class_distribution", {}),
        "status_distribution": dataset_report.get("status_distribution", {}),
        "main_problems": dataset_report.get("main_problems", []),
        "readiness_components": dataset_report.get("readiness_components", {}),
        "top_recommendations": recommendations[:5],
        "top_roadmap_items": roadmap[:5],
        "top_review_objects": top_review_objects,
    }

    return agent_context


def _build_top_review_objects(review_queue: pd.DataFrame, limit: int = 10) -> list[dict]:
    if review_queue is None or len(review_queue) == 0:
        return []

    queue = review_queue.copy()

    if "object_utility_score" in queue.columns:
        queue["object_utility_score"] = pd.to_numeric(
            queue["object_utility_score"],
            errors="coerce",
        ).fillna(0.0)

        queue = queue.sort_values("object_utility_score", ascending=False)

    objects = []

    preferred_columns = [
        "id",
        "file_path",
        "label",
        "predicted_label",
        "confidence",
        "status",
        "reasons",
        "recommendation",
        "object_utility_score",
        "entropy",
        "uncertainty_score",
        "label_error_probability",
        "quality_score",
        "duplicate_score",
        "class_deficit_score",
    ]

    for _, row in queue.head(limit).iterrows():
        item = {}

        for column in preferred_columns:
            if column in queue.columns:
                item[column] = _safe_json_value(row[column])

        objects.append(item)

    return objects


def _safe_json_value(value):
    if pd.isna(value):
        return None

    if hasattr(value, "item"):
        return value.item()

    return value


def export_agent_context_json(
    agent_context: dict,
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    context_path = output_dir / "agent_context.json"

    with context_path.open("w", encoding="utf-8") as file:
        json.dump(agent_context, file, ensure_ascii=False, indent=2)

    return context_path