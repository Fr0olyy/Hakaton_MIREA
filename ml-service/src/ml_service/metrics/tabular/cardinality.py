import pandas as pd

#слишком много уникальных значений

def analyze_high_cardinality(df: pd.DataFrame, threshold_fraction: float = 0.5) -> dict:
    total_rows = len(df)
    if total_rows == 0:
        return {"high_cardinality_columns": {}, "status": "ok"}

    cat_df = df.select_dtypes(include=['object', 'category'])
    
    if cat_df.empty:
        return {"high_cardinality_columns": {}, "status": "ok"}

    unique_counts = cat_df.nunique()
    
    unique_fractions = unique_counts/total_rows
    
    high_cardinality_cols = unique_counts[unique_fractions>threshold_fraction].to_dict()

    report = {
        "high_cardinality_columns": high_cardinality_cols,
        "status": "warning" if high_cardinality_cols else "ok"
    }
    
    return report