from pathlib import Path
import json
import pandas as pd

from ml_service.adapters.base import (
    AdapterInput,
    AdapterOutput,
    make_success_output,
    make_error_output,
)

from ml_service.pipeline.pipeline import run_pipeline


def run_image_classification_adapter(adapter_input: AdapterInput) -> AdapterOutput:
    try:
        if adapter_input.dataset_path is None:
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error="dataset_path is required for image_classification",
            )

        if adapter_input.data_dir is None:
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error="data_dir/images_dir is required for image_classification",
            )

        dataset_path = Path(adapter_input.dataset_path)
        images_dir = Path(adapter_input.data_dir)
        output_dir = Path(adapter_input.output_dir)

        if not dataset_path.exists():
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error=f"dataset_path not found: {dataset_path}",
            )

        if not images_dir.exists():
            return make_error_output(
                modality=adapter_input.modality,
                task_type=adapter_input.task_type,
                error=f"images_dir not found: {images_dir}",
            )
        
        result = run_pipeline(
            dataset_path=dataset_path,
            images_dir=images_dir,
            output_dir=output_dir,
        )

        files = result.get("files", {})
        data = result.get("data", {})

        results_csv = _resolve_output_path(
            files.get("results_csv"),
            output_dir / "results.csv",
        )
        review_queue_csv = _resolve_output_path(
            files.get("review_queue_csv"),
            output_dir / "review_queue.csv",
        )
        dataset_report_json = _resolve_output_path(
            files.get("dataset_report_json"),
            output_dir / "dataset_report.json",
        )
        recommendations_json = _resolve_output_path(
            files.get("recommendations_json"),
            output_dir / "recommendations.json",
        )
        roadmap_json = _resolve_output_path(
            files.get("roadmap_json"),
            output_dir / "roadmap.json",
        )
        dataset_v2_csv = _resolve_output_path(
            files.get("dataset_v2_csv"),
            output_dir / "dataset_v2.csv",
        )
        agent_context_json = _resolve_output_path(
            files.get("agent_context_json"),
            output_dir / "agent_context.json",
        )

        object_metrics = _read_csv_records(results_csv)
        review_queue = _read_csv_records(review_queue_csv)

        dataset_metrics = (
            data.get("dataset_report")
            or _read_json(dataset_report_json, default={})
        )

        recommendations = (
            data.get("recommendations")
            or _read_json(recommendations_json, default=[])
        )

        roadmap = (
            data.get("roadmap")
            or _read_json(roadmap_json, default=[])
        )

        exports = {
            "results_csv": str(results_csv),
            "review_queue_csv": str(review_queue_csv),
            "recommendations_json": str(recommendations_json),
            "roadmap_json": str(roadmap_json),
            "dataset_v2_csv": str(dataset_v2_csv),
            "dataset_report_json": str(dataset_report_json),
            "agent_context_json": str(agent_context_json),
        }

        metadata = {
            "dataset_path": str(dataset_path),
            "images_dir": str(images_dir),
            "legacy_pipeline": True,
            "pipeline_status": result.get("status", "unknown"),
            "objects_count": result.get("objects_count", len(object_metrics)),
            "review_queue_count": result.get("review_queue_count", len(review_queue)),
            "dataset_v2_objects_count": result.get("dataset_v2_objects_count", 0),
            "dataset_readiness_score": result.get("dataset_readiness_score"),
        }

        warnings = []

        if len(object_metrics) == 0:
            warnings.append("results.csv was not loaded or contains zero rows")

        if len(review_queue) == 0:
            warnings.append("review_queue.csv was not loaded or contains zero rows")

        return make_success_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            dataset_metrics=dataset_metrics,
            object_metrics=object_metrics,
            review_queue=review_queue,
            recommendations=recommendations,
            roadmap=roadmap,
            collection_tasks=[],
            synthetic_tasks=[],
            class_action_plan=[],
            exports=exports,
            metadata=metadata,
            warnings=warnings,
        )

    except Exception as error:
        return make_error_output(
            modality=adapter_input.modality,
            task_type=adapter_input.task_type,
            error=str(error),
        )


def _resolve_output_path(value, fallback: Path) -> Path:
    if value:
        return Path(value)

    return fallback


def _read_csv_records(path: Path) -> list[dict]:
    if not path.exists():
        return []

    df = pd.read_csv(path)
    return df.fillna("").to_dict(orient="records")


def _read_json(path: Path, default):
    if not path.exists():
        return default

    with path.open("r", encoding="utf-8") as file:
        return json.load(file)