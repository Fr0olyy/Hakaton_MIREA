import pandas as pd
import numpy as np

def generate_tabular_review_queue(df: pd.DataFrame, iqr_multiplier: float = 1.5) -> pd.DataFrame:
    if df.empty:
        return pd.DataFrame(columns=['row_index', 'reason'])

    is_duplicate = df.duplicated(keep='first')

    numeric_cols = df.select_dtypes(include=['number']).columns
    if len(numeric_cols) > 0:
        Q1 = df[numeric_cols].quantile(0.25)
        Q3 = df[numeric_cols].quantile(0.75)
        IQR = Q3 - Q1
        lower_bound = Q1 - iqr_multiplier * IQR
        upper_bound = Q3 + iqr_multiplier * IQR
        
        outliers_matrix = (df[numeric_cols] < lower_bound) | (df[numeric_cols] > upper_bound)
        has_outliers = outliers_matrix.any(axis=1)
    else:
        has_outliers = pd.Series(False, index=df.index)

    to_review_mask = is_duplicate|has_outliers

    review_df = df[to_review_mask].copy()

    if review_df.empty:
        return pd.DataFrame(columns=['row_index', 'reason'])

    review_df['row_index'] = review_df.index

    reasons = []
    for idx, row in review_df.iterrows():
        row_reasons = []
        if is_duplicate.loc[idx]:
            row_reasons.append("Полный дубликат строки")
        if has_outliers.loc[idx]:
            row_reasons.append("Обнаружены математические выбросы (outliers)")
        
        reasons.append("; ".join(row_reasons))

    review_df['reason'] = reasons

    cols_to_return = ['row_index', 'reason'] + [col for col in df.columns if col not in ['row_index', 'reason']]
    return review_df[cols_to_return]