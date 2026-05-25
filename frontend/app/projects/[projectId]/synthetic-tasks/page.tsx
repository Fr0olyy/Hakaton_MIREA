"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { Copy, Sparkles } from "lucide-react";
import { getSyntheticTasks, updateSyntheticTask } from "@/lib/api";
import type { SyntheticTask } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { PriorityBadge } from "@/components/status";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select } from "@/components/ui/select";

const statuses = ["open", "in_progress", "done", "rejected"];

export default function SyntheticTasksPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [tasks, setTasks] = useState<SyntheticTask[]>([]);
  const [localStatuses, setLocalStatuses] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getSyntheticTasks(projectId)
      .then(setTasks)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  async function changeStatus(task: SyntheticTask, nextStatus: string) {
    setLocalStatuses((current) => ({ ...current, [task.id]: nextStatus }));
    await updateSyntheticTask(projectId, task.id, { status: nextStatus }).catch(() => undefined);
  }

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading synthetic tasks" />
      </>
    );
  }
  if (error) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error} />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Synthetic Tasks" description="Synthetic data candidates that must be validated before they are used for training." />
      <div className="mb-4 rounded-md border border-amber-200 bg-amber-50 p-4 text-sm font-medium text-amber-900">Synthetic data must be validated before training.</div>

      {!tasks.length ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center gap-3 p-10 text-center">
            <Sparkles className="h-8 w-8 text-muted-foreground" />
            <div className="font-medium">No synthetic data is recommended for this dataset.</div>
            <p className="max-w-md text-sm text-muted-foreground">The current class plan does not require generated examples, or the agent has not produced synthetic tasks yet.</p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          {tasks.map((task) => (
            <Card key={task.id}>
              <CardHeader>
                <div className="flex items-start justify-between gap-3">
                  <CardTitle>{task.target || task.target_class || "Synthetic target"}</CardTitle>
                  <PriorityBadge priority={task.priority || "medium"} />
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid gap-2 text-sm sm:grid-cols-2">
                  <Info label="task_type" value={task.task_type || "generate_synthetic"} />
                  <Info label="target_count" value={String(task.target_count)} />
                  <Info label="expected_impact" value={task.expected_impact || "Improves weak-class coverage."} />
                  <Info label="owner_role" value={task.owner_role || "ML Engineer"} />
                  <Info label="risk" value={task.risk} />
                  <Info label="synthetic_bias_risk" value={task.synthetic_bias_risk || task.risk || "medium"} />
                  <Info label="requires_human_validation" value={String(task.requires_human_validation ?? true)} />
                  <div>
                    <div className="mb-1 text-xs text-muted-foreground">status</div>
                    <Select value={localStatuses[task.id] || normalizeStatus(task.status)} onChange={(event) => changeStatus(task, event.target.value)}>
                      {statuses.map((status) => (
                        <option key={status} value={status}>
                          {status}
                        </option>
                      ))}
                    </Select>
                  </div>
                </div>
                <PromptCard title="Prompt" text={task.prompt || "Generate diverse, realistic samples for the target class while preserving dataset style."} />
                {task.negative_prompt ? <PromptCard title="Negative prompt" text={task.negative_prompt} muted /> : null}
                <div className="flex flex-wrap gap-2">
                  <Badge variant={task.risk === "high" ? "danger" : task.risk === "medium" ? "warning" : "muted"}>{task.risk}</Badge>
                  <Button variant="outline" size="sm" onClick={() => navigator.clipboard.writeText(task.prompt || "")}>
                    <Copy />
                    Copy Prompt
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </>
  );
}

function PromptCard({ title, text, muted }: { title: string; text: string; muted?: boolean }) {
  return (
    <div className={`rounded-md border p-4 ${muted ? "bg-slate-50" : "bg-primary/5"}`}>
      <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-muted-foreground">{title}</div>
      <div className="whitespace-pre-wrap text-sm leading-6">{text}</div>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md bg-slate-50 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 break-words text-sm font-medium">{value}</div>
    </div>
  );
}

function normalizeStatus(status: string) {
  return status === "draft" ? "open" : status || "open";
}
