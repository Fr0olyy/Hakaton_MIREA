import sys
import os
import json
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from src.ml_service.adapters.cv_adapter import analyze_cv_dataset
from src.ml_service.metrics.cv.recommendations import generate_cv_recommendations

test_txt_path = "test_image.txt"

# Создаем грязную разметку
yolo_data = """0 0.5 0.5 0.2 0.2
1 1.2 0.5 0.1 0.1
2 0.5 -0.1 0.3 0.3
0 0.2 0.2 0.5
3 0.8 0.8 oops 0.2
4 0.1 0.1 0.1 0.1
5 0.5 0.5 0.00001 0.00001"""

with open(test_txt_path, "w") as f:
    f.write(yolo_data)

print("Запуск Главного CV Адаптера\n")
result = analyze_cv_dataset(test_txt_path)
print(json.dumps(result, indent=2, ensure_ascii=False))

print("\n" + "="*60)
print("Рекомендации по очистке CV разметки:")
print("="*60)

recommendations = generate_cv_recommendations(result["reports"])

for i, rec in enumerate(recommendations, 1):
    print(f"{i}. {rec}")
print("\n")

if os.path.exists(test_txt_path):
    os.remove(test_txt_path)