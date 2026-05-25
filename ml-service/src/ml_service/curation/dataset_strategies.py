from __future__ import annotations

from pathlib import Path
from typing import Any

import pandas as pd


def export_dataset_strategies(
    adapter_output: dict[str, Any],
    output_dir: str | Path,
) -> dict[str, str]:
    modality = str(adapter_output.get("modality", "unknown"))
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    object_metrics = adapter_output.get("object_metrics", [])

    if not object_metrics:
        return {}

    df = pd.DataFrame(object_metrics)

    if modality == "image_classification":
        return _export_image_strategies(df, output_dir)

    if modality == "tabular_classification":
        return _export_tabular_strategies(df, output_dir)

    if modality == "image_detection_yolo":
        return _export_yolo_strategies(df, output_dir)

    return _export_default_strategies(df, output_dir)


def _export_image_strategies(df: pd.DataFrame, output_dir: Path) -> dict[str, str]:
    status = _series_str(df, "status")
    action = _series_str(df, "recommended_action")

    conservative_mask = ~(
        status.isin(["duplicate", "bad_quality"])
        | action.isin(["exclude_duplicate", "exclude_bad_quality"])
    )

    balanced_mask = ~(
        status.isin(["duplicate", "bad_quality", "suspected_label_error"])
        | action.isin(
            [
                "exclude_duplicate",
                "exclude_bad_quality",
                "relabel_or_review",
                "send_to_expert",
            ]
        )
    )

    aggressive_mask = ~(
        status.isin(["duplicate", "bad_quality", "suspected_label_error"])
        | action.isin(
            [
                "exclude_duplicate",
                "exclude_bad_quality",
                "relabel_or_review",
                "send_to_expert",
                "review_duplicate_rare_class",
            ]
        )
    )

    return _write_strategy_files(
        df=df,
        output_dir=output_dir,
        conservative_mask=conservative_mask,
        balanced_mask=balanced_mask,
        aggressive_mask=aggressive_mask,
    )


def _export_tabular_strategies(df: pd.DataFrame, output_dir: Path) -> dict[str, str]:
    status = _series_str(df, "status")
    action = _series_str(df, "recommended_action")

    conservative_mask = pd.Series(True, index=df.index)

    balanced_mask = ~action.isin(
        [
            "review_or_remove_duplicate_row",
        ]
    )

    aggressive_mask = ~(
        status.isin(["warning", "error", "needs_review"])
        | action.isin(
            [
                "review_or_remove_duplicate_row",
                "review_outlier",
                "fix_missing_values",
                "review_row",
            ]
        )
    )

    return _write_strategy_files(
        df=df,
        output_dir=output_dir,
        conservative_mask=conservative_mask,
        balanced_mask=balanced_mask,
        aggressive_mask=aggressive_mask,
    )


def _export_yolo_strategies(df: pd.DataFrame, output_dir: Path) -> dict[str, str]:
    status = _series_str(df, "status")
    action = _series_str(df, "recommended_action")

    conservative_mask = pd.Series(True, index=df.index)

    balanced_mask = ~action.isin(
        [
            "fix_invalid_bboxes",
        ]
    )

    aggressive_mask = ~(
        status.isin(["warning", "error", "needs_review"])
        | action.isin(
            [
                "fix_invalid_bboxes",
                "review_tiny_boxes",
                "review_annotations",
            ]
        )
    )

    return _write_strategy_files(
        df=df,
        output_dir=output_dir,
        conservative_mask=conservative_mask,
        balanced_mask=balanced_mask,
        aggressive_mask=aggressive_mask,
    )


def _export_default_strategies(df: pd.DataFrame, output_dir: Path) -> dict[str, str]:
    status = _series_str(df, "status")

    conservative_mask = pd.Series(True, index=df.index)
    balanced_mask = ~status.isin(["error"])
    aggressive_mask = ~status.isin(["warning", "error", "needs_review"])

    return _write_strategy_files(
        df=df,
        output_dir=output_dir,
        conservative_mask=conservative_mask,
        balanced_mask=balanced_mask,
        aggressive_mask=aggressive_mask,
    )


def _write_strategy_files(
    df: pd.DataFrame,
    output_dir: Path,
    conservative_mask: pd.Series,
    balanced_mask: pd.Series,
    aggressive_mask: pd.Series,
) -> dict[str, str]:
    conservative_path = output_dir / "dataset_v2_conservative.csv"
    balanced_path = output_dir / "dataset_v2_balanced.csv"
    aggressive_path = output_dir / "dataset_v2_aggressive.csv"
    summary_path = output_dir / "dataset_strategy_summary.json"

    conservative_df = df[conservative_mask].copy()
    balanced_df = df[balanced_mask].copy()
    aggressive_df = df[aggressive_mask].copy()

    conservative_df.to_csv(conservative_path, index=False)
    balanced_df.to_csv(balanced_path, index=False)
    aggressive_df.to_csv(aggressive_path, index=False)

    summary = {
        "raw_objects_count": int(len(df)),
        "conservative_objects_count": int(len(conservative_df)),
        "balanced_objects_count": int(len(balanced_df)),
        "aggressive_objects_count": int(len(aggressive_df)),
        "strategies": {
            "conservative": "Remove only obvious harmful objects.",
            "balanced": "Remove obvious harmful objects and risky review candidates.",
            "aggressive": "Keep only objects that look safe for training.",
        },
    }

    summary_df = pd.DataFrame([summary])
    summary_df.to_json(summary_path, orient="records", force_ascii=False, indent=2)

    return {
        "dataset_v2_conservative_csv": str(conservative_path),
        "dataset_v2_balanced_csv": str(balanced_path),
        "dataset_v2_aggressive_csv": str(aggressive_path),
        "dataset_strategy_summary_json": str(summary_path),
    }


def _series_str(df: pd.DataFrame, column: str) -> pd.Series:
    if column not in df.columns:
        return pd.Series("", index=df.index)

    return df[column].fillna("").astype(str).str.strip()