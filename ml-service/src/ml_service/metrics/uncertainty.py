import pandas as pd 

#задача 11, расчет неуверенности
def uncertainty_metrics(df: pd.DataFrame) -> pd.DataFrame:
    df['uncertainty_score']=1-df['confidence']
    return df