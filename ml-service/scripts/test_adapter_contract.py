from ml_service.adapters.base import make_success_output


def main():
    output = make_success_output(
        modality="image_classification",
        task_type="classification",
        dataset_metrics={
            "total_objects": 100,
            "dataset_readiness_score": 78.5,
        },
        object_metrics=[
            {
                "id": "1",
                "label": "cat",
                "status": "ok",
            }
        ],
        recommendations=[
            {
                "type": "REMOVE_DUPLICATES",
                "priority": "medium",
                "title": "Check duplicates",
            }
        ],
    )

    print(output.to_dict())


if __name__ == "__main__":
    main()