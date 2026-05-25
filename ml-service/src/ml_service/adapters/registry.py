from ml_service.adapters.base import AdapterInput, AdapterOutput, make_error_output


def run_adapter(adapter_input: AdapterInput) -> AdapterOutput:
    modality = adapter_input.modality

    if modality == "tabular_classification":
        from ml_service.adapters.tabular_adapter import run_tabular_adapter

        return run_tabular_adapter(adapter_input)

    if modality == "image_detection_yolo":
        from ml_service.adapters.cv_adapter import run_cv_adapter

        return run_cv_adapter(adapter_input)
    
    if modality == "image_classification":
        from ml_service.adapters.image_classification_adapter import (
            run_image_classification_adapter,
        )

        return run_image_classification_adapter(adapter_input)

    return make_error_output(
        modality=adapter_input.modality,
        task_type=adapter_input.task_type,
        error=f"Unsupported modality: {modality}",
    )