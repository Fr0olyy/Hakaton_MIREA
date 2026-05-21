from pathlib import Path

import imagehash
import pandas as pd
from PIL import Image


def compute_single_phash(image_path: str | Path):
    image_path = Path(str(image_path))

    if not image_path.exists():
        return None

    try:
        with Image.open(image_path) as img:
            return imagehash.phash(img)
    except Exception:
        return None


def distance_to_duplicate_score(distance: int) -> float:
    if distance <= 4:
        return 1.0

    if distance <= 8:
        return 0.7

    return 0.0


def compute_phash_duplicates(df: pd.DataFrame) -> pd.DataFrame:
    df = df.copy()

    hashes = {}

    for _, row in df.iterrows():
        object_id = str(row["id"])
        resolved_path = str(row.get("resolved_path", "")).strip()

        if not resolved_path:
            continue

        img_hash = compute_single_phash(resolved_path)

        if img_hash is not None:
            hashes[object_id] = img_hash

    duplicate_scores = {str(row["id"]): 0.0 for _, row in df.iterrows()}
    duplicate_group_ids = {str(row["id"]): "" for _, row in df.iterrows()}
    duplicate_neighbor_ids = {str(row["id"]): "" for _, row in df.iterrows()}
    duplicate_distances = {str(row["id"]): None for _, row in df.iterrows()}
    phash_values = {str(row["id"]): "" for _, row in df.iterrows()}

    for object_id, img_hash in hashes.items():
        phash_values[object_id] = str(img_hash)

    hash_items = list(hashes.items())
    group_counter = 1

    for i in range(len(hash_items)):
        id_a, hash_a = hash_items[i]

        for j in range(i + 1, len(hash_items)):
            id_b, hash_b = hash_items[j]

            distance = hash_a - hash_b
            score = distance_to_duplicate_score(distance)

            if score <= 0:
                continue

            existing_group = duplicate_group_ids[id_a] or duplicate_group_ids[id_b]

            if existing_group:
                group_id = existing_group
            else:
                group_id = f"dup_{group_counter}"
                group_counter += 1

            if score > duplicate_scores[id_a]:
                duplicate_scores[id_a] = score
                duplicate_neighbor_ids[id_a] = id_b
                duplicate_distances[id_a] = int(distance)

            if score > duplicate_scores[id_b]:
                duplicate_scores[id_b] = score
                duplicate_neighbor_ids[id_b] = id_a
                duplicate_distances[id_b] = int(distance)

            duplicate_group_ids[id_a] = group_id
            duplicate_group_ids[id_b] = group_id

    df["phash"] = df["id"].astype(str).map(phash_values).fillna("")
    df["duplicate_score"] = df["id"].astype(str).map(duplicate_scores).fillna(0.0)
    df["duplicate_group_id"] = df["id"].astype(str).map(duplicate_group_ids).fillna("")
    df["duplicate_neighbor_id"] = df["id"].astype(str).map(duplicate_neighbor_ids).fillna("")
    df["duplicate_distance"] = df["id"].astype(str).map(duplicate_distances)

    return df