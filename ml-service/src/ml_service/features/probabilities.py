import pandas as pd


def prepare_probabilities_and_confidence(
    df: pd.DataFrame,
) -> tuple[pd.DataFrame, list[str]]:
    df = df.copy()

    df["label"] = _clean_label_column(df["label"])

    if "predicted_label" in df.columns:
        df["predicted_label"] = _clean_label_column(df["predicted_label"])
    else:
        df["predicted_label"] = df["label"]

    classes = sorted(
        set(df["label"].dropna().astype(str))
        | set(df["predicted_label"].dropna().astype(str))
    )

    prob_cols = [f"prob_{class_name}" for class_name in classes]

    if all(col in df.columns for col in prob_cols):
        df["confidence"] = df[prob_cols].max(axis=1)
        return df, prob_cols

    for col in prob_cols:
        df[col] = 0.0

    for idx, row in df.iterrows():
        pred = str(row.get("predicted_label", row["label"])).strip().lower()
        conf = float(row.get("confidence", 0.9))

        conf = max(0.0, min(1.0, conf))

        residual = 1.0 - conf
        other_classes_count = len(classes) - 1

        for class_name in classes:
            col_name = f"prob_{class_name}"

            if class_name == pred:
                df.at[idx, col_name] = conf
            else:
                df.at[idx, col_name] = (
                    residual / other_classes_count
                    if other_classes_count > 0
                    else 0.0
                )

    df["confidence"] = df[prob_cols].max(axis=1)

    return df, prob_cols


def _clean_label_column(series: pd.Series) -> pd.Series:
    return (
        series
        .astype("string")
        .str.replace("\u00a0", " ", regex=False)
        .str.replace(r"\s+", " ", regex=True)
        .str.strip()
        .str.lower()
    )