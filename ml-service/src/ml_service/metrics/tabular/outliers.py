import pandas as pd

def analyze_outliers(df: pd.DataFrame, iqr_multiplier: float = 1.5) -> dict:

    numeric_cols = df.select_dtypes(include=['number']).columns
    
    if len(numeric_cols) == 0:
        return {"outliers_count": {}, "rows_with_outliers_fraction": 0.0, "status": "ok"}
        
    Q1 = df[numeric_cols].quantile(0.25)
    Q3 = df[numeric_cols].quantile(0.75)
    IQR = Q3 - Q1
    
    lower_bound = Q1 - iqr_multiplier * IQR
    upper_bound = Q3 + iqr_multiplier * IQR
    
    outliers_mask = (df[numeric_cols] < lower_bound) | (df[numeric_cols] > upper_bound)
    
    outliers_per_column = outliers_mask.sum()
    outliers_stats = outliers_per_column[outliers_per_column > 0].to_dict()
    
    rows_with_any_outlier = outliers_mask.any(axis=1).mean()
    
    report = {
        "outliers_count": outliers_stats,
        "rows_with_outliers_fraction": float(rows_with_any_outlier),
        "status": "warning" if rows_with_any_outlier > 0.10 else "ok" 
    }
    
    return report