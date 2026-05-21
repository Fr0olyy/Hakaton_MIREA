from pathlib import Path

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
    
    if 'quality_reasons' in df.columns:
        df['quality_reasons'] = df['quality_reasons'].replace(r'^\s*$', np.nan, regex=True)
        df['quality_reasons'] = df['quality_reasons'].fillna("ok")

    print("\nПайплайн работает! Вот первые 5 строк результата:")
    columns_to_show = ['id', 'confidence', 'uncertainty_score', 'entropy', 'class_deficit_score']
    print(df[columns_to_show].head(20))

if __name__ == "__main__":
    main()