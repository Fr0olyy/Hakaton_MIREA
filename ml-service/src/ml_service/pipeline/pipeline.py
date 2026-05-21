from pathlib import Path
import numpy as np

from ml_service.io.load_dataset import load_dataset
from ml_service.io.export_results import export_results_csv

from ml_service.validators.files import validate_images

from ml_service.features.probabilities import prepare_probabilities_and_confidence

from ml_service.metrics.uncertainty import uncertainty_metrics
from ml_service.metrics.entropy import entropy_metrics
from ml_service.metrics.distribution import distribution_metrics
from ml_service.metrics.label_error import compute_label_error_probability
from ml_service.metrics.scoring import compute_object_utility_score
from ml_service.metrics.readiness import compute_dataset_readiness_score

from ml_service.quality.image_quality import compute_image_quality
from ml_service.duplicates.phash_duplicates import compute_phash_duplicates

from ml_service.curation.status import assign_object_status
from ml_service.curation.reasons import generate_object_reasons
from ml_service.curation.recommendation import generate_object_recommendations

from ml_service.curation.review_queue import (
    build_review_queue,
    export_review_queue_csv,
)

from ml_service.curation.recommendations import (
    build_recommendations,
    export_recommendations_json,
)

from ml_service.curation.roadmap import (
    build_roadmap,
    export_roadmap_json,
)

from ml_service.curation.dataset_v2 import (
    build_dataset_v2,
    export_dataset_v2_csv,
)

from ml_service.curation.dataset_report import (
    build_dataset_report,
    export_dataset_report_json,
)

from ml_service.agent.build_context import (
    build_agent_context,
    export_agent_context_json,
)


def run_pipeline(
    dataset_path: str | Path,
    images_dir: str | Path,
    output_dir: str | Path,
) -> dict:
    dataset_path = Path(dataset_path)
    images_dir = Path(images_dir)
    output_dir = Path(output_dir)

    output_dir.mkdir(parents=True, exist_ok=True)

    print("1. Загрузка и очистка датасета...")
    df = load_dataset(dataset_path)

    print("2. Проверка файлов изображений...")
    df = validate_images(df, images_dir)

    print("3. Подготовка вероятностей классов...")
    df, prob_cols = prepare_probabilities_and_confidence(df)

    print("4. Расчёт uncertainty...")
    df = uncertainty_metrics(df)

    print("5. Расчёт entropy...")
    df = entropy_metrics(df, prob_cols)

    print("6. Расчёт распределения классов и class_deficit_score...")
    df = distribution_metrics(df)

    print("7. Анализ качества изображений...")
    df = compute_image_quality(df)

    print("8. Поиск дублей через pHash...")
    df = compute_phash_duplicates(df)

    print("9. Расчёт label_error_probability...")
    df = compute_label_error_probability(df, prob_cols)

    print("10. Расчёт object_utility_score...")
    df = compute_object_utility_score(df)

    print("11. Назначение status объектам...")
    df = assign_object_status(df)

    print("12. Генерация reasons...")
    df = generate_object_reasons(df)

    print("13. Генерация recommendation для объектов...")
    df = generate_object_recommendations(df)

    print("14. Расчёт dataset_readiness_score...")
    df, readiness_report = compute_dataset_readiness_score(df)

    print("15. Экспорт results.csv...")
    results_path = export_results_csv(df, output_dir)

    print("16. Формирование review_queue.csv...")
    review_queue = build_review_queue(df)
    review_queue_path = export_review_queue_csv(review_queue, output_dir)

    print("17. Генерация recommendations.json...")
    recommendations = build_recommendations(df)
    recommendations_path = export_recommendations_json(recommendations, output_dir)

    print("18. Генерация roadmap.json...")
    roadmap = build_roadmap(df)
    roadmap_path = export_roadmap_json(roadmap, output_dir)

    print("19. Генерация dataset_v2.csv...")
    dataset_v2 = build_dataset_v2(df)
    dataset_v2_path = export_dataset_v2_csv(dataset_v2, output_dir)

    print("20. Генерация dataset_report.json...")
    dataset_report = build_dataset_report(
        df=df,
        review_queue=review_queue,
        dataset_v2=dataset_v2,
        readiness_report=readiness_report,
    )
    dataset_report_path = export_dataset_report_json(dataset_report, output_dir)

    print("21. Генерация agent_context.json...")
    agent_context = build_agent_context(
        dataset_report=dataset_report,
        roadmap=roadmap,
        recommendations=recommendations,
        review_queue=review_queue,
    )
    agent_context_path = export_agent_context_json(agent_context, output_dir)

    result = {
        "status": "completed",
        "objects_count": len(df),
        "dataset_v2_objects_count": len(dataset_v2),
        "review_queue_count": len(review_queue),
        "dataset_readiness_score": readiness_report.get("dataset_readiness_score", 0.0),
        "output_dir": str(output_dir),
        "files": {
            "results_csv": str(results_path),
            "review_queue_csv": str(review_queue_path),
            "recommendations_json": str(recommendations_path),
            "roadmap_json": str(roadmap_path),
            "dataset_v2_csv": str(dataset_v2_path),
            "dataset_report_json": str(dataset_report_path),
            "agent_context_json": str(agent_context_path),
        },
        "data": {
            "readiness_report": readiness_report,
            "dataset_report": dataset_report,
            "recommendations": recommendations,
            "roadmap": roadmap,
        },
    }

    if 'quality_reasons' in df.columns:
            df['quality_reasons'] = df['quality_reasons'].replace(r'^\s*$', np.nan, regex=True)
            df['quality_reasons'] = df['quality_reasons'].fillna("ok")

    print("\nПайплайн успешно завершён.")
    print(f"Objects: {len(df)}")
    print(f"Dataset v2 objects: {len(dataset_v2)}")
    print(f"Review queue objects: {len(review_queue)}")
    print(f"Dataset readiness: {result['dataset_readiness_score']:.2f}/100")
    print(f"Output dir: {output_dir}")

    return result


def main() -> None:
    run_pipeline(
        dataset_path=Path("data/demo/dataset.csv"),
        images_dir=Path("data/demo/images"),
        output_dir=Path("outputs/demo"),
    )


if __name__ == "__main__":
    main()