import os
from typing import Dict, Any

def analyze_tiny_boxes(txt_filepath: str, min_area_threshold: float = 0.0001) -> Dict[str, Any]:
    if not os.path.exists(txt_filepath):
        return {"status": "error", "error_msg": "File not found"}

    total_boxes = 0
    tiny_boxes_count = 0

    with open(txt_filepath, 'r') as file:
        lines = file.readlines()

    for line in lines:
        line = line.strip()
        
        if not line:
            continue
            
        total_boxes += 1
        
        parts = line.split()
        
        if len(parts) != 5:
            continue
            
        try:
            w = float(parts[3])
            h = float(parts[4])
            
            area = w * h
            
            # Проверяем на микро-размер
            if area < min_area_threshold:
                tiny_boxes_count += 1
                
        except ValueError:
            continue

    report = {
        "total_boxes_checked": total_boxes, 
        "tiny_boxes_count": tiny_boxes_count,
        "status": "warning" if tiny_boxes_count > 0 else "ok"
    }
    
    return report