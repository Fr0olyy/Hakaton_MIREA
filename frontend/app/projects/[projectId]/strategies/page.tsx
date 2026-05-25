"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Download } from "lucide-react";
import { getDashboard, listAllObjects } from "@/lib/api";
import type { Dashboard, DataObject } from "@/lib/types";
import { downloadBlob, formatNumber } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const strategies = [
  {
    key: "conservative",
    title: "Conservative",
    risk: "low",
    use: "Use for fastest safe iteration when review time is limited.",
    removes: "Only obvious duplicates, invalid rows, and bad quality objects.",
  },
  {
    key: "balanced",
    title: "Balanced",
    risk: "medium",
    use: "Use as the default candidate for training after spot-checking the queue.",
    removes: "High-risk suspected label errors, strong duplicates, and low quality objects.",
  },
  {
    key: "aggressive",
    title: "Aggressive",
    risk: "high",
    use: "Use when precision is more important than retaining borderline data.",
    removes: "All review candidates with high utility/risk scores plus quality issues.",
  },
];

export default function DatasetStrategiesPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [objects, setObjects] = useState<DataObject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([getDashboard(projectId), listAllObjects(projectId)])
      .then(([dashboardData, objectData]) => {
        setDashboard(dashboardData);
        setObjects(objectData.objects || []);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  const strategyObjects = useMemo(() => Object.fromEntries(strategies.map((strategy) => [strategy.key, filterObjects(objects, strategy.key)])), [objects]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading dataset strategies" />
      </>
    );
  }
  if (error || !dashboard) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Strategies are not available"} />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Dataset Strategies" description="Conservative is safest, Aggressive is strictest." />

      <div className="mb-4 grid gap-3 md:grid-cols-4">
        <Compare label="raw_objects_count" value={objects.length || dashboard.objects_count} />
        <Compare label="conservative_objects_count" value={strategyObjects.conservative.length} />
        <Compare label="balanced_objects_count" value={strategyObjects.balanced.length} />
        <Compare label="aggressive_objects_count" value={strategyObjects.aggressive.length} />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        {strategies.map((strategy) => {
          const rows = strategyObjects[strategy.key] || [];
          return (
            <Card key={strategy.key}>
              <CardHeader>
                <div className="flex items-start justify-between gap-3">
                  <CardTitle>{strategy.title}</CardTitle>
                  <Badge variant={strategy.risk === "high" ? "danger" : strategy.risk === "medium" ? "warning" : "success"}>{strategy.risk} risk</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <Compare label="objects_count" value={rows.length} />
                <Info label="what is removed" value={strategy.removes} />
                <Info label="when to use" value={strategy.use} />
                <Button variant="outline" onClick={() => downloadStrategy(strategy.key, rows)}>
                  <Download />
                  Download CSV
                </Button>
              </CardContent>
            </Card>
          );
        })}
      </div>
    </>
  );
}

function Compare({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-md border bg-card p-4">
      <div className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</div>
      <div className="mt-2 text-2xl font-semibold">{formatNumber(value)}</div>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md bg-slate-50 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-sm">{value}</div>
    </div>
  );
}

function filterObjects(objects: DataObject[], strategy: string) {
  return objects.filter((object) => {
    const metric = object.metrics;
    if (strategy === "conservative") {
      return !["duplicate", "bad_quality", "invalid", "missing_file"].includes(object.status) && (metric?.duplicate_score || 0) < 0.95 && (metric?.quality_score || 1) >= 0.25;
    }
    if (strategy === "balanced") {
      return !["duplicate", "bad_quality", "suspected_label_error", "exclude_candidate", "invalid", "missing_file"].includes(object.status) && (metric?.final_score || 0) < 0.75;
    }
    return object.status === "ok" && (metric?.final_score || 0) < 0.45 && (metric?.quality_score || 1) >= 0.5 && (metric?.label_error_probability || 0) < 0.5;
  });
}

function downloadStrategy(strategy: string, rows: DataObject[]) {
  const header = ["id", "file_path", "label", "predicted_label", "confidence", "status"];
  const body = rows.map((object) => [object.external_id || object.id, object.file_path || "", object.label || "", object.predicted_label || "", object.confidence || 0, object.status].map(csvEscape).join(","));
  downloadBlob(new Blob([[header.join(","), ...body].join("\n")], { type: "text/csv" }), `dataset_v2_${strategy}.csv`);
}

function csvEscape(value: string | number) {
  const text = String(value);
  if (/[",\n]/.test(text)) return `"${text.replace(/"/g, '""')}"`;
  return text;
}
