import os 
from typing import Dict, Any

"""
Анализирует разметку YOLO для одной картинки.
Объединяет проверки границ и микро-рамок в единый отчет.
"""
    
from src.ml_service.metrics.cv.yolo_validator import validate_yolo_bounds 
from src.ml_service.metrics.cv.yolo_tiny_boxes import analyze_tiny_boxes 

def analyze_cv_dataset(txt_filepath: str) -> Dict[str, Any]:
    if not os.path.exists(txt_filepath):
         return {"cv_quality_score": 0.0, "status": "error", "reports": {}}

    reports = {
        "out_of_bounds": validate_yolo_bounds(txt_filepath),
        "tiny_boxes": analyze_tiny_boxes(txt_filepath)
    }

    score = 100.0
    
    for report_name, report_data in reports.items():
        if report_data.get("status") == "warning":
            score -= 25

    score = max(0.0, score)

    result = {
        "cv_quality_score": score,
        "status": "ready" if score >= 75.0 else "needs_review",
        "reports": reports
    }

    return result