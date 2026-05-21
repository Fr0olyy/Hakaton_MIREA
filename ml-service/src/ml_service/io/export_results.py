from pathlib import Path

import pandas as pd


def export_results_csv(df: pd.DataFrame, output_dir: str | Path) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    results_path = output_dir / "results.csv"

    preferred_columns = [
        # base fields
        "id",
        "file_path",
        "resolved_path",
        "label",
        "predicted_label",
        "confidence",
        "split",
        "source",
        "annotator",

        # validation
        "file_exists",
        "image_opened",
        "image_error",

        # probabilities / uncertainty
        "uncertainty_score",
        "entropy",

        # class balance
        "class_deficit_score",

        # image quality
        "width",
        "height",
        "blur_score",
        "brightness",
        "contrast",
        "resolution_score",
        "corrupted_file",
        "quality_score",
        "quality_reasons",

        # duplicates
        "phash",
        "duplicate_score",
        "duplicate_group_id",
        "duplicate_neighbor_id",
        "duplicate_distance",

        # label quality
        "label_error_probability",

        # scoring
        "object_utility_score",
        "final_score",

        # curation
        "status",
        "reasons",
        "recommendation",

        # dataset-level repeated value
        "dataset_readiness_score",
    ]

    # prob_* колонки добавляем отдельно, потому что их имена зависят от классов
    probability_columns = [
        column for column in df.columns
        if column.startswith("prob_")
    ]

    export_columns = []

    for column in preferred_columns:
        if column in df.columns:
            export_columns.append(column)

    export_columns.extend(probability_columns)

    result_df = df[export_columns].copy()

    result_df.to_csv(results_path, index=False)

    return results_path