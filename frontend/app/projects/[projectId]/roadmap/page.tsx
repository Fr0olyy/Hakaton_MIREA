"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { CheckCircle2, CircleDot, Flag, Layers, Route, Sparkles } from "lucide-react";
import { getRoadmap } from "@/lib/api";
import type { RoadmapItem } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { PriorityBadge } from "@/components/status";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { BarChart } from "@/components/charts";

export default function RoadmapPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [items, setItems] = useState<RoadmapItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getRoadmap(projectId)
      .then((data) => setItems(data.slice().sort((a, b) => a.priority - b.priority)))
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        title="Dataset Roadmap"
        description="Prioritized next steps to move the dataset from Level 1 diagnostics to a cleaner training-ready version."
      />
      {loading ? <LoadingState label="Loading roadmap" /> : null}
      {error ? <ErrorState message={error} /> : null}
      {!loading && !error ? (
        <div className="grid gap-4 lg:grid-cols-[1fr_360px]">
          <div className="space-y-4">
            <RoadmapRail items={items} />
          </div>
          <div className="space-y-4">
            <Card>
              <CardContent className="space-y-4 p-4">
                <div className="flex items-center gap-2 font-semibold">
                  <Route className="h-4 w-4 text-primary" />
                  Roadmap visualization
                </div>
                <BarChart data={actionDistribution(items)} valueKind="count" color="orange" maxItems={8} />
              </CardContent>
            </Card>
            <Card>
              <CardContent className="grid grid-cols-3 gap-3 p-4 text-center">
                <SmallStat label="steps" value={String(items.length)} />
                <SmallStat label="high priority" value={String(items.filter((item) => item.priority <= 1).length)} />
                <SmallStat label="actions" value={String(Object.keys(actionDistribution(items)).length)} />
              </CardContent>
            </Card>
          </div>
          {!items.length ? <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">No roadmap items yet. Run analysis first.</div> : null}
        </div>
      ) : null}
    </>
  );
}

function RoadmapRail({ items }: { items: RoadmapItem[] }) {
  return (
    <div className="relative space-y-3">
      <div className="absolute bottom-4 left-[23px] top-4 hidden w-px bg-border md:block" />
      {items.map((item, index) => {
        const Icon = iconForAction(item.action_type);
        return (
          <Card key={item.id} className="relative overflow-hidden">
            <CardContent className="grid gap-4 p-4 md:grid-cols-[48px_minmax(0,1fr)_minmax(180px,240px)] md:items-start">
              <div className="relative z-10 flex h-12 w-12 items-center justify-center rounded-full border bg-card text-primary shadow-sm">
                <Icon className="h-5 w-5" />
              </div>
              <div className="min-w-0">
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <PriorityBadge priority={item.priority} />
                  <Badge variant="outline">{item.action_type || "action"}</Badge>
                  <span className="text-xs text-muted-foreground">step {index + 1}</span>
                </div>
                <h3 className="text-base font-semibold leading-6 text-slate-950">{item.title}</h3>
                <p className="mt-2 text-sm leading-6 text-muted-foreground">{item.description}</p>
              </div>
              <div className="rounded-md border bg-slate-50 p-3 text-sm leading-5 text-muted-foreground">
                <div className="mb-1 flex items-center gap-2 font-medium text-slate-800">
                  <Sparkles className="h-3.5 w-3.5 text-accent" />
                  Expected impact
                </div>
                {item.expected_impact || "Improves dataset readiness."}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}

function SmallStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md bg-slate-50 p-3">
      <div className="text-2xl font-semibold text-slate-950">{value}</div>
      <div className="mt-1 text-xs text-muted-foreground">{label}</div>
    </div>
  );
}

function actionDistribution(items: RoadmapItem[]) {
  return items.reduce<Record<string, number>>((counts, item) => {
    const key = item.action_type || "action";
    counts[key] = (counts[key] || 0) + 1;
    return counts;
  }, {});
}

function iconForAction(action: string) {
  const normalized = action.toLowerCase();
  if (normalized.includes("export")) return CheckCircle2;
  if (normalized.includes("collect")) return Layers;
  if (normalized.includes("fix")) return Flag;
  return CircleDot;
}
