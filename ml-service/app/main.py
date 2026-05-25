from pathlib import Path
from typing import Any

from fastapi import FastAPI
from pydantic import BaseModel, Field

from ml_service.pipeline.multimodal_pipeline import run_multimodal_pipeline


app = FastAPI(title="DataForge ML Service", version="2.0.0")


class AnalyzeRequest(BaseModel):
    modality: str = Field(default="image_classification")
    task_type: str = Field(default="classification")
    dataset_path: str | None = None
    data_dir: str | None = None
    output_dir: str = Field(default="outputs/api_analysis")
    project_id: str | None = None
    extra: dict[str, Any] = Field(default_factory=dict)


@app.get("/health")
def health():
    return {
        "status": "ok",
        "service": "ml-service",
        "version": "2.0.0",
    }


@app.post("/analyze")
def analyze(request: AnalyzeRequest):
    result = run_multimodal_pipeline(
        modality=request.modality,
        task_type=request.task_type,
        dataset_path=request.dataset_path,
        data_dir=request.data_dir,
        output_dir=request.output_dir,
        project_id=request.project_id,
        extra=request.extra,
    )

    return result