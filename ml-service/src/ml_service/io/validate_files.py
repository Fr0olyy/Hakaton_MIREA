import pandas as pd
from pathlib import Path
from PIL import Image

def validate_images(df: pd.DataFrame, images_dir: Path) -> pd.DataFrame:
    file_exists_list = []
    image_opened_list = []
    error_list = []
    
    for path_str in df['file_path']:
        full_path = images_dir / path_str
        
        # Задача №7: Существует ли файл
        if not full_path.exists():
            file_exists_list.append(False)
            image_opened_list.append(False)
            error_list.append("File not found")
            continue
            
        file_exists_list.append(True)
        
        # Задача №8: Открывается ли изображение
        try:
            with Image.open(full_path) as img:
                img.verify() # Проверяем, что файл не битый
            image_opened_list.append(True)
            error_list.append("")
        except Exception as e:
            image_opened_list.append(False)
            error_list.append(str(e))
            
    df['file_exists'] = file_exists_list
    df['image_opened'] = image_opened_list
    df['image_error'] = error_list
    return df