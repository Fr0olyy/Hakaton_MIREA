import pandas as pd
import numpy as np
import json
import sys
from pathlib import Path

sys.path.append(str(Path(__file__).resolve().parent.parent))

from src.ml_service.adapters.tabular_adapter import analyze_tabular_dataset
from src.ml_service.metrics.tabular.review_queue import generate_tabular_review_queue
from src.ml_service.metrics.tabular.recommendations import generate_tabular_recommendations

#Генерируем "грязные" данные
df = pd.DataFrame({
    "user_id": [f"U{i}" for i in range(100)],
    "country": ["Russia"] * 100,
    "age": [25, 30, 35, 150, -5] * 20,         
    "salary": [100, 120, np.nan, 90, 110] * 20,
    "is_churn": [0, 1, 0, 1, 0] * 20,
})

df["churn_hint_score"] = df["is_churn"] * 100 
df = pd.concat([df, df.head(2)], ignore_index=True) 

print("табличный адаптер.\n")
result = analyze_tabular_dataset(df, target_col="is_churn")
print(json.dumps(result, indent=2, ensure_ascii=False))

print("\n" + "="*60)
print("Очередь на ручную проверку (Review Queue):")
print("="*60)

review_df = generate_tabular_review_queue(df)
recommendations = generate_tabular_recommendations(result["reports"])

print("\n" + "="*60)
print("💡 Рекомендации по очистке датасета:")
print("="*60)

for i, rec in enumerate(recommendations, 1):
    print(f"{i}. {rec}")