from pathlib import Path
import json

from ml_service.pipeline.multimodal_pipeline import run_multimodal_pipeline


CASES = [
    {
        "name": "image",
        "modality": "image_classification",
        "task_type": "classification",
        "dataset_path": "data/demo/dataset.csv",
        "data_dir": "data/demo/images",
        "output_dir": "outputs/smoke_image",
    },
    {
        "name": "tabular",
        "modality": "tabular_classification",
        "task_type": "classification",
        "dataset_path": "data/demo/tabular.csv",
        "data_dir": None,
        "output_dir": "outputs/smoke_tabular",
    },
    {
        "name": "yolo",
        "modality": "image_detection_yolo",
        "task_type": "detection",
        "dataset_path": None,
        "data_dir": "data/demo/yolo",
        "output_dir": "outputs/smoke_yolo",
    },
]


REQUIRED_FILES = [
    "adapter_output.json",
    "dataset_metrics.json",
    "object_metrics.json",
    "review_queue.json",
    "recommendations.json",
    "roadmap.json",
    "class_action_plan.json",
    "collection_tasks.json",
    "synthetic_tasks.json",
    "dataset_v2_conservative.csv",
    "dataset_v2_balanced.csv",
    "dataset_v2_aggressive.csv",
    "dataset_strategy_summary.json",
    "agent_context.json",
]


def main():
    for case in CASES:
        print(f"\nRunning {case['name']}...")

        result = run_multimodal_pipeline(
            modality=case["modality"],
            task_type=case["task_type"],
            dataset_path=case["dataset_path"],
            data_dir=case["data_dir"],
            output_dir=case["output_dir"],
        )

        assert result["status"] == "completed", result
        assert result["objects_count"] > 0, result

        output_dir = Path(case["output_dir"])

        for filename in REQUIRED_FILES:
            path = output_dir / filename
            assert path.exists(), f"Missing file: {path}"

        with open(output_dir / "adapter_output.json", "r", encoding="utf-8") as file:
            adapter_output = json.load(file)

        assert adapter_output["metadata"]["data_action_policy_applied"] is True
        assert len(adapter_output["class_action_plan"]) > 0

        with open(output_dir / "agent_context.json", "r", encoding="utf-8") as file:
            agent_context = json.load(file)

        assert "project_summary" in agent_context
        assert "class_action_plan" in agent_context
        assert "collection_tasks" in agent_context
        assert "synthetic_tasks" in agent_context
        assert "dataset_strategies" in agent_context

        print(f"{case['name']} OK")
        print("objects:", result["objects_count"])
        print("review_queue:", result["review_queue_count"])
        print("files:", len(result["files"]))

    print("\nLEVEL 2 SMOKE TEST PASSED")


if __name__ == "__main__":
    main()