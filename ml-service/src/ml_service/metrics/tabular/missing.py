import pandas as pd

#Реализовать missing values analysis 

def analyze_missing_values(df: pd.DataFrame, warning_threshold: float = 0.20) -> dict:
    missing_fractions = df.isna().mean()
    
    missing_stats = missing_fractions[missing_fractions > 0].to_dict()
    
    critical_columns = [
        col for col, frac in missing_stats.items() if frac > warning_threshold
    ]
    
    rows_with_any_nan = df.isna().any(axis=1).mean()
    
    report = {
        "missing_fractions": missing_stats,
        "critical_columns": critical_columns,
        "rows_with_any_nan_fraction": float(rows_with_any_nan),
        "status": "warning" if critical_columns else "ok"
    }
    
    return report