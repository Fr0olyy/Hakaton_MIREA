import pandas as pd

def prepare_probabilities_and_confidence(df: pd.DataFrame) -> tuple[pd.DataFrame, list[str]]:
    classes = sorted(list(set(df['label'].unique()) | set(df.get('predicted_label', []))))
    prob_cols = [f"prob_{c}" for c in classes]
    
    if all(col in df.columns for col in prob_cols):
        df['confidence'] = df[prob_cols].max(axis=1)
        return df, prob_cols

    for col in prob_cols:
        df[col] = 0.0
        
    for idx, row in df.iterrows():
        pred = row.get('predicted_label', row['label'])
        conf = float(row.get('confidence', 0.9))
        
        residual = 1.0 - conf
        other_classes_count = len(classes) - 1
        
        for c in classes:
            col_name = f"prob_{c}"
            if c == pred:
                df.at[idx, col_name] = conf
            else:
                df.at[idx, col_name] = residual / other_classes_count if other_classes_count > 0 else 0.0
                
    df['confidence'] = df[prob_cols].max(axis=1)
    return df, prob_cols