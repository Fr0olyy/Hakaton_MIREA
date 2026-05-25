from pathlib import Path
from typing import Any, Dict

from ml_service.metrics.cv.yolo_validator import validate_yolo_bounds
from ml_service.metrics.cv.yolo_tiny_boxes import analyze_tiny_boxes

from ml_service.adapters.base import (
    AdapterInput,
    AdapterOutput,
    make_success_output,
    make_error_output,
)


def analyze_cv_dataset(txt_filepath: str | Path) -> Dict[str, Any]:
    txt_filepath = Path(txt_filepath)

    if not txt_filepath.exists() or not txt_filepath.is_file():
        return {
            "cv_quality_score": 0.0,
            "status": "error",
            "reports": {},
        }

    reports = {
        "out_of_bounds": validate_yolo_bounds(str(txt_filepath)),
        "tiny_boxes": analyze_tiny_boxes(str(txt_filepath)),
    }

    score = 100.0

    for _, report_data in reports.items():
        if report_data.get("status") == "warning":
            score -= 25
        elif report_data.get("status") == "error":
            score -= 50

    score = max(0.0, score)

    return {
        "cv_quality_score": score,
        "status": "ready" if score >= 75.0 else "needs_review",
        "reports": reports,
    }


def run_cv_adapter(adapter_input: AdapterInput) -> AdapterOutput:
    try:
        if adapter_input.data_dir is None:
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error="data_dir is required for image_detection_yolo",
            )

        data_path = Path(adapter_input.data_dir)

        if not data_path.exists():
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error=f"data_dir not found: {data_path}",
            )

        txt_files = _collect_yolo_txt_files(data_path)

        if not txt_files:
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error=f"No YOLO .txt label files found in: {data_path}",
                metadata={"data_dir": str(data_path)},
            )

        file_results = []

        for txt_file in txt_files:
            result = analyze_cv_dataset(txt_file)
            file_results.append(
                {
                    "file_path": str(txt_file),
                    **result,
                }
            )

        dataset_metrics = _build_cv_dataset_metrics(file_results)
        object_metrics = _build_cv_object_metrics(file_results)
        review_queue = [item for item in object_metrics if item["status"] != "ok"]
        recommendations = _build_cv_recommendations(dataset_metrics)

        return make_success_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            dataset_metrics=dataset_metrics,
            object_metrics=object_metrics,
            review_queue=review_queue,
            recommendations=recommendations,
            roadmap=[],
            collection_tasks=[],
            synthetic_tasks=[],
            class_action_plan=[],
            exports={},
            metadata={
                "data_dir": str(data_path),
                "label_files_count": len(txt_files),
            },
            warnings=[],
        )

    except Exception as error:
        return make_error_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            error=str(error),
        )


def _collect_yolo_txt_files(data_path: Path) -> list[Path]:
    if data_path.is_file() and data_path.suffix == ".txt":
        return [data_path]

    if data_path.is_dir():
        return sorted(data_path.rglob("*.txt"))

    return []


def _build_cv_dataset_metrics(file_results: list[dict]) -> dict:
    total_files = len(file_results)

    if total_files == 0:
        return {
            "total_label_files": 0,
            "cv_quality_score": 0.0,
            "files_needing_review": 0,
            "out_of_bounds_files_count": 0,
            "tiny_boxes_files_count": 0,
        }

    scores = [float(item.get("cv_quality_score", 0.0)) for item in file_results]

    files_needing_review = 0
    out_of_bounds_files_count = 0
    tiny_boxes_files_count = 0

    for item in file_results:
        if item.get("status") != "ready":
            files_needing_review += 1

        reports = item.get("reports", {})

        out_of_bounds = reports.get("out_of_bounds", {})
        tiny_boxes = reports.get("tiny_boxes", {})

        if out_of_bounds.get("status") in {"warning", "error"}:
            out_of_bounds_files_count += 1

        if tiny_boxes.get("status") in {"warning", "error"}:
            tiny_boxes_files_count += 1

    return {
        "total_label_files": int(total_files),
        "cv_quality_score": float(sum(scores) / total_files),
        "files_needing_review": int(files_needing_review),
        "out_of_bounds_files_count": int(out_of_bounds_files_count),
        "tiny_boxes_files_count": int(tiny_boxes_files_count),
    }


def _build_cv_object_metrics(file_results: list[dict]) -> list[dict]:
    object_metrics = []

    for index, item in enumerate(file_results):
        reports = item.get("reports", {})
        reasons = []

        out_of_bounds = reports.get("out_of_bounds", {})
        tiny_boxes = reports.get("tiny_boxes", {})

        if out_of_bounds.get("status") in {"warning", "error"}:
            reasons.append("bbox_out_of_bounds")

        if tiny_boxes.get("status") in {"warning", "error"}:
            reasons.append("tiny_boxes")

        status = "warning" if reasons else "ok"

        object_metrics.append(
            {
                "id": str(index),
                "file_path": item.get("file_path"),
                "status": status,
                "reasons": reasons,
                "recommendation": _cv_recommendation_from_reasons(reasons),
                "cv_quality_score": float(item.get("cv_quality_score", 0.0)),
            }
        )

    return object_metrics


def _cv_recommendation_from_reasons(reasons: list[str]) -> str:
    if "bbox_out_of_bounds" in reasons:
        return "fix_invalid_bboxes"

    if "tiny_boxes" in reasons:
        return "review_tiny_boxes"

    return "keep"


def _build_cv_recommendations(dataset_metrics: dict) -> list[dict]:
    recommendations = []

    if dataset_metrics.get("out_of_bounds_files_count", 0) > 0:
        recommendations.append(
            {
                "type": "FIX_INVALID_BBOXES",
                "priority": "high",
                "title": "Fix invalid YOLO bounding boxes",
                "description": "Some YOLO annotations contain coordinates outside the normalized 0..1 range.",
            }
        )

    if dataset_metrics.get("tiny_boxes_files_count", 0) > 0:
        recommendations.append(
            {
                "type": "REVIEW_TINY_BOXES",
                "priority": "medium",
                "title": "Review tiny bounding boxes",
                "description": "Some objects have extremely small bounding boxes and may add noise to detection training.",
            }
        )

    if not recommendations:
        recommendations.append(
            {
                "type": "YOLO_DATASET_OK",
                "priority": "low",
                "title": "No critical YOLO annotation issues found",
                "description": "The YOLO annotation files look stable based on current checks.",
            }
        )

    return recommendations