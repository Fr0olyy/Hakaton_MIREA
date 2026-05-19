from pathlib import Path

import cv2
import numpy as np
import pandas as pd
from PIL import Image

def analyze_single_image_quality(image_path: str | Path) -> dict:
    image_path = Path(image_path)
    if not image_path.exists():
        return {
            "blur_score": 0.0,
            "brightness": 0.0,
            "contrast": 0.0,
            "width": 0.0, 
            "height": 0.0,
            "resolution_score": 0.0,
            "corrupted_file": True,
            "quality_score": 0.0,
            "quality_reasons": "missing_file",
        }
    try:
        with Image.open(image_path) as img:
            img = img.convert("RGB")
            width, height = img.size()
            image_np = np.array(img)

    except Exception:
        return {
            "blur_score": 0.0,
            "brightness": 0.0,
            "contrast": 0.0,
            "width": 0.0, 
            "height": 0.0,
            "resolution_score": 0.0,
            "corrupted_file": True,
            "quality_score": 0.0,
            "quality_reasons": "missing_file",
        }
    
    gray = cv2.cvtColor(image_np, cv2.COLOR_RGB2GRAY)

    brightness = float(np.mean(gray))
    contrast = float(np.std(gray))
    blur_score = float(cv2.Laplacian(gray, cv2.CV_64F).var())

    min_side = min(width, height)

    if min_side >= 256:
        resolution_score = 1.0
    elif min_side >= 128:
        resolution_score = 0.7
    else:
        resolution_score = 0.3

    quality_score = 1.0
    reasons = []

    if resolution_score < 0.7:
        quality_score -= 0.25
        reasons.append("low_resolution")

    if blur_score < 60:
        quality_score -= 0.25
        reasons.append("blurry")

    if brightness < 35:
        quality_score -= 0.20
        reasons.append("too_dark")

    if brightness > 220:
        quality_score -= 0.20
        reasons.append("too_bright")

    if contrast < 20:
        quality_score -= 0.15
        reasons.append("low_contrast")

    quality_score = max(0.0, min(1.0, quality_score))

    return {
        "blur_score": blur_score,
        "brightness": brightness,
        "contrast": contrast,
        "width": int(width),
        "height": int(height),
        "resolution_score": float(resolution_score),
        "corrupted_file": False,
        "quality_score": float(quality_score),
        "quality_reasons": ";".join(reasons),
    }

def compute_image_quality(df: pd.DataFrame) -> pd.DataFrame:
    rows = []

    for _, row in df.iterrows():
        image_path = row.get("resolved_path") or row.get("absolute_path")

        if not image_path:
            image_path = row.get("file_path")

        quality = analyze_single_image_quality(image_path)

        rows.append({
            "id": row["id"],
            **quality,
        })

    quality_df = pd.DataFrame(rows)

    return df.merge(quality_df, on="id", how="left")