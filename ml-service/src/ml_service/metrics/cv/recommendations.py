from typing import Dict, Any, List

def generate_cv_recommendations(reports: Dict[str, Any]) -> List[str]: 
    recs = []

    bounds_report = reports.get("out_of_bounds", {})

    if bounds_report.get("status") == "warning":
        bad_boxes = bounds_report.get("out_of_bounds_count", 0)
        recs.append(f"Найдено {bad_boxes} рамок, выходящих за границы картинки (координаты < 0 или > 1). Удалите их или исправьте.")

    tiny_report = reports.get("tiny_boxes", {})

    if tiny_report.get("status") == "warning":
        micro_boxes = tiny_report.get("tiny_boxes_count", 0)
        recs.append(f"Обнаружено {micro_boxes} микро-рамок (слишком маленькая площадь). Удалите их, чтобы модель не училась на шуме.")

    if not recs:
        recs.append("YOLO разметка в идеальном состоянии!")

    return recs