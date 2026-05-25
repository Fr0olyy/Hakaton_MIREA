from pathlib import Path

import pandas as pd


def load_dataset(dataset_path: str | Path) -> pd.DataFrame:
    dataset_path = Path(dataset_path)

    if not dataset_path.exists():
        raise FileNotFoundError(f"Dataset file not found: {dataset_path}")

    df = pd.read_csv(dataset_path, skipinitialspace=True)

    df.columns = df.columns.astype(str).str.strip()

    required_cols = {"id", "file_path", "label"}
    missing = required_cols - set(df.columns)

    if missing:
        raise ValueError(
            f"В датасете не хватает обязательных колонок: {missing}"
        )

    for col in df.columns:
        if df[col].dtype == "object" or str(df[col].dtype) == "string":
            df[col] = _clean_text_column(df[col])

    for col in ["id", "file_path", "label"]:
        empty_mask = df[col].isna() | (df[col].astype(str).str.strip() == "")

        if empty_mask.any():
            bad_rows = df[empty_mask].index.tolist()[:10]
            raise ValueError(
                f"Column '{col}' contains empty values. Example rows: {bad_rows}"
            )

    df["id"] = df["id"].astype(str).str.strip()
    df["file_path"] = _clean_text_column(df["file_path"])
    df["label"] = _clean_label_column(df["label"])

    if "split" not in df.columns:
        df["split"] = "train"
    else:
        df["split"] = _clean_text_column(df["split"])
        df["split"] = df["split"].fillna("train")
        df.loc[df["split"] == "", "split"] = "train"

    if "predicted_label" not in df.columns:
        df["predicted_label"] = df["label"]
    else:
        df["predicted_label"] = _clean_label_column(df["predicted_label"])
        df["predicted_label"] = df["predicted_label"].fillna(df["label"])
        df.loc[df["predicted_label"] == "", "predicted_label"] = df["label"]

    if "confidence" in df.columns:
        df["confidence"] = pd.to_numeric(df["confidence"], errors="coerce")

    return df


def _clean_text_column(series: pd.Series) -> pd.Series:
    return (
        series
        .astype("string")
        .str.replace("\u00a0", " ", regex=False)
        .str.replace(r"\s+", " ", regex=True)
        .str.strip()
    )


def _clean_label_column(series: pd.Series) -> pd.Series:
    return (
        series
        .astype("string")
        .str.replace("\u00a0", " ", regex=False)
        .str.replace(r"\s+", " ", regex=True)
        .str.strip()
        .str.lower()
    )