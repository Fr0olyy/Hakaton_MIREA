import pandas as pd


def compute_imbalance_index(df: pd.DataFrame) -> float:
    class_counts = df["label"].value_counts()
    total_objects = len(df)

    if total_objects == 0 or len(class_counts) == 0:
        return 1.0

    class_ratios = class_counts / total_objects
    expected_ratio = 1 / len(class_counts)

    imbalance_index = sum(abs(ratio - expected_ratio) for ratio in class_ratios) / 2

    return float(imbalance_index)


def compute_dataset_readiness_score(df: pd.DataFrame) -> tuple[pd.DataFrame, dict]:
    df = df.copy()

    default_values = {
        "label_error_probability": 0.0,
        "quality_score": 1.0,
        "duplicate_score": 0.0,
    }

    for column, default_value in default_values.items():
        if column not in df.columns:
            df[column] = default_value

        df[column] = pd.to_numeric(df[column], errors="coerce").fillna(default_value)

    imbalance_index = compute_imbalance_index(df)

    class_balance_score = 1.0 - imbalance_index
    label_quality_score = 1.0 - df["label_error_probability"].mean()
    data_quality_score = df["quality_score"].mean()
    diversity_score = 1.0 - df["duplicate_score"].mean()

    duplicate_ratio = (df["duplicate_score"] > 0).mean()
    low_duplicate_score = 1.0 - duplicate_ratio

    readiness = (
        0.25 * class_balance_score
        + 0.25 * label_quality_score
        + 0.20 * data_quality_score
        + 0.15 * diversity_score
        + 0.15 * low_duplicate_score
    )

    readiness_score = max(0.0, min(100.0, readiness * 100))

    df["dataset_readiness_score"] = readiness_score

    readiness_report = {
        "dataset_readiness_score": float(readiness_score),
        "imbalance_index": float(imbalance_index),
        "class_balance_score": float(class_balance_score),
        "label_quality_score": float(label_quality_score),
        "data_quality_score": float(data_quality_score),
        "diversity_score": float(diversity_score),
        "duplicate_ratio": float(duplicate_ratio),
        "low_duplicate_score": float(low_duplicate_score),
    }

    return df, readiness_report