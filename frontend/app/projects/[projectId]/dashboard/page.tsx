"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { AlertTriangle, Database, RefreshCw } from "lucide-react";
import { getDashboard, getProbabilisticAnalysis, getRecommendations, getReviewQueue, listAllObjects, listProjects } from "@/lib/api";
import { isImageProject, isTabularProject, isYoloProject, modalityLabel } from "@/lib/level2";
import type { Dashboard, DataObject, ProbabilisticAnalysis, Recommendation, ReviewQueueItem } from "@/lib/types";
import { dashboardCounters } from "@/lib/derived";
import { formatNumber, formatScore } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { LoadingState, ErrorState } from "@/components/state";
import { MetricCard } from "@/components/metric-card";
import { BarChart, DonutChart } from "@/components/charts";
import { AISummary } from "@/components/ai-summary";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { StatusBadge } from "@/components/status";

export default function DashboardPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const router = useRouter();
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [probabilistic, setProbabilistic] = useState<ProbabilisticAnalysis | null>(null);
  const [queue, setQueue] = useState<ReviewQueueItem[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [objects, setObjects] = useState<DataObject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  function load() {
    setLoading(true);
    setError("");
    Promise.all([getDashboard(projectId), getProbabilisticAnalysis(projectId), getReviewQueue(projectId), getRecommendations(projectId), listAllObjects(projectId)])
      .then(([dashboardData, probabilisticData, queueData, recommendationData, objectData]) => {
        setDashboard(dashboardData);
        setProbabilistic(probabilisticData);
        setQueue(queueData);
        setRecommendations(recommendationData);
        setObjects(objectData.objects || []);
      })
      .catch(async (err: Error) => {
        if (err.message === "not found") {
          const projects = await listProjects().catch(() => []);
          if (projects[0]) {
            router.replace(`/projects/${projects[0].id}/dashboard`);
            return;
          }
        }
        setError(err.message);
      })
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading dashboard" />
      </>
    );
  }

  if (error || !dashboard) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Dashboard is not available. Upload data and run analysis first."} />
      </>
    );
  }

  const counters = dashboardCounters(dashboard, probabilistic?.metrics || [], queue);
  const modality = dashboard.project.modality;
  const tabularColumns = inferMetadataColumns(objects);
  const yoloInvalid = countObjectsByNeed(queue, ["out_of_bounds", "invalid bbox", "invalid_bbox"]);
  const yoloTiny = countObjectsByNeed(queue, ["tiny", "small bbox", "tiny_boxes"]);

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        eyebrow={dashboard.project.name}
        title="Dashboard"
        description={`Latest dataset version: ${dashboard.dataset_version.version_name}. Analysis source: ${dashboard.analysis_source || "not analyzed"}.`}
        actions={
          <>
            <Button variant="outline" onClick={load}>
              <RefreshCw />
              Refresh
            </Button>
            <StatusBadge status={dashboard.dataset_version.status} />
            <StatusBadge status={modalityLabel(modality)} />
          </>
        }
      />

      <div className="mb-4 grid metric-grid gap-3">
        <MetricCard label="status" value={dashboard.dataset_version.status} tone="slate" />
        <MetricCard label="modality" value={modalityLabel(modality)} tone="blue" />
        <MetricCard label="task_type" value={dashboard.project.task_type} tone="slate" />
        <MetricCard label="objects_count" value={formatNumber(counters.totalObjects)} tone="blue" />
        <MetricCard label="review_queue_count" value={formatNumber(counters.reviewQueueCount)} tone="orange" />
        <MetricCard label="recommendations_count" value={formatNumber(recommendations.length)} tone="green" />
      </div>

      <div className="grid metric-grid gap-3">
        {isImageProject(dashboard.project) ? (
          <>
            <MetricCard
              label="total_objects"
              value={formatNumber(counters.totalObjects)}
              tone="blue"
              detail={
                <div className="space-y-1">
                  <div>dataset_version: {formatNumber(counters.countSources.datasetVersion || 0)}</div>
                  <div>status_counts sum: {formatNumber(counters.countSources.statusCounts || 0)}</div>
                  <div>metrics rows: {formatNumber(counters.countSources.metrics || 0)}</div>
                </div>
              }
            />
            <MetricCard label="classes_count" value={formatNumber(counters.classesCount)} tone="green" />
            <MetricCard label="dataset_readiness_score" value={formatScore(counters.readinessScore, 1)} hint="0..100" tone="green" />
            <MetricCard label="dataset_v2_objects" value={formatNumber(Math.max(0, counters.totalObjects - counters.reviewQueueCount))} tone="green" />
            <MetricCard label="duplicates_count" value={formatNumber(counters.duplicates)} tone="orange" />
            <MetricCard label="bad_quality_count" value={formatNumber(counters.badQuality)} tone="red" />
            <MetricCard label="suspected_label_errors_count" value={formatNumber(counters.suspectedLabelErrors)} tone="red" />
            <MetricCard label="hard_examples_count" value={formatNumber(countMetrics(probabilistic?.metrics || [], (metric) => metric.uncertainty_score >= 0.5))} tone="orange" />
          </>
        ) : null}
        {isTabularProject(dashboard.project) ? (
          <>
            <MetricCard label="tabular_quality_score" value={formatScore(counters.readinessScore, 1)} hint="0..100" tone="green" />
            <MetricCard label="total_rows" value={formatNumber(counters.totalObjects)} tone="blue" />
            <MetricCard label="total_columns" value={formatNumber(tabularColumns.length)} tone="slate" />
            <MetricCard label="missing_columns_count" value={formatNumber(countColumnsByName(tabularColumns, ["missing", "nan", "null"]))} tone="orange" />
            <MetricCard label="rows_with_any_nan_fraction" value={formatScore(fractionByReason(queue, ["missing", "nan", "null"]), 3)} tone="orange" />
            <MetricCard label="outliers_total_count" value={formatNumber(countObjectsByNeed(queue, ["outlier"]))} tone="red" />
            <MetricCard label="high_cardinality_columns_count" value={formatNumber(countColumnsByName(tabularColumns, ["cardinality", "unique"]))} tone="slate" />
          </>
        ) : null}
        {isYoloProject(dashboard.project) ? (
          <>
            <MetricCard label="cv_quality_score" value={formatScore(counters.readinessScore, 1)} hint="0..100" tone="green" />
            <MetricCard label="total_label_files" value={formatNumber(counters.totalObjects)} tone="blue" />
            <MetricCard label="files_needing_review" value={formatNumber(counters.reviewQueueCount)} tone="orange" />
            <MetricCard label="out_of_bounds_files_count" value={formatNumber(yoloInvalid)} tone="red" />
            <MetricCard label="tiny_boxes_files_count" value={formatNumber(yoloTiny)} tone="orange" />
          </>
        ) : null}
      </div>

      {counters.hasCountMismatch ? (
        <div className="mt-4 flex items-start gap-3 rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
          <div>
            Counts from backend endpoints differ. The main number uses the largest reliable source, and the detail under total_objects shows the raw values.
          </div>
        </div>
      ) : null}

      <Card className="mt-4">
        <CardContent className="grid gap-4 p-4 md:grid-cols-[220px_1fr] md:items-center">
          <div className="flex items-center gap-3">
            <span className="flex h-11 w-11 items-center justify-center rounded-md bg-primary/10 text-primary">
              <Database className="h-5 w-5" />
            </span>
            <div>
              <div className="font-semibold">Dataset version</div>
              <div className="text-sm text-muted-foreground">{dashboard.dataset_version.id}</div>
            </div>
          </div>
          <div className="grid gap-3 text-sm sm:grid-cols-4">
            <Info label="version_name" value={dashboard.dataset_version.version_name} />
            <Info label="objects_count" value={formatNumber(dashboard.dataset_version.objects_count)} />
            <Info label="readiness" value={formatScore(dashboard.dataset_version.readiness_score, 1)} />
            <Info label="created_at" value={new Date(dashboard.dataset_version.created_at).toLocaleString()} />
          </div>
        </CardContent>
      </Card>

      <div className="mt-4 grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>class_distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <DonutChart data={dashboard.class_distribution} valueKind="ratio" />
            <div className="mt-5">
              <BarChart data={dashboard.class_distribution} valueKind="ratio" color="blue" />
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>status_distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <DonutChart data={dashboard.status_counts} valueKind="count" />
            <div className="mt-5">
              <BarChart data={dashboard.status_counts} valueKind="count" color="green" />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="mt-4">
        <AISummary projectId={projectId} />
      </div>
    </>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0 rounded-md bg-slate-50 p-3">
      <div className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{label}</div>
      <div className="mt-1 truncate text-sm font-medium" title={value}>
        {value}
      </div>
    </div>
  );
}

function inferMetadataColumns(objects: DataObject[]) {
  return Array.from(new Set(objects.flatMap((object) => Object.keys(object.metadata || {}).filter((key) => key !== "probabilities" && key !== "validation_errors")))).sort();
}

function countMetrics<T>(items: T[], predicate: (item: T) => boolean) {
  return items.reduce((count, item) => count + (predicate(item) ? 1 : 0), 0);
}

function countColumnsByName(columns: string[], needles: string[]) {
  return columns.filter((column) => needles.some((needle) => column.toLowerCase().includes(needle))).length;
}

function countObjectsByNeed(queue: ReviewQueueItem[], needles: string[]) {
  return queue.filter((item) => {
    const text = `${item.object.status} ${item.metric.recommendation || ""} ${(item.metric.reasons || []).join(" ")}`.toLowerCase();
    return needles.some((needle) => text.includes(needle));
  }).length;
}

function fractionByReason(queue: ReviewQueueItem[], needles: string[]) {
  if (!queue.length) return 0;
  return countObjectsByNeed(queue, needles) / queue.length;
}
