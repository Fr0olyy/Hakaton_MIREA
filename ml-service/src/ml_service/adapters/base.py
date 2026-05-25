from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


JsonDict = dict[str, Any]
JsonList = list[JsonDict]


@dataclass
class AdapterInput:
    """
    Универсальный вход для любого adapter.

    Для image classification:
        dataset_path = data/demo/dataset.csv
        data_dir = data/demo/images

    Для tabular:
        dataset_path = data/demo/table.csv
        data_dir можно оставить None

    Для YOLO:
        dataset_path = путь к metadata/csv или None
        data_dir = папка с images/labels
    """

    modality: str
    task_type: str
    dataset_path: str | Path | None
    data_dir: str | Path | None
    output_dir: str | Path
    project_id: str | None = None
    extra: dict[str, Any] = field(default_factory=dict)


@dataclass
class AdapterOutput:
    """
    Универсальный выход любого adapter.
    Его можно безопасно отдавать backend/frontend/agent.
    """

    status: str
    modality: str
    task_type: str

    dataset_metrics: JsonDict = field(default_factory=dict)
    object_metrics: JsonList = field(default_factory=list)
    review_queue: JsonList = field(default_factory=list)

    recommendations: JsonList = field(default_factory=list)
    roadmap: JsonList = field(default_factory=list)

    collection_tasks: JsonList = field(default_factory=list)
    synthetic_tasks: JsonList = field(default_factory=list)
    class_action_plan: JsonList = field(default_factory=list)

    exports: JsonDict = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)
    metadata: JsonDict = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return make_json_safe(
            {
                "status": self.status,
                "modality": self.modality,
                "task_type": self.task_type,
                "dataset_metrics": self.dataset_metrics,
                "object_metrics": self.object_metrics,
                "review_queue": self.review_queue,
                "recommendations": self.recommendations,
                "roadmap": self.roadmap,
                "collection_tasks": self.collection_tasks,
                "synthetic_tasks": self.synthetic_tasks,
                "class_action_plan": self.class_action_plan,
                "exports": self.exports,
                "warnings": self.warnings,
                "errors": self.errors,
                "metadata": self.metadata,
            }
        )


class ModalityAdapter(ABC):
    """
    Базовый интерфейс для всех adapters.
    Каждый adapter обязан реализовать run().
    """

    modality: str
    task_type: str

    @abstractmethod
    def run(self, adapter_input: AdapterInput) -> AdapterOutput:
        raise NotImplementedError


def make_success_output(
    modality: str,
    task_type: str,
    dataset_metrics: JsonDict | None = None,
    object_metrics: JsonList | None = None,
    review_queue: JsonList | None = None,
    recommendations: JsonList | None = None,
    roadmap: JsonList | None = None,
    collection_tasks: JsonList | None = None,
    synthetic_tasks: JsonList | None = None,
    class_action_plan: JsonList | None = None,
    exports: JsonDict | None = None,
    metadata: JsonDict | None = None,
    warnings: list[str] | None = None,
) -> AdapterOutput:
    return AdapterOutput(
        status="completed",
        modality=modality,
        task_type=task_type,
        dataset_metrics=dataset_metrics or {},
        object_metrics=object_metrics or [],
        review_queue=review_queue or [],
        recommendations=recommendations or [],
        roadmap=roadmap or [],
        collection_tasks=collection_tasks or [],
        synthetic_tasks=synthetic_tasks or [],
        class_action_plan=class_action_plan or [],
        exports=exports or {},
        metadata=metadata or {},
        warnings=warnings or [],
    )


def make_error_output(
    modality: str,
    task_type: str,
    error: str,
    metadata: JsonDict | None = None,
) -> AdapterOutput:
    return AdapterOutput(
        status="failed",
        modality=modality,
        task_type=task_type,
        errors=[error],
        metadata=metadata or {},
    )


def make_json_safe(value: Any) -> Any:
    """
    Приводит значения к JSON-friendly формату.

    Это нужно, потому что pandas/numpy часто возвращают типы,
    которые json.dump не умеет сохранять напрямую.
    """

    if value is None:
        return None

    if isinstance(value, Path):
        return str(value)

    if isinstance(value, dict):
        return {
            str(key): make_json_safe(val)
            for key, val in value.items()
        }

    if isinstance(value, list):
        return [make_json_safe(item) for item in value]

    if isinstance(value, tuple):
        return [make_json_safe(item) for item in value]

    try:
        import numpy as np

        if isinstance(value, np.integer):
            return int(value)

        if isinstance(value, np.floating):
            return float(value)

        if isinstance(value, np.ndarray):
            return value.tolist()

        if isinstance(value, float) and np.isnan(value):
            return None

    except Exception:
        pass

    try:
        import pandas as pd

        if pd.isna(value):
            return None

    except Exception:
        pass

    return value