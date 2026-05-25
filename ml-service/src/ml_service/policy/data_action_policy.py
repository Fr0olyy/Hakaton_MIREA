from __future__ import annotations

from typing import Any


def apply_data_action_policy(adapter_output: dict[str, Any]) -> dict[str, Any]:
    """
    Enrich AdapterOutput dict with recommended actions.

    Adds to every object:
    - recommended_action
    - action_confidence
    - action_reason
    - action_risk
    """

    output = dict(adapter_output)

    modality = str(output.get("modality", "unknown"))

    object_metrics = output.get("object_metrics", [])
    review_queue = output.get("review_queue", [])

    output["object_metrics"] = [
        enrich_object_with_action(obj, modality)
        for obj in object_metrics
    ]

    output["review_queue"] = [
        enrich_object_with_action(obj, modality)
        for obj in review_queue
    ]

    output.setdefault("metadata", {})
    output["metadata"]["data_action_policy_applied"] = True

    return output


def enrich_object_with_action(
    obj: dict[str, Any],
    modality: str,
) -> dict[str, Any]:
    item = dict(obj)

    if modality == "image_classification":
        action = _image_classification_action(item)

    elif modality == "tabular_classification":
        action = _tabular_action(item)

    elif modality == "image_detection_yolo":
        action = _yolo_action(item)

    else:
        action = _default_action(item)

    item["recommended_action"] = action["recommended_action"]
    item["action_confidence"] = action["action_confidence"]
    item["action_reason"] = action["action_reason"]
    item["action_risk"] = action["action_risk"]

    return item


def _image_classification_action(obj: dict[str, Any]) -> dict[str, Any]:
    status = _clean_str(obj.get("status", "ok"))
    reasons = _parse_reasons(obj.get("reasons", []))

    duplicate_score = _safe_float(obj.get("duplicate_score", 0.0))
    quality_score = _safe_float(obj.get("quality_score", 1.0))
    label_error_probability = _safe_float(
        obj.get("label_error_probability", 0.0)
    )
    entropy = _safe_float(obj.get("entropy", 0.0))
    uncertainty_score = _safe_float(obj.get("uncertainty_score", 0.0))
    class_deficit_score = _safe_float(obj.get("class_deficit_score", 0.0))

    is_rare = (
        status == "rare_class_candidate"
        or class_deficit_score >= 0.5
        or "rare_class" in reasons
        or "class_deficit" in reasons
    )

    is_duplicate = (
        status == "duplicate"
        or duplicate_score >= 0.7
        or "duplicate" in reasons
    )

    is_bad_quality = (
        status == "bad_quality"
        or quality_score < 0.65
        or "bad_quality" in reasons
        or "blurry" in reasons
        or "low_resolution" in reasons
        or "too_dark" in reasons
        or "too_bright" in reasons
    )

    is_label_error = (
        status == "suspected_label_error"
        or label_error_probability >= 0.5
        or "suspected_label_error" in reasons
        or "model_disagreement" in reasons
    )

    is_hard_example = (
        status == "hard_example"
        or entropy >= 0.65
        or uncertainty_score >= 0.45
        or "high_entropy" in reasons
        or "low_confidence" in reasons
    )

    # 1. Label error is high priority: do not train before review
    if is_label_error:
        return {
            "recommended_action": "relabel_or_review",
            "action_confidence": _clip01(max(label_error_probability, 0.75)),
            "action_reason": "Object has high label error probability or model disagreement.",
            "action_risk": "high",
        }

    # 2. Duplicate logic
    if is_duplicate and is_rare:
        return {
            "recommended_action": "review_duplicate_rare_class",
            "action_confidence": _clip01(max(duplicate_score, 0.7)),
            "action_reason": "Object looks duplicated, but belongs to a rare class. Do not remove automatically.",
            "action_risk": "medium",
        }

    if is_duplicate:
        return {
            "recommended_action": "exclude_duplicate",
            "action_confidence": _clip01(max(duplicate_score, 0.8)),
            "action_reason": "Object is likely an exact or near duplicate.",
            "action_risk": "low",
        }

    # 3. Bad quality logic
    if is_bad_quality and is_rare:
        return {
            "recommended_action": "send_to_expert",
            "action_confidence": _clip01(max(1.0 - quality_score, 0.7)),
            "action_reason": "Object has low quality, but class is rare. Expert should decide before exclusion.",
            "action_risk": "high",
        }

    if is_bad_quality:
        return {
            "recommended_action": "exclude_bad_quality",
            "action_confidence": _clip01(max(1.0 - quality_score, 0.7)),
            "action_reason": "Object quality is too low for reliable training.",
            "action_risk": "medium",
        }

    # 4. Rare class logic
    if is_rare and quality_score >= 0.75:
        if class_deficit_score >= 0.75:
            return {
                "recommended_action": "collect_more_real_data_and_generate_synthetic",
                "action_confidence": _clip01(class_deficit_score),
                "action_reason": "Class is highly underrepresented. Need real collection and synthetic candidates.",
                "action_risk": "medium",
            }

        return {
            "recommended_action": "augment_existing_data",
            "action_confidence": _clip01(class_deficit_score),
            "action_reason": "Class is underrepresented and object quality is good. Useful for augmentation.",
            "action_risk": "low",
        }

    # 5. Hard examples are useful, not garbage
    if is_hard_example and quality_score >= 0.75:
        return {
            "recommended_action": "add_to_next_train",
            "action_confidence": _clip01(max(entropy, uncertainty_score, 0.65)),
            "action_reason": "Object is difficult but good quality, useful as a hard training example.",
            "action_risk": "medium",
        }

    return {
        "recommended_action": "keep",
        "action_confidence": 0.9,
        "action_reason": "No critical issues detected.",
        "action_risk": "low",
    }


def _tabular_action(obj: dict[str, Any]) -> dict[str, Any]:
    status = _clean_str(obj.get("status", "ok"))
    reasons = _parse_reasons(obj.get("reasons", []))

    if "duplicate_row" in reasons:
        return {
            "recommended_action": "review_or_remove_duplicate_row",
            "action_confidence": 0.85,
            "action_reason": "Row appears to be duplicated.",
            "action_risk": "medium",
        }

    if "outlier" in reasons:
        return {
            "recommended_action": "review_outlier",
            "action_confidence": 0.75,
            "action_reason": "Row contains statistical outlier values.",
            "action_risk": "medium",
        }

    if "missing_values" in reasons:
        return {
            "recommended_action": "fix_missing_values",
            "action_confidence": 0.8,
            "action_reason": "Row contains missing values.",
            "action_risk": "medium",
        }

    if status in {"warning", "needs_review"}:
        return {
            "recommended_action": "review_row",
            "action_confidence": 0.7,
            "action_reason": "Row was marked as problematic by tabular adapter.",
            "action_risk": "medium",
        }

    return {
        "recommended_action": "keep",
        "action_confidence": 0.9,
        "action_reason": "No critical tabular issues detected.",
        "action_risk": "low",
    }


def _yolo_action(obj: dict[str, Any]) -> dict[str, Any]:
    status = _clean_str(obj.get("status", "ok"))
    reasons = _parse_reasons(obj.get("reasons", []))
    cv_quality_score = _safe_float(obj.get("cv_quality_score", 100.0))

    if "bbox_out_of_bounds" in reasons:
        return {
            "recommended_action": "fix_invalid_bboxes",
            "action_confidence": 0.95,
            "action_reason": "YOLO annotation contains bounding boxes outside normalized 0..1 range.",
            "action_risk": "high",
        }

    if "tiny_boxes" in reasons:
        return {
            "recommended_action": "review_tiny_boxes",
            "action_confidence": 0.8,
            "action_reason": "YOLO annotation contains very small bounding boxes that may add noise.",
            "action_risk": "medium",
        }

    if status in {"warning", "needs_review"} or cv_quality_score < 75:
        return {
            "recommended_action": "review_annotations",
            "action_confidence": 0.75,
            "action_reason": "Detection annotation quality is below safe threshold.",
            "action_risk": "medium",
        }

    return {
        "recommended_action": "keep",
        "action_confidence": 0.9,
        "action_reason": "No critical YOLO annotation issues detected.",
        "action_risk": "low",
    }


def _default_action(obj: dict[str, Any]) -> dict[str, Any]:
    status = _clean_str(obj.get("status", "ok"))

    if status in {"warning", "error", "needs_review"}:
        return {
            "recommended_action": "review",
            "action_confidence": 0.7,
            "action_reason": "Object has non-ok status and should be reviewed.",
            "action_risk": "medium",
        }

    return {
        "recommended_action": "keep",
        "action_confidence": 0.8,
        "action_reason": "No critical issues detected.",
        "action_risk": "low",
    }


def _parse_reasons(value: Any) -> list[str]:
    if value is None:
        return []

    if isinstance(value, list):
        return [_clean_str(item) for item in value if _clean_str(item)]

    if isinstance(value, tuple):
        return [_clean_str(item) for item in value if _clean_str(item)]

    if isinstance(value, str):
        value = value.strip()

        if not value:
            return []

        # supports "a;b;c", "a,b,c", "['a', 'b']" roughly
        value = value.replace("[", "").replace("]", "").replace("'", "").replace('"', "")

        if ";" in value:
            return [_clean_str(item) for item in value.split(";") if _clean_str(item)]

        if "," in value:
            return [_clean_str(item) for item in value.split(",") if _clean_str(item)]

        return [_clean_str(value)]

    return [_clean_str(value)]


def _clean_str(value: Any) -> str:
    return str(value).strip()


def _safe_float(value: Any, default: float = 0.0) -> float:
    try:
        if value is None:
            return default

        text = str(value).strip()

        if text == "":
            return default

        return float(text)

    except Exception:
        return default


def _clip01(value: float) -> float:
    return max(0.0, min(1.0, float(value)))