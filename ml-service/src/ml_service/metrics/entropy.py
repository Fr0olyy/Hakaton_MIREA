import pandas as pd
import numpy as np

#задача 12, расчет энтропии 
def entropy_metrics(df: pd.DataFrame, prob_cols: list[str]) -> pd.DataFrame:
    probs = df[prob_cols].values 
    df['entropy'] = -np.sum(probs * np.log(probs + 1e-9), axis=1)
    return df