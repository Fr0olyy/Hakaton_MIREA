import pandas as pd
#Находит константные колонки
#Реализовать constant columns.  

def analyze_constant_columns(df: pd.DataFrame) -> dict:

    unique_counts = df.nunique()
    
    constant_cols_series = unique_counts[unique_counts == 1]

    constant_columns = constant_cols_series.index.tolist() 
    
    report = {
        "constant_columns": constant_columns,
        "status": "warning" if constant_columns else "ok"
    }
    
    return report