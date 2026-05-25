from ml_service.pipeline.multimodal_pipeline import run_multimodal_pipeline


def main():
    tabular_result = run_multimodal_pipeline(
        modality="tabular_classification",
        task_type="classification",
        dataset_path="data/demo/tabular.csv",
        data_dir=None,
        output_dir="outputs/demo_tabular",
    )

    print("\nTABULAR RESULT")
    print(tabular_result)

    yolo_result = run_multimodal_pipeline(
        modality="image_detection_yolo",
        task_type="detection",
        dataset_path=None,
        data_dir="data/demo/yolo",
        output_dir="outputs/demo_yolo",
    )

    print("\nYOLO RESULT")
    print(yolo_result)

    image_result = run_multimodal_pipeline(
    modality="image_classification",
    task_type="classification",
    dataset_path="data/demo/dataset.csv",
    data_dir="data/demo/images",
    output_dir="outputs/demo_image_multimodal",
)

    print("\nIMAGE RESULT")
    print(image_result)


if __name__ == "__main__":
    main()