"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Bot, Wand2 } from "lucide-react";
import { getClassActionPlan, getCollectionTasks, getDashboard, getRecommendations, getReviewQueue, getRoadmap, getSyntheticTasks, postAgentEndpoint } from "@/lib/api";
import type { AgentResponse, ClassActionPlanItem, CollectionTask, Dashboard, Recommendation, ReviewQueueItem, RoadmapItem, SyntheticTask } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function AgentSummaryPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [roadmap, setRoadmap] = useState<RoadmapItem[]>([]);
  const [queue, setQueue] = useState<ReviewQueueItem[]>([]);
  const [classPlan, setClassPlan] = useState<ClassActionPlanItem[]>([]);
  const [collectionTasks, setCollectionTasks] = useState<CollectionTask[]>([]);
  const [syntheticTasks, setSyntheticTasks] = useState<SyntheticTask[]>([]);
  const [agentText, setAgentText] = useState("");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      getDashboard(projectId),
      getRecommendations(projectId),
      getRoadmap(projectId),
      getReviewQueue(projectId),
      getClassActionPlan(projectId),
      getCollectionTasks(projectId),
      getSyntheticTasks(projectId),
    ])
      .then(([dash, recs, road, review, plan, collection, synthetic]) => {
        setDashboard(dash);
        setRecommendations(recs || []);
        setRoadmap(road || []);
        setQueue(review || []);
        setClassPlan(plan || []);
        setCollectionTasks(collection || []);
        setSyntheticTasks(synthetic || []);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  async function generate(label: string, endpoint: "summary" | "collection-plan" | "synthetic-plan" | "dataset-summary") {
    if (!dashboard) return;
    setBusy(label);
    setError("");
    try {
      const response: AgentResponse = await postAgentEndpoint(projectId, endpoint);
      setAgentText(response.summary || fallbackSummary(label, dashboard, queue, collectionTasks, syntheticTasks));
    } catch {
      setAgentText(fallbackSummary(label, dashboard, queue, collectionTasks, syntheticTasks));
    } finally {
      setBusy("");
    }
  }

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading AI summary context" />
      </>
    );
  }
  if (error || !dashboard) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Agent context is not available"} />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="AI Data Curator Summary" description="Agent-ready context and generated explanations for dataset curation decisions." />

      <div className="mb-4 flex flex-wrap gap-2">
        <Button onClick={() => generate("dataset", "dataset-summary")} disabled={busy !== ""}>
          <Wand2 />
          Generate Dataset Summary
        </Button>
        <Button variant="outline" onClick={() => generate("queue", "summary")} disabled={busy !== ""}>
          Explain Review Queue
        </Button>
        <Button variant="outline" onClick={() => generate("collection", "collection-plan")} disabled={busy !== ""}>
          Explain Collection Plan
        </Button>
        <Button variant="outline" onClick={() => generate("synthetic", "synthetic-plan")} disabled={busy !== ""}>
          Explain Synthetic Plan
        </Button>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_380px]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Bot className="h-5 w-5" />
              Agent output
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="min-h-56 whitespace-pre-wrap rounded-md border bg-slate-50 p-4 text-sm leading-6">
              {busy ? "Generating..." : agentText || fallbackSummary("dataset", dashboard, queue, collectionTasks, syntheticTasks)}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>project_summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <Info label="objects" value={String(dashboard.objects_count)} />
            <Info label="review_queue" value={String(dashboard.review_items)} />
            <Info label="readiness" value={dashboard.readiness_score.toFixed(1)} />
            <Info label="modality" value={dashboard.project.modality} />
          </CardContent>
        </Card>
      </div>

      <div className="mt-4 grid gap-4 lg:grid-cols-3">
        <ContextCard title="top_problems" items={topProblems(dashboard, queue)} />
        <ContextCard title="recommended_actions_summary" items={recommendations.slice(0, 5).map((item) => `${item.priority}: ${item.title}`)} />
        <ContextCard title="class_action_plan" items={classPlan.slice(0, 5).map((item) => `${item.class || item.target_class || "target"}: ${item.recommended_action || item.problem}`)} />
        <ContextCard title="collection_tasks" items={collectionTasks.slice(0, 5).map((item) => `${item.target_class}: ${item.target_count} (${item.priority})`)} />
        <ContextCard title="synthetic_tasks" items={syntheticTasks.slice(0, 5).map((item) => `${item.target_class}: ${item.prompt || "generate"}`)} />
        <ContextCard title="dataset_strategies" items={["Conservative: safest", "Balanced: default", "Aggressive: strictest"]} />
        <ContextCard title="top_roadmap_items" items={roadmap.slice(0, 5).map((item) => `${item.priority}. ${item.title}`)} />
      </div>
    </>
  );
}

function fallbackSummary(kind: string, dashboard: Dashboard, queue: ReviewQueueItem[], collectionTasks: CollectionTask[], syntheticTasks: SyntheticTask[]) {
  return [
    `Project ${dashboard.project.name} contains ${dashboard.objects_count} objects with readiness ${dashboard.readiness_score.toFixed(1)}.`,
    `The review queue has ${dashboard.review_items} objects. Main visible problems: ${topProblems(dashboard, queue).join("; ") || "none"}.`,
    kind === "collection" ? `Collection focus: ${collectionTasks.slice(0, 3).map((item) => item.target_class).join(", ") || "no open collection tasks"}.` : "",
    kind === "synthetic" ? `Synthetic focus: ${syntheticTasks.slice(0, 3).map((item) => item.target_class).join(", ") || "no synthetic data recommended"}.` : "",
  ]
    .filter(Boolean)
    .join("\n\n");
}

function ContextCard({ title, items }: { title: string; items: string[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {items.length ? (
          items.map((item) => (
            <div key={item} className="rounded-md bg-slate-50 p-2">
              {item}
            </div>
          ))
        ) : (
          <div className="text-muted-foreground">No data yet</div>
        )}
      </CardContent>
    </Card>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md bg-slate-50 p-2">
      <span className="text-muted-foreground">{label}</span>
      <Badge variant="muted">{value}</Badge>
    </div>
  );
}

function topProblems(dashboard: Dashboard, queue: ReviewQueueItem[]) {
  const statuses = Object.entries(dashboard.status_counts || {})
    .filter(([status, count]) => status !== "ok" && count > 0)
    .map(([status, count]) => `${status}: ${count}`);
  const reasons = queue.flatMap((item) => item.metric.reasons || []).slice(0, 5);
  return Array.from(new Set([...statuses, ...reasons])).slice(0, 6);
}
