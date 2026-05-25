import sys
import os
import json
import pandas as pd
import numpy as np
from pathlib import Path

# Добавляем корень в пути
sys.path.append(str(Path(__file__).resolve().parent.parent))

from src.ml_service.adapters.tabular_adapter import analyze_tabular_dataset
from src.ml_service.adapters.cv_adapter import analyze_cv_dataset
from src.ml_service.policy_engine import evaluate_dataset_policy

def run_pipeline():
    print("Запуск DataForge ML Pipeline...\n")

    #Генерация фейковых данных
    df = pd.DataFrame({"age": [25, 150, 30], "salary": [100, np.nan, 120], "is_churn": [0, 1, 0]})
    test_cv_path = "temp_cv.txt"
    with open(test_cv_path, "w") as f:
        f.write("0 0.5 0.5 0.2 0.2\n1 1.2 0.5 0.1 0.1")

    tab_result = analyze_tabular_dataset(df, target_col="is_churn")
    tab_score = tab_result.get("tabular_quality_score", 0.0)
    print(f"Табличный скор: {tab_score}")

    cv_result = analyze_cv_dataset(test_cv_path)
    cv_score = cv_result.get("cv_quality_score", 0.0)
    print(f"CV скор: {cv_score}\n")

    final_decision = evaluate_dataset_policy(tab_score, cv_score)

    print("Вердикт системы:")
    print(json.dumps(final_decision, indent=2, ensure_ascii=False))

    # Убираем за собой
    if os.path.exists(test_cv_path):
        os.remove(test_cv_path)

if __name__ == "__main__":
    run_pipeline()