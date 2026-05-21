from pathlib import Path
import json

import pandas as pd


def build_recommendations(df: pd.DataFrame) -> list[dict]:
    df = df.copy()

    defaults = {
        "label_error_probability": 0.0,
        "duplicate_score": 0.0,
        "quality_score": 1.0,
        "entropy": 0.0,
        "uncertainty_score": 0.0,
        "class_deficit_score": 0.0,
    }

    for column, default_value in defaults.items():
        if column not in df.columns:
            df[column] = default_value

        df[column] = pd.to_numeric(df[column], errors="coerce").fillna(default_value)

    recommendations = []

    suspected_label_errors = df[df["label_error_probability"] >= 0.5]

    if len(suspected_label_errors) > 0:
        recommendations.append({
            "type": "RECHECK_LABELS",
            "priority": "high",
            "title": f"Проверить {len(suspected_label_errors)} объектов с вероятной ошибкой разметки",
            "description": "У этих объектов текущая метка не совпадает с уверенным предсказанием модели или вероятность ошибки разметки высокая.",
            "affected_objects_count": int(len(suspected_label_errors)),
            "object_ids": suspected_label_errors["id"].astype(str).head(50).tolist(),
        })

    duplicates = df[df["duplicate_score"] >= 0.7]

    if len(duplicates) > 0:
        recommendations.append({
            "type": "REMOVE_DUPLICATES",
            "priority": "medium",
            "title": f"Удалить или проверить {len(duplicates)} дубликатов",
            "description": "В датасете найдены одинаковые или почти одинаковые изображения. Лучше оставить только один представитель каждой группы.",
            "affected_objects_count": int(len(duplicates)),
            "object_ids": duplicates["id"].astype(str).head(50).tolist(),
        })

    bad_quality = df[df["quality_score"] < 0.65]

    if len(bad_quality) > 0:
        recommendations.append({
            "type": "EXCLUDE_BAD_QUALITY",
            "priority": "medium",
            "title": f"Проверить {len(bad_quality)} изображений плохого качества",
            "description": "Есть изображения с размытием, низким разрешением, плохой яркостью или низким контрастом.",
            "affected_objects_count": int(len(bad_quality)),
            "object_ids": bad_quality["id"].astype(str).head(50).tolist(),
        })

    hard_examples = df[
        (df["entropy"] >= 0.65)
        | (df["uncertainty_score"] >= 0.45)
    ]

    if len(hard_examples) > 0:
        recommendations.append({
            "type": "REVIEW_HARD_EXAMPLES",
            "priority": "medium",
            "title": f"Разобрать {len(hard_examples)} сложных объектов",
            "description": "Модель не уверена в этих объектах. Их стоит проверить вручную и использовать как hard examples.",
            "affected_objects_count": int(len(hard_examples)),
            "object_ids": hard_examples["id"].astype(str).head(50).tolist(),
        })

    rare_classes = _find_rare_classes(df)

    for class_name, objects_count in rare_classes[:3]:
        recommendations.append({
            "type": "COLLECT_MORE_DATA",
            "priority": "high",
            "title": f"Дособрать данные для класса {class_name}",
            "description": f"Класс {class_name} представлен слабее остальных. Сейчас объектов этого класса: {objects_count}.",
            "affected_objects_count": 0,
            "object_ids": [],
            "target_class": class_name,
        })

    if not recommendations:
        recommendations.append({
            "type": "DATASET_OK",
            "priority": "low",
            "title": "Критических проблем не найдено",
            "description": "Датасет выглядит достаточно стабильным. Следующий шаг — запустить benchmark v1 vs v2.",
            "affected_objects_count": 0,
            "object_ids": [],
        })

    return recommendations


def export_recommendations_json(
    recommendations: list[dict],
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    recommendations_path = output_dir / "recommendations.json"

    with recommendations_path.open("w", encoding="utf-8") as file:
        json.dump(recommendations, file, ensure_ascii=False, indent=2)

    return recommendations_path


def _find_rare_classes(df: pd.DataFrame) -> list[tuple[str, int]]:
    if "label" not in df.columns or len(df) == 0:
        return []

    class_counts = df["label"].value_counts().to_dict()

    if not class_counts:
        return []

    max_count = max(class_counts.values())
    rare_classes = []

    for class_name, count in class_counts.items():
        deficit_score = 1 - count / max_count

        if deficit_score >= 0.5:
            rare_classes.append((str(class_name), int(count)))

    rare_classes.sort(key=lambda item: item[1])

    return rare_classes