from pathlib import Path
from typing import Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field

from ml_service.pipeline.pipeline import run_pipeline


app = FastAPI(
    title="DataForge ML Service",
    description="ML analysis service for dataset diagnostics and curation",
    version="0.1.0",
)


class AnalyzeRequest(BaseModel):
    dataset_path: str = Field(
        ...,
        description="Path to dataset.csv",
        examples=["data/demo/dataset.csv"],
    )
    images_dir: str = Field(
        ...,
        description="Path to images directory",
        examples=["data/demo/images"],
    )
    output_dir: str = Field(
        ...,
        description="Path to output directory",
        examples=["outputs/demo"],
    )


class AnalyzeResponse(BaseModel):
    status: str
    objects_count: int
    dataset_v2_objects_count: int
    review_queue_count: int
    dataset_readiness_score: float
    output_dir: str
    files: dict[str, str]
    data: dict[str, Any]


@app.get("/health")
def health_check() -> dict:
    return {
        "status": "ok",
        "service": "dataforge-ml-service",
    }


@app.post("/analyze", response_model=AnalyzeResponse)
def analyze_dataset(request: AnalyzeRequest) -> dict:
    dataset_path = Path(request.dataset_path)
    images_dir = Path(request.images_dir)
    output_dir = Path(request.output_dir)

    if not dataset_path.exists():
        raise HTTPException(
            status_code=400,
            detail=f"dataset_path not found: {dataset_path}",
        )

    if not images_dir.exists():
        raise HTTPException(
            status_code=400,
            detail=f"images_dir not found: {images_dir}",
        )

    try:
        result = run_pipeline(
            dataset_path=dataset_path,
            images_dir=images_dir,
            output_dir=output_dir,
        )
        return result

    except Exception as error:
        raise HTTPException(
            status_code=500,
            detail=f"ML analysis failed: {str(error)}",
        ) from error