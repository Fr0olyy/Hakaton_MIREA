from __future__ import annotations

import argparse
import csv
import random
import shutil
from pathlib import Path


LABEL_MAP = {
    "cane": "dog",
    "cavallo": "horse",
    "elefante": "elephant",
    "farfalla": "butterfly",
    "gallina": "chicken",
    "gatto": "cat",
    "mucca": "cow",
    "pecora": "sheep",
    "ragno": "spider",
    "scoiattolo": "squirrel",
}


IMAGE_EXTENSIONS = {".jpg", ".jpeg", ".png", ".webp"}


def get_label(folder_name: str) -> str:
    return LABEL_MAP.get(folder_name, folder_name)


def collect_images(raw_dir: Path) -> dict[str, list[Path]]:
    classes: dict[str, list[Path]] = {}

    for class_dir in raw_dir.iterdir():
        if not class_dir.is_dir():
            continue

        label = get_label(class_dir.name)

        images = [
            path
            for path in class_dir.rglob("*")
            if path.suffix.lower() in IMAGE_EXTENSIONS
        ]

        if images:
            classes[label] = images

    return classes


def split_name(index: int, total: int) -> str:
    train_end = int(total * 0.8)
    val_end = int(total * 0.9)

    if index < train_end:
        return "train"
    if index < val_end:
        return "val"
    return "test"


def make_predicted_label(
    true_label: str,
    all_labels: list[str],
    confidence_noise_rate: float,
) -> tuple[str, float]:
    """
    Для MVP мы пока генерируем predicted_label/confidence искусственно.
    Потом заменим это на реальную модель.
    """

    # Часть объектов делаем спорными: низкая уверенность.
    if random.random() < confidence_noise_rate:
        predicted_label = random.choice(all_labels)
        confidence = round(random.uniform(0.45, 0.68), 3)
        return predicted_label, confidence

    # Большинство объектов считаем уверенно распознанными.
    predicted_label = true_label
    confidence = round(random.uniform(0.82, 0.98), 3)
    return predicted_label, confidence


def maybe_corrupt_label(
    true_label: str,
    all_labels: list[str],
    label_noise_rate: float,
    split: str,
) -> str:
    """
    Искусственно портим часть label в train/val,
    чтобы система могла найти suspected_label_error.
    Test лучше не портить.
    """

    if split == "test":
        return true_label

    if random.random() >= label_noise_rate:
        return true_label

    other_labels = [label for label in all_labels if label != true_label]

    if not other_labels:
        return true_label

    return random.choice(other_labels)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--raw-dir",
        default="data/raw/animals10/raw-img",
        help="Папка, где лежат class folders Animals-10",
    )
    parser.add_argument(
        "--output-images-dir",
        default="data/demo/images",
        help="Куда скопировать выбранные изображения",
    )
    parser.add_argument(
        "--output-csv",
        default="data/demo/dataset.csv",
        help="Куда сохранить dataset.csv",
    )
    parser.add_argument(
        "--max-per-class",
        type=int,
        default=120,
        help="Сколько максимум картинок брать на класс",
    )
    parser.add_argument(
        "--label-noise-rate",
        type=float,
        default=0.06,
        help="Доля искусственно неправильных label",
    )
    parser.add_argument(
        "--confidence-noise-rate",
        type=float,
        default=0.12,
        help="Доля объектов с низкой уверенностью модели",
    )
    parser.add_argument("--seed", type=int, default=42)
    args = parser.parse_args()

    random.seed(args.seed)

    raw_dir = Path(args.raw_dir)
    output_images_dir = Path(args.output_images_dir)
    output_csv = Path(args.output_csv)

    output_images_dir.mkdir(parents=True, exist_ok=True)
    output_csv.parent.mkdir(parents=True, exist_ok=True)

    classes = collect_images(raw_dir)

    if not classes:
        raise RuntimeError(f"No images found in {raw_dir}")

    all_labels = sorted(classes.keys())
    rows = []
    object_id = 1

    for label in all_labels:
        images = classes[label]
        random.shuffle(images)

        selected_images = images[: args.max_per_class]
        total = len(selected_images)

        for index, source_path in enumerate(selected_images):
            split = split_name(index, total)

            file_name = f"{label}_{object_id:06d}{source_path.suffix.lower()}"
            target_path = output_images_dir / file_name

            shutil.copy2(source_path, target_path)

            dataset_label = maybe_corrupt_label(
                true_label=label,
                all_labels=all_labels,
                label_noise_rate=args.label_noise_rate,
                split=split,
            )

            predicted_label, confidence = make_predicted_label(
                true_label=label,
                all_labels=all_labels,
                confidence_noise_rate=args.confidence_noise_rate,
            )

            rows.append(
                {
                    "id": object_id,
                    "file_path": file_name,
                    "label": dataset_label,
                    "predicted_label": predicted_label,
                    "confidence": confidence,
                    "split": split,
                    "source": "animals10",
                    "annotator": f"ann_{random.randint(1, 4)}",
                    "true_label_for_demo": label,
                }
            )

            object_id += 1

    with output_csv.open("w", newline="", encoding="utf-8") as file:
        writer = csv.DictWriter(
            file,
            fieldnames=[
                "id",
                "file_path",
                "label",
                "predicted_label",
                "confidence",
                "split",
                "source",
                "annotator",
                "true_label_for_demo",
            ],
        )
        writer.writeheader()
        writer.writerows(rows)

    print(f"Created CSV: {output_csv}")
    print(f"Copied images to: {output_images_dir}")
    print(f"Rows: {len(rows)}")
    print(f"Classes: {all_labels}")


if __name__ == "__main__":
    main()