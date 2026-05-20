import sys
from pathlib import Path

# Добавляем корень папки ml-service в пути поиска Python, чтобы он увидел папку src
sys.path.append(str(Path(__file__).resolve().parent.parent))


from ml_service.io.load_dataset import load_dataset
from ml_service.validators.files import validate_images
from ml_service.features.probabilities import prepare_probabilities_and_confidence

#импорты метрик
from ml_service.metrics.uncertainty import uncertainty_metrics
from ml_service.metrics.distribution import distribution_metrics
from ml_service.metrics.entropy import entropy_metrics


def main():
    # Пути к данным (относительно папки ml-service)
    csv_path = Path("data/demo/dataset.csv")
    images_dir = Path("data/demo/images")

    print("1. Загрузка и очистка датасета...")
    df = load_dataset(csv_path)

    print("2. Проверка файлов изображений...")
    df = validate_images(df, images_dir)

    print("3. Подготовка вероятностей классов...")
    df, prob_cols = prepare_probabilities_and_confidence(df)
    
    print('4. Расчет метрик')
    df = uncertainty_metrics(df)
    df = entropy_metrics(df, prob_cols)
    df = distribution_metrics(df)
    

    print("\nПайплайн работает! Вот первые 5 строк результата:")
    columns_to_show = ['id', 'file_path', 'file_exists', 'confidence'] + prob_cols[:2]
    print(df[columns_to_show].head())

if __name__ == "__main__":
    main()