"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getClassActionPlan, getProject } from "@/lib/api";
import { isImageProject, isTabularProject, isYoloProject } from "@/lib/level2";
import type { ClassActionPlanItem, Project } from "@/lib/types";
import { formatScore } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { PriorityBadge } from "@/components/status";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function ClassActionPlanPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [items, setItems] = useState<ClassActionPlanItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([getProject(projectId), getClassActionPlan(projectId)])
      .then(([projectData, plan]) => {
        setProject(projectData);
        setItems(plan || []);
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading class action plan" />
      </>
    );
  }
  if (error || !project) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Class action plan is not available"} />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Class Action Plan" description="Class-level issues, recommended actions, priorities, and risks for the current dataset." />
      <Card>
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              {isImageProject(project) ? <Header columns={["target class", "objects_count", "target_count", "class_deficit_score", "review_objects_count", "label_error_count", "hard_examples_count", "recommended_action", "priority", "risk", "reason"]} /> : null}
              {isTabularProject(project) ? <Header columns={["missing_values", "numeric_outliers", "high_cardinality_columns", "recommended_action", "priority", "risk", "reason"]} /> : null}
              {isYoloProject(project) ? <Header columns={["invalid_bboxes", "tiny_boxes", "affected_files_count", "recommended_action", "priority", "risk", "reason"]} /> : null}
            </TableHeader>
            <TableBody>{items.map((item, index) => rowFor(project, item, index))}</TableBody>
          </Table>
          {!items.length ? <div className="p-8 text-center text-sm text-muted-foreground">No class-level actions are currently needed.</div> : null}
        </CardContent>
      </Card>
    </>
  );
}

function Header({ columns }: { columns: string[] }) {
  return (
    <TableRow>
      {columns.map((column) => (
        <TableHead key={column}>{column}</TableHead>
      ))}
    </TableRow>
  );
}

function rowFor(project: Project, item: ClassActionPlanItem, index: number) {
  const risk = item.risk || "low";
  const priority = item.priority || priorityFromScore(item.real_collection_priority || item.synthetic_data_candidate_score || 0);
  if (isImageProject(project)) {
    return (
      <TableRow key={`${item.class}-${index}`}>
        <TableCell className="font-medium">{item.target_class || item.class || "unknown"}</TableCell>
        <TableCell>{item.objects_count ?? "-"}</TableCell>
        <TableCell>{item.target_count ?? "-"}</TableCell>
        <TableCell>{formatScore(item.class_deficit_score ?? item.real_collection_priority, 2)}</TableCell>
        <TableCell>{item.review_objects_count ?? "-"}</TableCell>
        <TableCell>{item.label_error_count ?? "-"}</TableCell>
        <TableCell>{item.hard_examples_count ?? "-"}</TableCell>
        <TableCell>{item.recommended_action || "-"}</TableCell>
        <TableCell>
          <PriorityBadge priority={priority} />
        </TableCell>
        <TableCell>
          <RiskBadge risk={risk} />
        </TableCell>
        <TableCell className="max-w-80 text-muted-foreground">{item.reason || item.problem || item.expected_impact || "-"}</TableCell>
      </TableRow>
    );
  }
  if (isYoloProject(project)) {
    return (
      <TableRow key={`yolo-${index}`}>
        <TableCell>{item.invalid_bboxes ?? 0}</TableCell>
        <TableCell>{item.tiny_boxes ?? 0}</TableCell>
        <TableCell>{item.affected_files_count ?? 0}</TableCell>
        <TableCell>{item.recommended_action || "-"}</TableCell>
        <TableCell>
          <PriorityBadge priority={priority} />
        </TableCell>
        <TableCell>
          <RiskBadge risk={risk} />
        </TableCell>
        <TableCell className="max-w-80 text-muted-foreground">{item.reason || item.problem || item.expected_impact || "-"}</TableCell>
      </TableRow>
    );
  }
  return (
    <TableRow key={`tabular-${index}`}>
      <TableCell>{item.missing_values ?? 0}</TableCell>
      <TableCell>{item.numeric_outliers ?? 0}</TableCell>
      <TableCell>{item.high_cardinality_columns ?? 0}</TableCell>
      <TableCell>{item.recommended_action || "-"}</TableCell>
      <TableCell>
        <PriorityBadge priority={priority} />
      </TableCell>
      <TableCell>
        <RiskBadge risk={risk} />
      </TableCell>
      <TableCell className="max-w-80 text-muted-foreground">{item.reason || item.problem || item.expected_impact || "-"}</TableCell>
    </TableRow>
  );
}

function RiskBadge({ risk }: { risk: string }) {
  return <Badge variant={risk === "high" ? "danger" : risk === "medium" ? "warning" : "muted"}>{risk}</Badge>;
}

function priorityFromScore(score: number) {
  if (score >= 0.65) return "high";
  if (score >= 0.25) return "medium";
  return "low";
}
