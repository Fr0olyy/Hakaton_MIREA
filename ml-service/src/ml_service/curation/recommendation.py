import pandas as pd


def generate_object_recommendations(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    if "status" not in df.columns:
        raise ValueError("Cannot generate recommendations: column 'status' is missing")

    recommendations = []

    for _, row in df.iterrows():
        status = str(row.get("status", "ok")).strip()

        duplicate_score = _safe_float(row.get("duplicate_score", 0.0))
        quality_score = _safe_float(row.get("quality_score", 1.0))
        label_error_probability = _safe_float(row.get("label_error_probability", 0.0))
        entropy = _safe_float(row.get("entropy", 0.0))
        uncertainty_score = _safe_float(row.get("uncertainty_score", 0.0))
        class_deficit_score = _safe_float(row.get("class_deficit_score", 0.0))

        recommendation = "keep"

        if status == "duplicate" or duplicate_score >= 0.7:
            recommendation = "exclude_duplicate"

        elif status == "bad_quality" or quality_score < 0.65:
            recommendation = "exclude_bad_quality"

        elif status == "suspected_label_error" or label_error_probability >= 0.5:
            recommendation = "recheck_label"

        elif status == "hard_example" or entropy >= 0.65 or uncertainty_score >= 0.45:
            recommendation = "send_to_review"

        elif status == "rare_class_candidate" or class_deficit_score >= 0.5:
            recommendation = "add_to_next_train"

        elif status == "ok":
            recommendation = "keep"

        else:
            recommendation = "keep_but_monitor"

        recommendations.append(recommendation)

    df["recommendation"] = recommendations

    return df


def _safe_float(value, default: float = 0.0) -> float:
    try:
        if pd.isna(value):
            return default
        return float(value)
    except Exception:
        return default