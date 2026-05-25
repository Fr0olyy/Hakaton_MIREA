import pandas as pd

#duplicate rows 

def analyze_duplicates(df: pd.DataFrame) -> dict:
    total_rows = len(df)
    
    #чтобы не взять пустой датасет
    if total_rows == 0:
        return {"duplicate_count": 0, "duplicate_fraction": 0.0, "status": "ok"}

    duplicate_count = int(df.duplicated(keep='first').sum())
    duplicate_fraction = duplicate_count / total_rows

    report = {
        "duplicate_count": duplicate_count,
        "duplicate_fraction": float(duplicate_fraction),
        "status": "warning" if duplicate_count > 0 else "ok"
    }
    
    return report