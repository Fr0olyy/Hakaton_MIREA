import pandas as pd


def assign_object_status(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    default_values = {
        "duplicate_score": 0.0,
        "quality_score": 1.0,
        "label_error_probability": 0.0,
        "entropy": 0.0,
        "uncertainty_score": 0.0,
        "class_deficit_score": 0.0,
        "object_utility_score": 0.0,
    }

    for column, default_value in default_values.items():
        if column not in df.columns:
            df[column] = default_value

        df[column] = pd.to_numeric(df[column], errors="coerce").fillna(default_value)

    statuses = []

    for _, row in df.iterrows():
        duplicate_score = float(row["duplicate_score"])
        quality_score = float(row["quality_score"])
        label_error_probability = float(row["label_error_probability"])
        entropy = float(row["entropy"])
        uncertainty_score = float(row["uncertainty_score"])
        class_deficit_score = float(row["class_deficit_score"])

        if duplicate_score >= 0.7:
            status = "duplicate"

        elif quality_score < 0.65:
            status = "bad_quality"

        elif label_error_probability >= 0.5:
            status = "suspected_label_error"

        elif entropy >= 0.65 or uncertainty_score >= 0.45:
            status = "hard_example"

        elif class_deficit_score >= 0.5:
            status = "rare_class_candidate"

        else:
            status = "ok"

        statuses.append(status)

    df["status"] = statuses

    return df