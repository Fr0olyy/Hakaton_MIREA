from __future__ import annotations

from typing import Any


def build_synthetic_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    modality = str(adapter_output.get("modality", "unknown"))

    if modality == "image_classification":
        return _build_image_synthetic_tasks(adapter_output)

    if modality == "image_detection_yolo":
        return _build_yolo_synthetic_tasks(adapter_output)

    return []


def _build_image_synthetic_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    class_action_plan = adapter_output.get("class_action_plan", [])
    tasks = []

    for item in class_action_plan:
        action = str(item.get("recommended_action", "")).strip()
        target_class = str(item.get("target", "")).strip()
        priority = str(item.get("priority", "medium")).strip()
        risk = str(item.get("risk", "medium")).strip()

        if not target_class:
            continue

        if action == "collect_more_real_data_and_generate_synthetic":
            target_count = int(item.get("target_count", 30) or 30)

            tasks.append(
                {
                    "task_type": "generate_synthetic_images",
                    "target_class": target_class,
                    "target_count": max(target_count, 20),
                    "priority": priority,
                    "status": "draft",
                    "reason": item.get(
                        "reason",
                        "Class is underrepresented and needs synthetic candidates.",
                    ),
                    "prompt": _build_image_prompt(target_class),
                    "negative_prompt": (
                        "cartoon, low quality, blurry, distorted anatomy, "
                        "wrong class, watermark, text, duplicate image"
                    ),
                    "expected_impact": (
                        "Increase class diversity and reduce rare-class weakness. "
                        "Synthetic samples must be validated before training."
                    ),
                    "risk": risk,
                    "synthetic_bias_risk": "medium",
                    "requires_human_validation": True,
                    "owner_role": "ML Engineer",
                    "source_action": action,
                }
            )

        elif action == "augment_existing_data":
            target_count = int(item.get("target_count", 10) or 10)

            tasks.append(
                {
                    "task_type": "augment_existing_images",
                    "target_class": target_class,
                    "target_count": max(target_count, 5),
                    "priority": priority,
                    "status": "draft",
                    "reason": item.get(
                        "reason",
                        "Class can benefit from safe augmentation.",
                    ),
                    "augmentations": [
                        "horizontal_flip",
                        "small_rotation",
                        "brightness_shift",
                        "contrast_shift",
                        "random_crop",
                    ],
                    "expected_impact": (
                        "Improve class robustness without changing semantic label."
                    ),
                    "risk": risk,
                    "synthetic_bias_risk": "low",
                    "requires_human_validation": False,
                    "owner_role": "ML Engineer",
                    "source_action": action,
                }
            )

    return tasks


def _build_yolo_synthetic_tasks(adapter_output: dict[str, Any]) -> list[dict[str, Any]]:
    class_action_plan = adapter_output.get("class_action_plan", [])
    tasks = []

    for item in class_action_plan:
        action = str(item.get("recommended_action", "")).strip()

        if action == "review_tiny_boxes":
            tasks.append(
                {
                    "task_type": "collect_or_generate_detection_edge_cases",
                    "target": "small_objects",
                    "target_count": int(item.get("affected_files_count", 10) or 10),
                    "priority": "medium",
                    "status": "draft",
                    "reason": "Detection dataset contains tiny boxes. More valid small-object examples may be needed.",
                    "prompt": (
                        "Generate realistic images with small but visible objects, "
                        "clear background, valid object boundaries, detection-friendly composition."
                    ),
                    "negative_prompt": (
                        "object too small to see, blurry, occluded object, invalid bbox, distorted scene"
                    ),
                    "expected_impact": "Improve detector stability on small objects.",
                    "risk": "medium",
                    "synthetic_bias_risk": "medium",
                    "requires_human_validation": True,
                    "owner_role": "ML Engineer",
                    "source_action": action,
                }
            )

    return tasks


def _build_image_prompt(target_class: str) -> str:
    return (
        f"Generate realistic photos of {target_class} in diverse environments, "
        f"different lighting conditions, backgrounds, camera angles and poses. "
        f"Images should be natural, sharp, correctly labeled as {target_class}, "
        f"and useful for image classification training."
    )