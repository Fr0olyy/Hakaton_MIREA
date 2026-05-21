import pandas as pd


def generate_object_reasons(df: pd.DataFrame) -> pd.DataFrame:
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

    if "quality_reasons" not in df.columns:
        df["quality_reasons"] = ""

    if "predicted_label" not in df.columns:
        df["predicted_label"] = df["label"]

    all_reasons = []

    for _, row in df.iterrows():
        reasons = []

        duplicate_score = float(row["duplicate_score"])
        quality_score = float(row["quality_score"])
        label_error_probability = float(row["label_error_probability"])
        entropy = float(row["entropy"])
        uncertainty_score = float(row["uncertainty_score"])
        class_deficit_score = float(row["class_deficit_score"])
        object_utility_score = float(row["object_utility_score"])

        label = str(row.get("label", "")).strip()
        predicted_label = str(row.get("predicted_label", "")).strip()

        quality_reasons = str(row.get("quality_reasons", "")).strip()

        if duplicate_score >= 1.0:
            reasons.append("duplicate")
        elif duplicate_score >= 0.7:
            reasons.append("possible_duplicate")

        if quality_score < 0.65:
            reasons.append("bad_quality")

        if quality_reasons:
            for reason in quality_reasons.split(";"):
                reason = reason.strip()
                if reason:
                    reasons.append(reason)

        if label and predicted_label and label != predicted_label:
            reasons.append("model_disagreement")

        if label_error_probability >= 0.5:
            reasons.append("suspected_label_error")

        if entropy >= 0.65:
            reasons.append("high_entropy")

        if uncertainty_score >= 0.45:
            reasons.append("low_confidence")

        if class_deficit_score >= 0.5:
            reasons.append("rare_class")
            reasons.append("class_deficit")

        if object_utility_score >= 0.5:
            reasons.append("valuable_for_training")

        unique_reasons = []
        for reason in reasons:
            if reason not in unique_reasons:
                unique_reasons.append(reason)

        all_reasons.append(";".join(unique_reasons))

    df["reasons"] = all_reasons

    return df