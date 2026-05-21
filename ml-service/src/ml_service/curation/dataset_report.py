from pathlib import Path
import json

import pandas as pd


def build_dataset_report(
    df: pd.DataFrame,
    review_queue: pd.DataFrame | None = None,
    dataset_v2: pd.DataFrame | None = None,
    readiness_report: dict | None = None,
) -> dict:
    df = df.copy()

    readiness_report = readiness_report or {}

    total_objects = int(len(df))
    dataset_v2_objects = int(len(dataset_v2)) if dataset_v2 is not None else 0
    review_queue_count = int(len(review_queue)) if review_queue is not None else 0

    class_distribution = _value_counts_dict(df, "label")
    status_distribution = _value_counts_dict(df, "status")

    report = {
        "total_objects": total_objects,
        "dataset_v2_objects": dataset_v2_objects,
        "classes_count": int(df["label"].nunique()) if "label" in df.columns else 0,

        "dataset_readiness_score": float(
            readiness_report.get(
                "dataset_readiness_score",
                _safe_first_value(df, "dataset_readiness_score", 0.0),
            )
        ),

        "duplicates_count": _count_condition(df, "duplicate_score", ">=", 0.7),
        "bad_quality_count": _count_condition(df, "quality_score", "<", 0.65),
        "suspected_label_errors_count": _count_condition(df, "label_error_probability", ">=", 0.5),
        "hard_examples_count": _count_hard_examples(df),
        "rare_class_candidates_count": _count_condition(df, "class_deficit_score", ">=", 0.5),

        "review_queue_count": review_queue_count,

        "class_distribution": class_distribution,
        "status_distribution": status_distribution,

        "main_problems": _detect_main_problems(df, readiness_report),

        "readiness_components": {
            "imbalance_index": float(readiness_report.get("imbalance_index", 0.0)),
            "class_balance_score": float(readiness_report.get("class_balance_score", 0.0)),
            "label_quality_score": float(readiness_report.get("label_quality_score", 0.0)),
            "data_quality_score": float(readiness_report.get("data_quality_score", 0.0)),
            "diversity_score": float(readiness_report.get("diversity_score", 0.0)),
            "duplicate_ratio": float(readiness_report.get("duplicate_ratio", 0.0)),
            "low_duplicate_score": float(readiness_report.get("low_duplicate_score", 0.0)),
        },
    }

    return report


def export_dataset_report_json(
    dataset_report: dict,
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    report_path = output_dir / "dataset_report.json"

    with report_path.open("w", encoding="utf-8") as file:
        json.dump(dataset_report, file, ensure_ascii=False, indent=2)

    return report_path


def _value_counts_dict(df: pd.DataFrame, column: str) -> dict:
    if column not in df.columns:
        return {}

    return {
        str(key): int(value)
        for key, value in df[column].value_counts().to_dict().items()
    }


def _count_condition(
    df: pd.DataFrame,
    column: str,
    operator: str,
    value: float,
) -> int:
    if column not in df.columns:
        return 0

    series = pd.to_numeric(df[column], errors="coerce").fillna(0.0)

    if operator == ">=":
        return int((series >= value).sum())

    if operator == ">":
        return int((series > value).sum())

    if operator == "<=":
        return int((series <= value).sum())

    if operator == "<":
        return int((series < value).sum())

    raise ValueError(f"Unsupported operator: {operator}")


def _count_hard_examples(df: pd.DataFrame) -> int:
    entropy = pd.to_numeric(df.get("entropy", 0.0), errors="coerce").fillna(0.0)
    uncertainty = pd.to_numeric(df.get("uncertainty_score", 0.0), errors="coerce").fillna(0.0)

    return int(((entropy >= 0.65) | (uncertainty >= 0.45)).sum())


def _safe_first_value(df: pd.DataFrame, column: str, default: float) -> float:
    if column not in df.columns or len(df) == 0:
        return default

    value = pd.to_numeric(df[column], errors="coerce").dropna()

    if len(value) == 0:
        return default

    return float(value.iloc[0])


def _detect_main_problems(
    df: pd.DataFrame,
    readiness_report: dict,
) -> list[str]:
    problems = []

    duplicates_ratio = _ratio_condition(df, "duplicate_score", ">=", 0.7)
    bad_quality_ratio = _ratio_condition(df, "quality_score", "<", 0.65)
    label_error_ratio = _ratio_condition(df, "label_error_probability", ">=", 0.5)

    imbalance_index = float(readiness_report.get("imbalance_index", 0.0))

    if label_error_ratio >= 0.03:
        problems.append("label_noise")

    if duplicates_ratio >= 0.03:
        problems.append("duplicates")

    if bad_quality_ratio >= 0.03:
        problems.append("bad_quality")

    if imbalance_index >= 0.2:
        problems.append("class_imbalance")

    hard_examples_count = _count_hard_examples(df)
    if len(df) > 0 and hard_examples_count / len(df) >= 0.05:
        problems.append("many_hard_examples")

    return problems


def _ratio_condition(
    df: pd.DataFrame,
    column: str,
    operator: str,
    value: float,
) -> float:
    if len(df) == 0:
        return 0.0

    return _count_condition(df, column, operator, value) / len(df)