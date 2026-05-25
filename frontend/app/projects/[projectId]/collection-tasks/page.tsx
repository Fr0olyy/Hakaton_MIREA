"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams } from "next/navigation";
import { getCollectionTasks, updateCollectionTask } from "@/lib/api";
import type { CollectionTask } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { PriorityBadge } from "@/components/status";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Select } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

const statuses = ["open", "in_progress", "done", "rejected"];

export default function CollectionTasksPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [tasks, setTasks] = useState<CollectionTask[]>([]);
  const [localStatuses, setLocalStatuses] = useState<Record<string, string>>({});
  const [owner, setOwner] = useState("all");
  const [priority, setPriority] = useState("all");
  const [status, setStatus] = useState("all");
  const [taskType, setTaskType] = useState("all");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getCollectionTasks(projectId)
      .then(setTasks)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  const filtered = useMemo(() => {
    return tasks.filter((task) => {
      const currentStatus = localStatuses[task.id] || normalizeStatus(task.status);
      const currentOwner = task.owner_role || "Data Analyst";
      const currentType = task.task_type || "collect_real";
      return (
        (owner === "all" || currentOwner === owner) &&
        (priority === "all" || task.priority === priority) &&
        (status === "all" || currentStatus === status) &&
        (taskType === "all" || currentType === taskType)
      );
    });
  }, [tasks, localStatuses, owner, priority, status, taskType]);

  async function changeStatus(task: CollectionTask, nextStatus: string) {
    setLocalStatuses((current) => ({ ...current, [task.id]: nextStatus }));
    await updateCollectionTask(projectId, task.id, { status: nextStatus }).catch(() => undefined);
  }

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading collection tasks" />
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
      <PageHeader title="Collection Tasks" description="Operational tasks for real data collection, annotation, and dataset repair." />
      <div className="mb-4 grid gap-3 md:grid-cols-4">
        <Filter value={owner} onChange={setOwner} label="All owners" options={unique(tasks.map((task) => task.owner_role || "Data Analyst"))} />
        <Filter value={priority} onChange={setPriority} label="All priorities" options={unique(tasks.map((task) => task.priority || "medium"))} />
        <Filter value={status} onChange={setStatus} label="All statuses" options={statuses} />
        <Filter value={taskType} onChange={setTaskType} label="All task types" options={unique(tasks.map((task) => task.task_type || "collect_real"))} />
      </div>
      <Card>
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              <TableRow>
                {["task_type", "target_class / target", "target_count", "priority", "status", "reason", "expected_impact", "risk", "owner_role"].map((column) => (
                  <TableHead key={column}>{column}</TableHead>
                ))}
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((task) => (
                <TableRow key={task.id}>
                  <TableCell>{task.task_type || "collect_real"}</TableCell>
                  <TableCell className="font-medium">{task.target || task.target_class || "-"}</TableCell>
                  <TableCell>{task.target_count}</TableCell>
                  <TableCell>
                    <PriorityBadge priority={task.priority} />
                  </TableCell>
                  <TableCell>
                    <Select value={localStatuses[task.id] || normalizeStatus(task.status)} onChange={(event) => changeStatus(task, event.target.value)}>
                      {statuses.map((item) => (
                        <option key={item} value={item}>
                          {item}
                        </option>
                      ))}
                    </Select>
                  </TableCell>
                  <TableCell className="max-w-72 text-muted-foreground">{task.reason || "Class needs more reliable real examples."}</TableCell>
                  <TableCell className="max-w-72 text-muted-foreground">{task.expected_impact || "Improves class balance and review coverage."}</TableCell>
                  <TableCell>
                    <Badge variant={task.risk === "high" ? "danger" : task.risk === "medium" ? "warning" : "muted"}>{task.risk}</Badge>
                  </TableCell>
                  <TableCell>{task.owner_role || ownerForTask(task.task_type)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          {!filtered.length ? <div className="p-8 text-center text-sm text-muted-foreground">No collection tasks match these filters.</div> : null}
        </CardContent>
      </Card>
    </>
  );
}

function Filter({ value, onChange, label, options }: { value: string; onChange: (value: string) => void; label: string; options: string[] }) {
  return (
    <Select value={value} onChange={(event) => onChange(event.target.value)}>
      <option value="all">{label}</option>
      {options.map((option) => (
        <option key={option} value={option}>
          {option}
        </option>
      ))}
    </Select>
  );
}

function unique(values: string[]) {
  return Array.from(new Set(values.filter(Boolean))).sort();
}

function normalizeStatus(status: string) {
  return status === "draft" ? "open" : status || "open";
}

function ownerForTask(taskType?: string) {
  if (taskType?.includes("label")) return "Annotator";
  if (taskType?.includes("analysis")) return "Data Analyst";
  return "Domain Expert";
}
