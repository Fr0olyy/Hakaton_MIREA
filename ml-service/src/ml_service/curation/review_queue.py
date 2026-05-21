from pathlib import Path

import pandas as pd


REVIEW_STATUSES = {
    "suspected_label_error",
    "hard_example",
    "rare_class_candidate",
    "bad_quality",
    "duplicate",
    "needs_review",
}


def build_review_queue(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    if "status" not in df.columns:
        raise ValueError("Cannot build review queue: column 'status' is missing")

    if "object_utility_score" not in df.columns:
        df["object_utility_score"] = 0.0

    queue = df[df["status"].isin(REVIEW_STATUSES)].copy()

    queue["object_utility_score"] = pd.to_numeric(
        queue["object_utility_score"],
        errors="coerce",
    ).fillna(0.0)

    queue = queue.sort_values(
        by="object_utility_score",
        ascending=False,
    )

    return queue


def export_review_queue_csv(
    review_queue: pd.DataFrame,
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    review_queue_path = output_dir / "review_queue.csv"

    preferred_columns = [
        "id",
        "file_path",
        "resolved_path",
        "label",
        "predicted_label",
        "confidence",
        "status",
        "reasons",
        "recommendation",

        "object_utility_score",
        "final_score",

        "entropy",
        "uncertainty_score",
        "label_error_probability",

        "quality_score",
        "quality_reasons",
        "blur_score",
        "brightness",
        "contrast",
        "width",
        "height",

        "duplicate_score",
        "duplicate_group_id",
        "duplicate_neighbor_id",
        "duplicate_distance",

        "class_deficit_score",
        "dataset_readiness_score",
    ]

    probability_columns = [
        column for column in review_queue.columns
        if column.startswith("prob_")
    ]

    export_columns = []

    for column in preferred_columns:
        if column in review_queue.columns:
            export_columns.append(column)

    export_columns.extend(probability_columns)

    review_queue[export_columns].to_csv(review_queue_path, index=False)

    return review_queue_path