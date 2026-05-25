"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Check, MessageSquare, Send, ShieldX, Tag } from "lucide-react";
import { addObjectComment, createObjectAction, getObject, getProject } from "@/lib/api";
import { actionReason, actionRisk, isImageProject, isTabularProject, isYoloProject, recommendedAction } from "@/lib/level2";
import { apiPath, formatPercent, formatScore, numericProbabilityEntries } from "@/lib/utils";
import type { ObjectDetail, Project } from "@/lib/types";
import { useProjectRole } from "@/components/role-selector";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { StatusBadge } from "@/components/status";
import { BarChart } from "@/components/charts";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

export default function ObjectDetailPage() {
  const { projectId, objectId } = useParams<{ projectId: string; objectId: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [object, setObject] = useState<ObjectDetail | null>(null);
  const [comment, setComment] = useState("");
  const [localActions, setLocalActions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const { role } = useProjectRole();

  useEffect(() => {
    Promise.all([getProject(projectId), getObject(projectId, objectId)])
      .then(([projectData, objectData]) => {
        setProject(projectData);
        setObject(objectData);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId, objectId]);

  async function applyAction(action: string, newValue = "") {
    if (!object) return;
    setLocalActions((items) => [`${action}${newValue ? `: ${newValue}` : ""}`, ...items]);
    await createObjectAction(projectId, object.id, { action, old_value: object.label || "", new_value: newValue, comment }).catch(() => undefined);
  }

  async function submitComment() {
    if (!object || !comment.trim()) return;
    setLocalActions((items) => [`comment: ${comment.trim()}`, ...items]);
    await addObjectComment(projectId, object.id, comment.trim()).catch(() => undefined);
    setComment("");
  }

  const probabilityData = useMemo(() => Object.fromEntries(numericProbabilityEntries(object?.metrics?.probabilities).map((entry) => [entry.label, entry.value])), [object]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading object card" />
      </>
    );
  }
  if (error || !project || !object) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Object is not available"} />
      </>
    );
  }

  const metric = object.metrics;
  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        eyebrow={project.name}
        title="Object Card"
        description={object.file_path || object.external_id || object.id}
        actions={
          <>
            <StatusBadge status={object.status} />
            <Badge variant={actionRisk(object, metric) === "high" ? "danger" : actionRisk(object, metric) === "medium" ? "warning" : "muted"}>{actionRisk(object, metric)}</Badge>
          </>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <div className="space-y-4">
          {isImageProject(project) ? (
            <Card>
              <CardContent className="p-3">
                <div className="aspect-square overflow-hidden rounded-md border bg-muted">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img src={apiPath(`/api/projects/${projectId}/objects/${object.id}/file`)} alt={object.file_path || object.id} className="h-full w-full object-contain" />
                </div>
              </CardContent>
            </Card>
          ) : null}
          {isTabularProject(project) ? <MetadataCard title="Row values" values={object.metadata || {}} /> : null}
          {isYoloProject(project) ? <MetadataCard title="Annotation file" values={{ file_path: object.file_path, ...(object.metadata || {}) }} /> : null}

          <Card>
            <CardHeader>
              <CardTitle>Why this object needs attention</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2 text-sm text-muted-foreground">
              <p>{actionReason(object, metric)}</p>
              <div className="flex flex-wrap gap-2">
                {(metric?.reasons || []).map((reason) => (
                  <Badge key={reason} variant="warning">
                    {reason}
                  </Badge>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        <div className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Decision context</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3 md:grid-cols-3">
              <Info label="label" value={object.label || "unlabeled"} />
              <Info label="predicted_label" value={object.predicted_label || "-"} />
              <Info label="confidence" value={formatPercent(object.confidence || 0)} />
              <Info label="recommended_action" value={recommendedAction(object, metric)} />
              <Info label="action_reason" value={actionReason(object, metric)} />
              <Info label="status" value={object.status} />
            </CardContent>
          </Card>

          {isImageProject(project) ? (
            <Card>
              <CardHeader>
                <CardTitle>Probabilities</CardTitle>
              </CardHeader>
              <CardContent>
                <BarChart data={probabilityData} valueKind="ratio" />
              </CardContent>
            </Card>
          ) : null}

          <Card>
            <CardHeader>
              <CardTitle>All metrics</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-2 md:grid-cols-3">
              {metricRows(metric).map(([label, value]) => (
                <Info key={label} label={label} value={formatScore(value, 3)} />
              ))}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Comments and actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="flex flex-wrap gap-2">
                <Button variant="secondary" onClick={() => applyAction("approve")}>
                  <Check />
                  Approve
                </Button>
                <Button variant="destructive" onClick={() => applyAction("exclude")}>
                  <ShieldX />
                  Exclude
                </Button>
                <Button
                  variant="outline"
                  onClick={() => {
                    const next = window.prompt("New label", object.label || "");
                    if (next) applyAction("relabel", next);
                  }}
                >
                  <Tag />
                  Relabel
                </Button>
                <Button variant="outline" onClick={() => applyAction("send_to_expert")}>
                  <Send />
                  Send to Expert
                </Button>
                {role === "domain_expert" ? (
                  <Button variant="secondary" onClick={() => applyAction("expert_approve")}>
                    <Check />
                    Send Final Decision
                  </Button>
                ) : null}
              </div>
              <div className="flex gap-2">
                <Input value={comment} onChange={(event) => setComment(event.target.value)} placeholder="Add comment" />
                <Button variant="outline" onClick={submitComment}>
                  <MessageSquare />
                  Add
                </Button>
              </div>
              <div className="space-y-2 text-sm text-muted-foreground">
                {localActions.map((item, index) => (
                  <div key={`${item}-${index}`} className="rounded-md border bg-slate-50 p-2">
                    {item}
                  </div>
                ))}
                {(object.comments || []).map((item) => (
                  <div key={item.id} className="rounded-md border bg-slate-50 p-2">
                    {item.text}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </>
  );
}

function MetadataCard({ title, values }: { title: string; values: Record<string, unknown> }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent className="max-h-[420px] overflow-auto">
        <div className="grid gap-2">
          {Object.entries(values).map(([key, value]) => (
            <Info key={key} label={key} value={typeof value === "object" ? JSON.stringify(value) : String(value ?? "-")} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-card p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 break-words text-sm font-medium">{value}</div>
    </div>
  );
}

function metricRows(metric: ObjectDetail["metrics"]) {
  return [
    ["entropy", metric?.entropy || 0],
    ["uncertainty", metric?.uncertainty_score || 0],
    ["quality_score", metric?.quality_score || 0],
    ["duplicate_score", metric?.duplicate_score || 0],
    ["label_error_probability", metric?.label_error_probability || 0],
    ["object_utility_score", metric?.object_utility_score || 0],
    ["class_deficit_score", metric?.class_deficit_score || 0],
    ["final_score", metric?.final_score || 0],
  ] as Array<[string, number]>;
}
