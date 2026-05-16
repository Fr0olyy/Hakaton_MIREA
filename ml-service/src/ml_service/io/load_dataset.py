import numpy as np
import pandas as pd 



def load_dataset(dataset_path):
    df = pd.read_csv(dataset_path)

    df.columns = df.columns.str.strip()
    
    print(df.columns.tolist())
    return df

