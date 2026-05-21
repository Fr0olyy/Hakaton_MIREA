from pathlib import Path

import pandas as pd


EXCLUDE_STATUSES = {
    "duplicate",
    "bad_quality",
    "suspected_label_error",
    "exclude_candidate",
}


KEEP_STATUSES = {
    "ok",
    "rare_class_candidate",
    "hard_example",
    "add_to_next_train",
}


def build_dataset_v2(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    if "status" not in df.columns:
        raise ValueError("Cannot build dataset_v2: column 'status' is missing")

    dataset_v2 = df[~df["status"].isin(EXCLUDE_STATUSES)].copy()

    # В dataset_v2 не нужно тащить все технические метрики.
    # Это должен быть CSV, который можно снова использовать как датасет.
    preferred_columns = [
        "id",
        "file_path",
        "label",
        "predicted_label",
        "confidence",
        "split",
        "source",
        "annotator",
        "status",
        "reasons",
        "recommendation",
    ]

    export_columns = [
        column for column in preferred_columns
        if column in dataset_v2.columns
    ]

    return dataset_v2[export_columns]


def export_dataset_v2_csv(
    dataset_v2: pd.DataFrame,
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    dataset_v2_path = output_dir / "dataset_v2.csv"

    dataset_v2.to_csv(dataset_v2_path, index=False)

    return dataset_v2_path