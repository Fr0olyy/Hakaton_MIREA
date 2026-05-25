import type { DataObject, Modality, ObjectMetric, Project, ProjectRole, ReviewQueueItem } from "@/lib/types";

export const modalityOptions: Array<{ value: Modality; label: string; taskType: string; description: string }> = [
  {
    value: "image_classification",
    label: "Image Classification",
    taskType: "classification",
    description: "dataset.csv plus images.zip with class labels and optional model probabilities.",
  },
  {
    value: "tabular_classification",
    label: "Tabular Classification",
    taskType: "classification",
    description: "A single tabular.csv with rows, features, and optional target label.",
  },
  {
    value: "image_detection_yolo",
    label: "YOLO Detection",
    taskType: "detection",
    description: "A YOLO dataset archive with labels, images, or a manifest CSV.",
  },
];

export const roleOptions: Array<{ value: ProjectRole; label: string }> = [
  { value: "admin", label: "Project Admin" },
  { value: "ml_engineer", label: "ML Engineer" },
  { value: "annotator", label: "Annotator" },
  { value: "domain_expert", label: "Domain Expert" },
  { value: "data_analyst", label: "Data Analyst" },
];

export function modalityLabel(modality?: string) {
  return modalityOptions.find((option) => option.value === modality)?.label || (modality === "image" ? "Image Classification" : modality || "unknown");
}

export function taskTypeForModality(modality: string) {
  return modalityOptions.find((option) => option.value === modality)?.taskType || "classification";
}

export function isImageProject(project?: Project | null) {
  return project?.modality === "image" || project?.modality === "image_classification";
}

export function isTabularProject(project?: Project | null) {
  return project?.modality === "tabular_classification";
}

export function isYoloProject(project?: Project | null) {
  return project?.modality === "image_detection_yolo";
}

export function recommendedAction(object?: DataObject, metric?: ObjectMetric) {
  const raw = stringMetadata(object, "recommended_action") || stringMetadata(object, "action") || metric?.recommendation || object?.status || "review";
  return raw.replace(/\s+/g, "_").toLowerCase();
}

export function actionRisk(object?: DataObject, metric?: ObjectMetric) {
  const raw = stringMetadata(object, "action_risk") || stringMetadata(object, "risk");
  if (raw) return raw.toLowerCase();
  const score = metric?.final_score ?? metric?.object_utility_score ?? 0;
  if (score >= 0.7) return "high";
  if (score >= 0.45) return "medium";
  return "low";
}

export function actionConfidence(object?: DataObject, metric?: ObjectMetric) {
  const raw = object?.metadata?.action_confidence;
  if (typeof raw === "number") return raw;
  if (typeof raw === "string") {
    const parsed = Number(raw);
    if (Number.isFinite(parsed)) return parsed;
  }
  return Math.max(metric?.final_score || 0, metric?.object_utility_score || 0);
}

export function actionReason(object?: DataObject, metric?: ObjectMetric) {
  return stringMetadata(object, "action_reason") || metric?.reasons?.join(", ") || metric?.recommendation || "Object should be reviewed before it is used for training.";
}

export function objectIdLabel(object: DataObject) {
  return object.external_id || object.file_path || String(object.metadata?.row_number || object.id);
}

export function reviewItemFromObject(object: DataObject): ReviewQueueItem {
  return {
    object,
    metric: object.metrics || emptyMetric(object.id),
  };
}

export function emptyMetric(objectId: string): ObjectMetric {
  return {
    id: `${objectId}-metric`,
    object_id: objectId,
    entropy: 0,
    uncertainty_score: 0,
    label_error_probability: 0,
    duplicate_score: 0,
    rarity_score: 0,
    class_deficit_score: 0,
    quality_score: 0,
    novelty_score: 0,
    object_utility_score: 0,
    final_score: 0,
    reasons: [],
    recommendation: "",
    probabilities: {},
    created_at: "",
  };
}

export function stringMetadata(object: DataObject | undefined, key: string) {
  const value = object?.metadata?.[key];
  return typeof value === "string" ? value : "";
}
