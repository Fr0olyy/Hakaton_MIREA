import type { Dashboard, ObjectMetric, ReviewQueueItem } from "@/lib/types";

export function dashboardCounters(dashboard: Dashboard, metrics: ObjectMetric[], queue: ReviewQueueItem[]) {
  const queueByObject = new Map(queue.map((item) => [item.object.id, item.object.status]));
  const countSources = {
    dashboard: dashboard.objects_count,
    datasetVersion: dashboard.dataset_version?.objects_count,
    statusCounts: sumValues(dashboard.status_counts),
    labelCounts: sumValues(dashboard.label_counts),
    metrics: metrics.length,
  };
  const reliableCounts = Object.values(countSources).filter((value) => typeof value === "number" && Number.isFinite(value) && value > 0);

  const suspectedLabelErrors = countWhere(metrics, (metric) => statusForMetric(metric, queueByObject) === "suspected_label_error" || metric.label_error_probability >= 0.5);
  const duplicates = countWhere(metrics, (metric) => statusForMetric(metric, queueByObject) === "duplicate" || metric.duplicate_score >= 0.8);
  const badQuality = countWhere(metrics, (metric) => statusForMetric(metric, queueByObject) === "bad_quality" || (metric.quality_score > 0 && metric.quality_score < 0.5));
  const rareClasses = Object.values(dashboard.class_distribution || {}).filter((share) => share > 0 && share < rareClassCutoff(dashboard)).length;

  return {
    totalObjects: reliableCounts.length ? Math.max(...reliableCounts) : 0,
    classesCount: dashboard.project.classes?.length || Object.keys(dashboard.class_distribution || dashboard.label_counts || {}).length,
    readinessScore: dashboard.readiness_score,
    reviewQueueCount: Math.max(dashboard.review_items || 0, queue.length),
    suspectedLabelErrors,
    duplicates,
    badQuality,
    rareClasses,
    countSources,
    hasCountMismatch: new Set(reliableCounts).size > 1,
  };
}

export function reasonOptions(queue: ReviewQueueItem[]) {
  return Array.from(new Set(queue.flatMap((item) => item.metric.reasons || []))).sort();
}

export function classOptions(queue: ReviewQueueItem[]) {
  return Array.from(new Set(queue.map((item) => item.object.label || "unlabeled"))).sort();
}

export function statusOptions(queue: ReviewQueueItem[]) {
  return Array.from(new Set(queue.map((item) => item.object.status || "unknown"))).sort();
}

function countWhere<T>(items: T[], predicate: (item: T) => boolean) {
  return items.reduce((count, item) => count + (predicate(item) ? 1 : 0), 0);
}

function sumValues(values: Record<string, number> | undefined) {
  return Object.values(values || {}).reduce((sum, value) => sum + (Number.isFinite(value) ? value : 0), 0);
}

function statusForMetric(metric: ObjectMetric, queueByObject: Map<string, string>) {
  return String(metric.probabilities?.ml_status || queueByObject.get(metric.object_id) || "");
}

function rareClassCutoff(dashboard: Dashboard) {
  const classes = dashboard.project.classes?.length || Object.keys(dashboard.class_distribution || {}).length || 1;
  return Math.min(0.1, 0.5 / classes);
}
