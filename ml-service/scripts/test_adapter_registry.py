from ml_service.adapters.base import AdapterInput
from ml_service.adapters.registry import run_adapter


def print_output(title: str, output):
    print(f"\n===== {title} =====")
    result = output.to_dict()
    print("status:", result["status"])
    print("modality:", result["modality"])
    print("task_type:", result["task_type"])
    print("errors:", result["errors"])
    print("dataset_metrics keys:", list(result["dataset_metrics"].keys()))
    print("object_metrics count:", len(result["object_metrics"]))
    print("review_queue count:", len(result["review_queue"]))
    print("recommendations count:", len(result["recommendations"]))


def test_unknown():
    adapter_input = AdapterInput(
        modality="unknown",
        task_type="classification",
        dataset_path=None,
        data_dir=None,
        output_dir="outputs/test_unknown",
    )

    output = run_adapter(adapter_input)
    print_output("UNKNOWN", output)


def test_tabular():
    adapter_input = AdapterInput(
        modality="tabular_classification",
        task_type="classification",
        dataset_path="data/demo/tabular.csv",
        data_dir=None,
        output_dir="outputs/demo_tabular",
    )

    output = run_adapter(adapter_input)
    print_output("TABULAR", output)


def test_yolo():
    adapter_input = AdapterInput(
        modality="image_detection_yolo",
        task_type="detection",
        dataset_path=None,
        data_dir="data/demo/yolo",
        output_dir="outputs/demo_yolo",
    )

    output = run_adapter(adapter_input)
    print_output("YOLO", output)


def main():
    test_unknown()
    test_tabular()
    test_yolo()


if __name__ == "__main__":
    main()