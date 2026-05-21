from pathlib import Path
import json

import pandas as pd


def build_roadmap(df: pd.DataFrame) -> list[dict]:
    roadmap = []
    priority = 1

    defaults = {
        "label_error_probability": 0.0,
        "duplicate_score": 0.0,
        "quality_score": 1.0,
        "entropy": 0.0,
        "uncertainty_score": 0.0,
        "class_deficit_score": 0.0,
    }

    df = df.copy()

    for col, default_value in defaults.items():
        if col not in df.columns:
            df[col] = default_value

        df[col] = pd.to_numeric(df[col], errors="coerce").fillna(default_value)

    # 1. Ошибки разметки
    suspected_label_errors = df[df["label_error_probability"] >= 0.5]

    if len(suspected_label_errors) > 0:
        roadmap.append(
            {
                "priority": priority,
                "title": f"Проверить {len(suspected_label_errors)} объектов с вероятной ошибкой разметки",
                "description": "У этих объектов текущая метка не совпадает с уверенным предсказанием модели или вероятность ошибки разметки высокая.",
                "action_type": "RECHECK_LABELS",
                "expected_impact": "Снижение шума в обучающей выборке и улучшение качества обучения.",
            }
        )
        priority += 1

    # 2. Дубли
    duplicates = df[df["duplicate_score"] >= 0.7]

    if len(duplicates) > 0:
        roadmap.append(
            {
                "priority": priority,
                "title": f"Удалить или проверить {len(duplicates)} дубликатов",
                "description": "В датасете найдены одинаковые или почти одинаковые изображения.",
                "action_type": "REMOVE_DUPLICATES",
                "expected_impact": "Повышение разнообразия данных и снижение риска переобучения.",
            }
        )
        priority += 1

    # 3. Плохое качество
    bad_quality = df[df["quality_score"] < 0.65]

    if len(bad_quality) > 0:
        roadmap.append(
            {
                "priority": priority,
                "title": f"Проверить {len(bad_quality)} изображений плохого качества",
                "description": "Есть изображения с низким разрешением, размытием, плохой яркостью или низким контрастом.",
                "action_type": "CHECK_BAD_QUALITY",
                "expected_impact": "Снижение визуального шума в датасете.",
            }
        )
        priority += 1

    # 4. Редкие классы
    rare_classes = _find_rare_classes(df)

    for class_name, objects_count in rare_classes[:2]:
        roadmap.append(
            {
                "priority": priority,
                "title": f"Дособрать данные для класса {class_name}",
                "description": f"Класс {class_name} представлен слабее остальных. Сейчас объектов этого класса: {objects_count}.",
                "action_type": "COLLECT_MORE_DATA",
                "target_class": class_name,
                "expected_impact": "Улучшение качества модели на редких классах и снижение дисбаланса.",
            }
        )
        priority += 1

    # 5. Сложные объекты
    hard_examples = df[
        (df["entropy"] >= 0.65)
        | (df["uncertainty_score"] >= 0.45)
    ]

    if len(hard_examples) > 0:
        roadmap.append(
            {
                "priority": priority,
                "title": f"Разобрать {len(hard_examples)} сложных объектов",
                "description": "Модель не уверена в этих объектах. Их стоит проверить и использовать как hard examples.",
                "action_type": "REVIEW_HARD_EXAMPLES",
                "expected_impact": "Улучшение устойчивости модели на сложных и пограничных примерах.",
            }
        )
        priority += 1

    # Если проблем почти нет
    if not roadmap:
        roadmap.append(
            {
                "priority": 1,
                "title": "Критических проблем не найдено",
                "description": "Датасет выглядит достаточно стабильным. Следующий шаг — запустить benchmark v1 vs v2.",
                "action_type": "RUN_BENCHMARK",
                "expected_impact": "Проверка, даёт ли текущий датасет прирост качества модели.",
            }
        )

    return roadmap[:5]


def _find_rare_classes(df: pd.DataFrame) -> list[tuple[str, int]]:
    if "label" not in df.columns:
        return []

    class_counts = df["label"].value_counts().to_dict()

    if not class_counts:
        return []

    max_count = max(class_counts.values())

    rare_classes = []

    for class_name, count in class_counts.items():
        if max_count == 0:
            continue

        deficit_score = 1 - count / max_count

        if deficit_score >= 0.5:
            rare_classes.append((class_name, int(count)))

    rare_classes.sort(key=lambda item: item[1])

    return rare_classes


def export_roadmap_json(
    roadmap: list[dict],
    output_dir: str | Path,
) -> Path:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    roadmap_path = output_dir / "roadmap.json"

    with roadmap_path.open("w", encoding="utf-8") as file:
        json.dump(roadmap, file, ensure_ascii=False, indent=2)

    return roadmap_path