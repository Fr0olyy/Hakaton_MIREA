"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { Check, Eye, MessageSquare, Send, ShieldX, Tag } from "lucide-react";
import { getReviewQueue } from "@/lib/api";
import { actionConfidence, actionReason, actionRisk, recommendedAction } from "@/lib/level2";
import type { ReviewQueueItem } from "@/lib/types";
import { apiPath, formatPercent, formatScore, numericProbabilityEntries } from "@/lib/utils";
import { classOptions, reasonOptions, statusOptions } from "@/lib/derived";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { StatusBadge } from "@/components/status";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Dialog } from "@/components/ui/dialog";
import { Select } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { BarChart } from "@/components/charts";

type Decision = {
  status: string;
  note?: string;
};

export default function ActiveLearningQueuePage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [queue, setQueue] = useState<ReviewQueueItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("all");
  const [reason, setReason] = useState("all");
  const [label, setLabel] = useState("all");
  const [quickFilter, setQuickFilter] = useState("all");
  const [selected, setSelected] = useState<ReviewQueueItem | null>(null);
  const [decisions, setDecisions] = useState<Record<string, Decision>>({});

  useEffect(() => {
    getReviewQueue(projectId)
      .then(setQueue)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  const filtered = useMemo(() => {
    return queue
      .filter((item) => {
        const itemStatus = decisions[item.object.id]?.status || item.object.status || "unknown";
        const reasons = item.metric.reasons || [];
        const itemLabel = item.object.label || "unlabeled";
        return (
          (status === "all" || itemStatus === status) &&
          (reason === "all" || reasons.includes(reason)) &&
          (label === "all" || itemLabel === label) &&
          matchesQuickFilter(item, quickFilter)
        );
      })
      .sort((a, b) => b.metric.object_utility_score - a.metric.object_utility_score);
  }, [queue, status, reason, label, quickFilter, decisions]);

  function applyDecision(item: ReviewQueueItem, nextStatus: string, note?: string) {
    setDecisions((current) => ({
      ...current,
      [item.object.id]: { status: nextStatus, note },
    }));
  }

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading active learning queue" />
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
      <PageHeader title="Active Learning Queue" description="Objects sorted by utility and review need. Use filters to focus on statuses, reasons, and classes." />

      <div className="mb-4 grid gap-3 md:grid-cols-4">
        <Select value={status} onChange={(event) => setStatus(event.target.value)}>
          <option value="all">All statuses</option>
          {statusOptions(queue).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
          {Array.from(new Set(Object.values(decisions).map((decision) => decision.status))).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </Select>
        <Select value={reason} onChange={(event) => setReason(event.target.value)}>
          <option value="all">All reasons</option>
          {reasonOptions(queue).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </Select>
        <Select value={label} onChange={(event) => setLabel(event.target.value)}>
          <option value="all">All classes</option>
          {classOptions(queue).map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </Select>
        <Select value={quickFilter} onChange={(event) => setQuickFilter(event.target.value)}>
          <option value="all">All risk groups</option>
          <option value="high">high risk</option>
          <option value="medium">medium risk</option>
          <option value="label_errors">label errors</option>
          <option value="bad_quality">bad quality</option>
          <option value="duplicates">duplicates</option>
          <option value="outliers">outliers</option>
          <option value="invalid_bboxes">invalid bboxes</option>
          <option value="tiny_boxes">tiny boxes</option>
        </Select>
      </div>

      <Card>
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>image preview</TableHead>
                <TableHead>label</TableHead>
                <TableHead>predicted_label</TableHead>
                <TableHead>confidence</TableHead>
                <TableHead>entropy</TableHead>
                <TableHead>object_utility_score</TableHead>
                <TableHead>status</TableHead>
                <TableHead>reasons</TableHead>
                <TableHead>recommended_action</TableHead>
                <TableHead>action_confidence</TableHead>
                <TableHead>action_reason</TableHead>
                <TableHead>action_risk</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((item) => {
                const decision = decisions[item.object.id];
                return (
                  <TableRow key={item.object.id} className={actionRisk(item.object, item.metric) === "high" ? "bg-red-50/60" : ""}>
                    <TableCell>
                      <button className="h-16 w-20 overflow-hidden rounded-md border bg-muted" onClick={() => setSelected(item)}>
                        {/* eslint-disable-next-line @next/next/no-img-element */}
                        <img
                          src={apiPath(`/api/projects/${projectId}/objects/${item.object.id}/file`)}
                          alt={item.object.file_path || item.object.id}
                          className="h-full w-full object-cover"
                          onError={(event) => {
                            event.currentTarget.style.display = "none";
                          }}
                        />
                      </button>
                    </TableCell>
                    <TableCell className="font-medium">{item.object.label || "unlabeled"}</TableCell>
                    <TableCell>{item.object.predicted_label || "-"}</TableCell>
                    <TableCell>{formatPercent(item.object.confidence || 0)}</TableCell>
                    <TableCell>{formatScore(item.metric.entropy, 3)}</TableCell>
                    <TableCell>{formatScore(item.metric.object_utility_score, 3)}</TableCell>
                    <TableCell>
                      <StatusBadge status={decision?.status || item.object.status} />
                    </TableCell>
                    <TableCell>
                      <div className="flex max-w-xs flex-wrap gap-1">
                        {(item.metric.reasons || []).map((reason) => (
                          <Badge key={reason} variant="outline">
                            {reason}
                          </Badge>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="max-w-48 text-muted-foreground">{recommendedAction(item.object, item.metric)}</TableCell>
                    <TableCell>{formatPercent(actionConfidence(item.object, item.metric))}</TableCell>
                    <TableCell className="max-w-56 text-muted-foreground">{actionReason(item.object, item.metric)}</TableCell>
                    <TableCell>
                      <Badge variant={actionRisk(item.object, item.metric) === "high" ? "danger" : actionRisk(item.object, item.metric) === "medium" ? "warning" : "muted"}>
                        {actionRisk(item.object, item.metric)}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button variant="outline" size="sm" onClick={() => setSelected(item)}>
                        <Eye />
                        Open
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
          {!filtered.length ? <div className="p-8 text-center text-sm text-muted-foreground">No queue items match the current filters.</div> : null}
        </CardContent>
      </Card>

      <ObjectCard
        projectId={projectId}
        item={selected}
        decision={selected ? decisions[selected.object.id] : undefined}
        onClose={() => setSelected(null)}
        onDecision={applyDecision}
      />
    </>
  );
}

function ObjectCard({
  projectId,
  item,
  decision,
  onClose,
  onDecision,
}: {
  projectId: string;
  item: ReviewQueueItem | null;
  decision?: Decision;
  onClose: () => void;
  onDecision: (item: ReviewQueueItem, status: string, note?: string) => void;
}) {
  if (!item) return null;

  const probabilityData = Object.fromEntries(numericProbabilityEntries(item.metric.probabilities).map((entry) => [entry.label, entry.value]));
  const metricRows = [
    ["confidence", item.object.confidence || 0],
    ["entropy", item.metric.entropy],
    ["uncertainty_score", item.metric.uncertainty_score],
    ["label_error_probability", item.metric.label_error_probability],
    ["duplicate_score", item.metric.duplicate_score],
    ["quality_score", item.metric.quality_score],
    ["rarity_score", item.metric.rarity_score],
    ["class_deficit_score", item.metric.class_deficit_score],
    ["novelty_score", item.metric.novelty_score],
    ["object_utility_score", item.metric.object_utility_score],
    ["final_score", item.metric.final_score],
  ];

  return (
    <Dialog open={Boolean(item)} onOpenChange={(open) => !open && onClose()} title="Object Card" description={item.object.file_path || item.object.id}>
      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <div className="space-y-3">
          <div className="aspect-square overflow-hidden rounded-lg border bg-muted">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img src={apiPath(`/api/projects/${projectId}/objects/${item.object.id}/file`)} alt={item.object.file_path || item.object.id} className="h-full w-full object-contain" />
          </div>
          <div className="grid grid-cols-2 gap-2 text-sm">
            <Info label="label" value={item.object.label || "unlabeled"} />
            <Info label="predicted_label" value={item.object.predicted_label || "-"} />
            <Info label="status" value={decision?.status || item.object.status} />
            <Info label="recommendation" value={item.metric.recommendation || "-"} />
          </div>
          {decision?.note ? <div className="rounded-md border bg-muted p-3 text-sm text-muted-foreground">{decision.note}</div> : null}
        </div>

        <div className="space-y-4">
          <div>
            <h3 className="mb-2 text-sm font-semibold">Class probabilities</h3>
            <BarChart data={probabilityData} valueKind="ratio" maxItems={12} />
          </div>

          <div>
            <h3 className="mb-2 text-sm font-semibold">All metrics</h3>
            <div className="grid gap-2 sm:grid-cols-2">
              {metricRows.map(([label, value]) => (
                <Info key={label as string} label={label as string} value={formatScore(value as number, 3)} />
              ))}
            </div>
          </div>

          <div>
            <h3 className="mb-2 text-sm font-semibold">Queue reasons</h3>
            <div className="flex flex-wrap gap-2">
              {(item.metric.reasons || []).map((reason) => (
                <Badge key={reason} variant="warning">
                  {reason}
                </Badge>
              ))}
            </div>
          </div>

          <div className="flex flex-wrap gap-2 border-t pt-4">
            <Button variant="secondary" onClick={() => onDecision(item, "approved")}>
              <Check />
              Approve
            </Button>
            <Button variant="destructive" onClick={() => onDecision(item, "exclude_candidate")}>
              <ShieldX />
              Exclude
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                const nextLabel = window.prompt("New label", item.object.label || "");
                if (nextLabel) onDecision(item, "relabel_requested", `Relabel to: ${nextLabel}`);
              }}
            >
              <Tag />
              Relabel
            </Button>
            <Button variant="outline" onClick={() => onDecision(item, "expert_review")}>
              <Send />
              Send to expert review
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                const comment = window.prompt("Comment");
                if (comment) onDecision(item, decision?.status || item.object.status, comment);
              }}
            >
              <MessageSquare />
              Add Comment
            </Button>
          </div>
        </div>
      </div>
    </Dialog>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-card p-2">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 break-words text-sm font-medium">{value}</div>
    </div>
  );
}

function matchesQuickFilter(item: ReviewQueueItem, filter: string) {
  if (filter === "all") return true;
  if (filter === "high" || filter === "medium") return actionRisk(item.object, item.metric) === filter;
  const text = `${item.object.status} ${item.metric.recommendation || ""} ${(item.metric.reasons || []).join(" ")}`.toLowerCase();
  const needles: Record<string, string[]> = {
    label_errors: ["label", "suspected_label_error"],
    bad_quality: ["bad_quality", "quality", "blur", "dark"],
    duplicates: ["duplicate"],
    outliers: ["outlier"],
    invalid_bboxes: ["invalid_bbox", "invalid bbox", "out_of_bounds"],
    tiny_boxes: ["tiny", "small bbox", "tiny_boxes"],
  };
  return (needles[filter] || []).some((needle) => text.includes(needle));
}
