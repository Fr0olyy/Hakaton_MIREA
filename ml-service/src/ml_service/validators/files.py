from pathlib import Path

import pandas as pd
from PIL import Image


def validate_images(df: pd.DataFrame, images_dir: Path) -> pd.DataFrame:
    df = df.copy()
    images_dir = Path(images_dir)

    file_exists_list = []
    image_opened_list = []
    error_list = []
    resolved_path_list = []

    for path_str in df["file_path"]:
        clean_path = str(path_str).strip()
        full_path = images_dir / clean_path

        # fallback: если file_path без подпапки, пробуем найти файл рекурсивно
        if not full_path.exists():
            matches = list(images_dir.rglob(Path(clean_path).name))
            if matches:
                full_path = matches[0]

        # ВАЖНО: resolved_path добавляем всегда, даже если файла нет
        resolved_path_list.append(str(full_path))

        if not full_path.exists():
            file_exists_list.append(False)
            image_opened_list.append(False)
            error_list.append("File not found")
            continue

        file_exists_list.append(True)

        try:
            with Image.open(full_path) as img:
                img.verify()

            image_opened_list.append(True)
            error_list.append("")

        except Exception as error:
            image_opened_list.append(False)
            error_list.append(str(error))

    df["file_exists"] = file_exists_list
    df["image_opened"] = image_opened_list
    df["image_error"] = error_list
    df["resolved_path"] = resolved_path_list

    return df