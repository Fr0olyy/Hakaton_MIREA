import pandas as pd

#зачада 13, расчет class_distribution
def distribution_metrics(df: pd.DataFrame) -> pd.DataFrame:
    class_counts = df['label'].value_counts()
    total_objects = len(df['id'])
    class_fractions = class_counts/total_objects
    df['class_distribution'] = df['label'].map(class_fractions)
    
    #задача 14, расчет imbalance_index
    max_count = class_counts.max()
    imbalance = (max_count-class_counts)/max_count
    df['imbalance_index'] = df['label'].map(imbalance)
    
    #задача 15, расчет class_deficit_score
    df['class_deficit_score'] = 1-df['class_distribution']
    
    return df 

