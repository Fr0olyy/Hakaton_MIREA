from typing import Dict, Any, List

def generate_tabular_recommendations(reports: Dict[str, Any]) -> List[str]:
    recs = []

    # 1. Пропуски
    missing_report = reports.get("missing_values", {})
    if missing_report.get("status") == "warning":
        critical_cols = missing_report.get("critical_columns", [])
        if critical_cols:
            recs.append(f"Удалите колонки {critical_cols} с критическим количеством пропусков.")

    # 2. Дубликаты
    dup_report = reports.get("duplicates", {})
    if dup_report.get("status") == "warning":
        dup_count = dup_report.get("duplicate_count", 0)
        recs.append(f"Удалите {dup_count} полных дубликатов строк для чистоты кросс-валидации.")

    # 3. Константные колонки
    const_report = reports.get("constant_columns", {})
    if const_report.get("status") == "warning":
        const_cols = const_report.get('constant_columns', [])
        if const_cols:
            recs.append(f"Удалите константные колонки {const_cols}, они не несут полезной информации.")

    # 4. Высокая кардинальность
    card_report = reports.get("high_cardinality", {})
    if card_report.get("status") == "warning":
        card_cols_dict = card_report.get('high_cardinality_columns', {})
        if card_cols_dict:
            card_cols = list(card_cols_dict.keys())
            recs.append(f"Колонки {', '.join(card_cols)} имеют слишком много уникальных значений. Исключите их или примените Target Encoding.")

    # 5. Утечка таргета
    leak_report = reports.get("target_leakage")
    if leak_report and leak_report.get("status") == "warning":
        leak_features = leak_report.get("leakage_features", {})
        if leak_features:
            leak_cols = list(leak_features.keys())
            recs.append(f"КРИТИЧНО! Возможна утечка таргета в колонках: {', '.join(leak_cols)}. Немедленно удалите их из признаков.")

    if not recs:
        recs.append("Датасет в отличном состоянии, аномалий не найдено.")

    return recs