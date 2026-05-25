import pandas as pd

#Реализовать target leakage hints

def analyze_target_leakage(df: pd.DataFrame, target_col: str, threshold: float = 0.85) -> dict:
    if target_col not in df.columns:
        return {"leakage_features": {}, "status": "ok"}
        
    numeric_cols = df.select_dtypes(include=['number']).columns.tolist()
    
    if target_col not in numeric_cols:
        return {"leakage_features": {}, "status": "ok", "note": "Target is not numeric"}
        
    correlations = df[numeric_cols].corr()[target_col]
    
    correlations = correlations.abs().drop(labels=[target_col], errors='ignore')
    
    leakage_candidates = correlations[correlations > threshold].to_dict()
    
    report = {
        "leakage_features": leakage_candidates,
        "status": "warning" if leakage_candidates else "ok"
    }
    
    return report