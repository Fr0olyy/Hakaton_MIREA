import os
from typing import Dict, Any

def validate_yolo_bounds(txt_filepath: str) -> Dict[str, Any]:
    if not os.path.exists(txt_filepath):
        return {"status": "error", "error_msg": "File not found"}

    total_boxes = 0
    out_of_bounds_boxes = 0
    with open(txt_filepath, 'r') as file:
        lines = file.readlines()

    for line in lines:
        line = line.strip()
        if not line:
            continue #Пропускаем пустые строки
            
        total_boxes += 1
        
        #Разделяем строку по пробелам на список строк
        parts = line.split()
        
        # Если в строке не ровно 5 элементов — это уже ошибка формата
        if len(parts) != 5:
            out_of_bounds_boxes += 1
            continue
            
        try:
            x = float(parts[1])
            y = float(parts[2])
            w = float(parts[3])
            h = float(parts[4])
            
            if x < 0.0 or x > 1.0 or y < 0.0 or y > 1.0 or w < 0.0 or w > 1.0 or h < 0.0 or h > 1.0:
                out_of_bounds_boxes+=1
                
        except ValueError:  # Если вместо чисел в файле оказался текст (буквы)
            out_of_bounds_boxes += 1

    # Собираем итоговый отчет
    report = {
        "total_boxes_checked": total_boxes,
        "out_of_bounds_count": out_of_bounds_boxes,
        "status": "warning" if out_of_bounds_boxes > 0 else "ok"
    }

    return report

'''Вводная: Как работает формат YOLO
В задачах детекции (когда мы ищем объекты на картинке) нейросети формата YOLO требуют очень 
специфическую разметку. Для каждой картинки image.jpg создается текстовый файл image.txt.

Внутри этого текстового файла каждая строка — это один объект (одна рамка/bounding box). 
Строка состоит из 5 чисел, разделенных пробелом:
class_id x_center y_center width height

Главное правило YOLO: Все координаты (x, y, ширина, высота) нормализованы. Это значит, 
что они измеряются не в пикселях, а в долях от размера картинки, и строго должны лежать в
диапазоне от 0.0 до 1.0. Если координата равна 1.5 или -0.2 — разметка сломана, и модель при 
обучении выдаст ошибку.'''