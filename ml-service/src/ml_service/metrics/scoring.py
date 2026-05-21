import pandas as pd


def compute_object_utility_score(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    default_values = {
        "uncertainty_score": 0.0,
        "label_error_probability": 0.0,
        "class_deficit_score": 0.0,
        "quality_score": 0.0,
        "duplicate_score": 0.0,
    }

    for column, default_value in default_values.items():
        if column not in df.columns:
            df[column] = default_value

        df[column] = pd.to_numeric(df[column], errors="coerce").fillna(default_value)

    df["object_utility_score"] = (
        0.30 * df["uncertainty_score"]
        + 0.25 * df["label_error_probability"]
        + 0.20 * df["class_deficit_score"]
        + 0.15 * df["quality_score"]
        - 0.30 * df["duplicate_score"]
    )

    # Ограничиваем score в диапазоне 0–1
    df["object_utility_score"] = df["object_utility_score"].clip(0.0, 1.0)

    df["final_score"] = df["object_utility_score"]

    return df