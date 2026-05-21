import pandas as pd


def compute_label_error_probability(
    df: pd.DataFrame,
    prob_cols: list[str],
) -> pd.DataFrame:
    df = df.copy()

    label_error_scores = []

    for _, row in df.iterrows():
        label = str(row["label"]).strip()
        predicted_label = str(row.get("predicted_label", label)).strip()
        confidence = float(row.get("confidence", 0.0))

        # Если текущая метка совпадает с предсказанием,
        # считаем, что явной ошибки разметки нет.
        if predicted_label == label:
            label_error_scores.append(0.0)
            continue

        # Если модель уверенно предсказывает другой класс,
        # это сильный сигнал возможной ошибки label.
        if confidence >= 0.8:
            label_error_scores.append(confidence)
            continue

        # Если confidence не очень высокая, смотрим на разницу вероятностей.
        label_prob_col = f"prob_{label}"
        predicted_prob_col = f"prob_{predicted_label}"

        p_label = float(row.get(label_prob_col, 0.0))
        p_predicted = float(row.get(predicted_prob_col, confidence))

        score = max(0.0, p_predicted - p_label)

        label_error_scores.append(score)

    df["label_error_probability"] = label_error_scores

    return df