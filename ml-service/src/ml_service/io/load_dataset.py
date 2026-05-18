import pandas as pd
from pathlib import Path

def load_dataset(dataset_path: str | Path) -> pd.DataFrame:
    # Читаем CSV файл
    df = pd.read_csv(dataset_path)
    
    # 1. Чистим пробелы в НАЗВАНИЯХ колонок (" label   " -> "label")
    df.columns = df.columns.str.strip()
    
    # 2. Проверяем наличие обязательных колонок (Задача №6)
    required_cols = {'id', 'file_path', 'label'}
    missing = required_cols - set(df.columns)
    if missing:
        raise ValueError(f"В датасете не хватает обязательных колонок: {missing}")
        
    # 3. ИСПРАВЛЕНИЕ: Чистим пробелы в САМИХ ЗНАЧЕНИЯХ (внутри ячеек)
    # Пробегаемся по всем колонкам, и если там текст (object), убираем скрытые пробелы
    for col in df.columns:
        if df[col].dtype == 'object': 
            df[col] = df[col].astype(str).str.strip()
            
    # Приводим id к строке для стабильности бэкенда
    df['id'] = df['id'].astype(str)
    
    # Если колонки split нет, создаем её по умолчанию
    if 'split' not in df.columns:
        df['split'] = 'train'
        
    return df