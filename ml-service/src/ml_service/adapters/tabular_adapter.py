from pathlib import Path
from typing import Any, Dict

import pandas as pd

from ml_service.metrics.tabular.missing import analyze_missing_values
from ml_service.metrics.tabular.duplicates import analyze_duplicates
from ml_service.metrics.tabular.outliers import analyze_outliers
from ml_service.metrics.tabular.leakage import analyze_target_leakage
from ml_service.metrics.tabular.constant import analyze_constant_columns
from ml_service.metrics.tabular.cardinality import analyze_high_cardinality

from ml_service.adapters.base import (
    AdapterInput,
    AdapterOutput,
    make_success_output,
    make_error_output,
)


def analyze_tabular_dataset(
    df: pd.DataFrame,
    target_col: str | None = None,
) -> Dict[str, Any]:
    if df.empty:
        return {
            "tabular_quality_score": 0.0,
            "status": "error",
            "reports": {},
        }

    reports = {
        "missing_values": analyze_missing_values(df),
        "duplicates": analyze_duplicates(df),
        "outliers": analyze_outliers(df),
        "constant_columns": analyze_constant_columns(df),
        "high_cardinality": analyze_high_cardinality(df),
    }

    if target_col:
        reports["target_leakage"] = analyze_target_leakage(df, target_col=target_col)

    score = 100.0

    for _, report_data in reports.items():
        if report_data.get("status") == "warning":
            score -= 15

    score = max(0.0, score)

    return {
        "tabular_quality_score": score,
        "status": "ready" if score >= 70.0 else "needs_review",
        "reports": reports,
    }


def run_tabular_adapter(adapter_input: AdapterInput) -> AdapterOutput:
    try:
        if adapter_input.dataset_path is None:
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error="dataset_path is required for tabular_classification",
            )

        dataset_path = Path(adapter_input.dataset_path)

        if not dataset_path.exists():
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error=f"dataset_path not found: {dataset_path}",
            )

        df = pd.read_csv(dataset_path)

        target_col = adapter_input.extra.get("target_col")
        if not target_col and "target" in df.columns:
            target_col = "target"

        result = analyze_tabular_dataset(df, target_col=target_col)

        dataset_metrics = _build_tabular_dataset_metrics(df, result)
        object_metrics = _build_tabular_object_metrics(df)
        review_queue = [item for item in object_metrics if item["status"] != "ok"]
        recommendations = _build_tabular_recommendations(result)

        return make_success_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            dataset_metrics=dataset_metrics,
            object_metrics=object_metrics,
            review_queue=review_queue,
            recommendations=recommendations,
            roadmap=[],
            collection_tasks=[],
            synthetic_tasks=[],
            class_action_plan=[],
            exports={},
            metadata={
                "dataset_path": str(dataset_path),
                "adapter_status": result.get("status", "unknown"),
                "target_col": target_col,
            },
            warnings=[],
        )

    except Exception as error:
        return make_error_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            error=str(error),
        )


def _build_tabular_dataset_metrics(df: pd.DataFrame, result: dict) -> dict:
    reports = result.get("reports", {})

    missing_report = reports.get("missing_values", {})
    duplicates_report = reports.get("duplicates", {})
    outliers_report = reports.get("outliers", {})
    constant_report = reports.get("constant_columns", {})
    cardinality_report = reports.get("high_cardinality", {})
    leakage_report = reports.get("target_leakage", {})

    outliers_by_column = outliers_report.get("outliers_count", {})
    total_outliers = (
        sum(outliers_by_column.values())
        if isinstance(outliers_by_column, dict)
        else 0
    )

    high_cardinality_columns = cardinality_report.get("high_cardinality_columns", {})
    constant_columns = constant_report.get("constant_columns", [])

    return {
        "total_rows": int(len(df)),
        "total_columns": int(len(df.columns)),
        "tabular_quality_score": float(result.get("tabular_quality_score", 0.0)),
        "adapter_status": result.get("status", "unknown"),
        "missing_columns_count": int(len(missing_report.get("missing_fractions", {}))),
        "rows_with_any_nan_fraction": float(
            missing_report.get("rows_with_any_nan_fraction", 0.0)
        ),
        "duplicate_rows_count": int(duplicates_report.get("duplicate_count", 0)),
        "duplicate_rows_fraction": float(duplicates_report.get("duplicate_fraction", 0.0)),
        "outliers_total_count": int(total_outliers),
        "rows_with_outliers_fraction": float(
            outliers_report.get("rows_with_outliers_fraction", 0.0)
        ),
        "constant_columns_count": int(len(constant_columns)),
        "high_cardinality_columns_count": int(len(high_cardinality_columns)),
        "target_leakage_status": leakage_report.get("status", "not_checked"),
    }


def _build_tabular_object_metrics(df: pd.DataFrame) -> list[dict]:
    duplicate_mask = df.duplicated(keep=False)
    missing_mask = df.isna().any(axis=1)
    outlier_mask = _build_outlier_mask(df)

    object_metrics = []

    for index, row in df.iterrows():
        reasons = []

        if bool(duplicate_mask.iloc[index]):
            reasons.append("duplicate_row")

        if bool(missing_mask.iloc[index]):
            reasons.append("missing_values")

        if bool(outlier_mask.iloc[index]):
            reasons.append("outlier")

        status = "warning" if reasons else "ok"

        object_id = row["id"] if "id" in df.columns else index

        object_metrics.append(
            {
                "id": str(object_id),
                "row_index": int(index),
                "status": status,
                "reasons": reasons,
                "recommendation": _recommendation_from_reasons(reasons),
            }
        )

    return object_metrics


def _build_outlier_mask(df: pd.DataFrame) -> pd.Series:
    numeric_df = df.select_dtypes(include="number")

    if numeric_df.empty:
        return pd.Series(False, index=df.index)

    outlier_mask = pd.Series(False, index=df.index)

    for column in numeric_df.columns:
        if column.lower() in {"id", "target", "label"}:
            continue

        series = numeric_df[column].dropna()

        if series.empty:
            continue

        q1 = series.quantile(0.25)
        q3 = series.quantile(0.75)
        iqr = q3 - q1

        if iqr == 0:
            continue

        lower_bound = q1 - 1.5 * iqr
        upper_bound = q3 + 1.5 * iqr

        column_outliers = (df[column] < lower_bound) | (df[column] > upper_bound)
        outlier_mask = outlier_mask | column_outliers.fillna(False)

    return outlier_mask


def _recommendation_from_reasons(reasons: list[str]) -> str:
    if "duplicate_row" in reasons:
        return "review_or_remove_duplicate_row"

    if "outlier" in reasons:
        return "review_outlier"

    if "missing_values" in reasons:
        return "fix_missing_values"

    return "keep"


def _build_tabular_recommendations(result: dict) -> list[dict]:
    reports = result.get("reports", {})
    recommendations = []

    missing_report = reports.get("missing_values", {})
    duplicates_report = reports.get("duplicates", {})
    outliers_report = reports.get("outliers", {})
    constant_report = reports.get("constant_columns", {})
    cardinality_report = reports.get("high_cardinality", {})
    leakage_report = reports.get("target_leakage", {})

    if missing_report.get("rows_with_any_nan_fraction", 0.0) > 0:
        recommendations.append(
            {
                "type": "FIX_MISSING_VALUES",
                "priority": "medium",
                "title": "Fix missing values",
                "description": "Some rows contain missing values. Fill them or remove problematic rows.",
            }
        )

    if duplicates_report.get("duplicate_count", 0) > 0:
        recommendations.append(
            {
                "type": "REMOVE_DUPLICATE_ROWS",
                "priority": "medium",
                "title": "Review duplicate rows",
                "description": "Duplicate rows were found. Remove duplicates or keep one representative row.",
            }
        )

    outliers_count = outliers_report.get("outliers_count", {})
    if isinstance(outliers_count, dict) and sum(outliers_count.values()) > 0:
        recommendations.append(
            {
                "type": "REVIEW_OUTLIERS",
                "priority": "medium",
                "title": "Review numeric outliers",
                "description": "Some numeric columns contain statistical outliers based on IQR.",
            }
        )

    constant_columns = constant_report.get("constant_columns", [])
    if constant_columns:
        recommendations.append(
            {
                "type": "DROP_CONSTANT_COLUMNS",
                "priority": "low",
                "title": "Drop constant columns",
                "description": "Constant columns do not add useful signal for training.",
                "columns": constant_columns,
            }
        )

    high_cardinality_columns = cardinality_report.get("high_cardinality_columns", {})
    if high_cardinality_columns:
        recommendations.append(
            {
                "type": "REVIEW_HIGH_CARDINALITY_COLUMNS",
                "priority": "medium",
                "title": "Review high-cardinality columns",
                "description": "Some columns look like IDs or contain too many unique values.",
                "columns": list(high_cardinality_columns.keys()),
            }
        )

    if leakage_report.get("status") == "warning":
        recommendations.append(
            {
                "type": "CHECK_TARGET_LEAKAGE",
                "priority": "high",
                "title": "Check possible target leakage",
                "description": "Some features may leak target information and inflate model quality.",
            }
        )

    if not recommendations:
        recommendations.append(
            {
                "type": "TABULAR_DATASET_OK",
                "priority": "low",
                "title": "No critical tabular issues found",
                "description": "The tabular dataset looks stable based on current checks.",
            }
        )

    return recommendations