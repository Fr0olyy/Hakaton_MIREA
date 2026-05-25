from pathlib import Path
import json

from ml_service.adapters.base import AdapterInput, AdapterOutput
from ml_service.adapters.registry import run_adapter
from ml_service.policy.data_action_policy import apply_data_action_policy
from ml_service.curation.class_action_plan import build_class_action_plan
from ml_service.curation.collection_tasks import build_collection_tasks
from ml_service.curation.collection_tasks import build_collection_tasks
from ml_service.curation.synthetic_tasks import build_synthetic_tasks
from ml_service.curation.dataset_strategies import export_dataset_strategies
from ml_service.curation.agent_context import save_level2_agent_context

def run_multimodal_pipeline(
    modality: str,
    task_type: str,
    output_dir: str | Path,
    dataset_path: str | Path | None = None,
    data_dir: str | Path | None = None,
    project_id: str | None = None,
    extra: dict | None = None,
) -> dict:
    output_dir = Path(output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    adapter_input = AdapterInput(
        modality=modality,
        task_type=task_type,
        dataset_path=dataset_path,
        data_dir=data_dir,
        output_dir=output_dir,
        project_id=project_id,
        extra=extra or {},
    )

    adapter_output = run_adapter(adapter_input)
    output_dict = adapter_output.to_dict()

    output_dict = apply_data_action_policy(output_dict)
    output_dict["class_action_plan"] = build_class_action_plan(output_dict)
    output_dict["collection_tasks"] = build_collection_tasks(output_dict)
    output_dict["synthetic_tasks"] = build_synthetic_tasks(output_dict)

    strategy_exports = export_dataset_strategies(output_dict, output_dir)

    output_dict.setdefault("exports", {})
    output_dict["exports"].update(strategy_exports)

    agent_context_path = save_level2_agent_context(output_dict, output_dir)
    output_dict["exports"]["agent_context_json"] = agent_context_path

    outputs = _save_adapter_outputs(output_dict, output_dir)
    outputs.update(strategy_exports)
    outputs["agent_context_json"] = agent_context_path

    result = {
        "status": output_dict["status"],
        "modality": output_dict["modality"],
        "task_type": output_dict["task_type"],
        "dataset_metrics": output_dict["dataset_metrics"],
        "objects_count": len(output_dict["object_metrics"]),
        "review_queue_count": len(output_dict["review_queue"]),
        "recommendations_count": len(output_dict["recommendations"]),
        "output_dir": str(output_dir),
        "files": outputs,
        "errors": output_dict["errors"],
        "warnings": output_dict["warnings"],
    }

    return result


def _save_adapter_outputs(output_dict: dict, output_dir: Path) -> dict:
    files = {}

    full_output_path = output_dir / "adapter_output.json"
    _write_json(full_output_path, output_dict)
    files["adapter_output_json"] = str(full_output_path)

    dataset_metrics_path = output_dir / "dataset_metrics.json"
    _write_json(dataset_metrics_path, output_dict.get("dataset_metrics", {}))
    files["dataset_metrics_json"] = str(dataset_metrics_path)

    object_metrics_path = output_dir / "object_metrics.json"
    _write_json(object_metrics_path, output_dict.get("object_metrics", []))
    files["object_metrics_json"] = str(object_metrics_path)

    review_queue_path = output_dir / "review_queue.json"
    _write_json(review_queue_path, output_dict.get("review_queue", []))
    files["review_queue_json"] = str(review_queue_path)

    recommendations_path = output_dir / "recommendations.json"
    _write_json(recommendations_path, output_dict.get("recommendations", []))
    files["recommendations_json"] = str(recommendations_path)

    roadmap_path = output_dir / "roadmap.json"
    _write_json(roadmap_path, output_dict.get("roadmap", []))
    files["roadmap_json"] = str(roadmap_path)

    collection_tasks_path = output_dir / "collection_tasks.json"
    _write_json(collection_tasks_path, output_dict.get("collection_tasks", []))
    files["collection_tasks_json"] = str(collection_tasks_path)

    synthetic_tasks_path = output_dir / "synthetic_tasks.json"
    _write_json(synthetic_tasks_path, output_dict.get("synthetic_tasks", []))
    files["synthetic_tasks_json"] = str(synthetic_tasks_path)

    class_action_plan_path = output_dir / "class_action_plan.json"
    _write_json(class_action_plan_path, output_dict.get("class_action_plan", []))
    files["class_action_plan_json"] = str(class_action_plan_path)

    return files


def _write_json(path: Path, data) -> None:
    with path.open("w", encoding="utf-8") as file:
        json.dump(data, file, ensure_ascii=False, indent=2)