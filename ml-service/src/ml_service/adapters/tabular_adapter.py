import pandas as pd
from typing import Dict, Any

from src.ml_service.metrics.tabular.missing import analyze_missing_values
from src.ml_service.metrics.tabular.duplicates import analyze_duplicates
from src.ml_service.metrics.tabular.outliers import analyze_outliers
from src.ml_service.metrics.tabular.leakage import analyze_target_leakage
from src.ml_service.metrics.tabular.constant import analyze_constant_columns
from src.ml_service.metrics.tabular.cardinality import analyze_high_cardinality

def analyze_tabular_dataset(
    df: pd.DataFrame, 
    target_col: str = None
) -> Dict[str, Any]:
    
    # Если датафрейм пустой, возвращаем заглушку
    if df.empty:
         return {"tabular_quality_score": 0.0, "status": "error", "reports": {}}

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
    
    for report_name, report_data in reports.items():
        if report_data.get("status") == "warning":
            score -= 15

    score = max(0.0, score)

    result = {
        "tabular_quality_score": score,
        "status": "ready" if score >= 70.0 else "needs_review",
        "reports": reports
    }

    return result