"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { AlertTriangle, Database, RefreshCw } from "lucide-react";
import { getDashboard, getProbabilisticAnalysis, getReviewQueue } from "@/lib/api";
import type { Dashboard, ProbabilisticAnalysis, ReviewQueueItem } from "@/lib/types";
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
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [probabilistic, setProbabilistic] = useState<ProbabilisticAnalysis | null>(null);
  const [queue, setQueue] = useState<ReviewQueueItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  function load() {
    setLoading(true);
    setError("");
    Promise.all([getDashboard(projectId), getProbabilisticAnalysis(projectId), getReviewQueue(projectId)])
      .then(([dashboardData, probabilisticData, queueData]) => {
        setDashboard(dashboardData);
        setProbabilistic(probabilisticData);
        setQueue(queueData);
      })
      .catch((err: Error) => setError(err.message))
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
          </>
        }
      />

      <div className="grid metric-grid gap-3">
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
        <MetricCard label="review_queue_count" value={formatNumber(counters.reviewQueueCount)} tone="orange" />
        <MetricCard label="suspected_label_errors" value={formatNumber(counters.suspectedLabelErrors)} tone="red" />
        <MetricCard label="duplicates_count" value={formatNumber(counters.duplicates)} tone="orange" />
        <MetricCard label="bad_quality_count" value={formatNumber(counters.badQuality)} tone="red" />
        <MetricCard label="rare_classes_count" value={formatNumber(counters.rareClasses)} tone="slate" />
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
