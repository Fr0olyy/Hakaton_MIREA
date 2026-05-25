"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Eye, Search } from "lucide-react";
import { getProject, listObjects } from "@/lib/api";
import { actionRisk, isImageProject, isTabularProject, isYoloProject, objectIdLabel, recommendedAction } from "@/lib/level2";
import { apiPath, formatPercent, formatScore } from "@/lib/utils";
import type { DataObject, Project } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { StatusBadge } from "@/components/status";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

export default function DatasetExplorerPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [objects, setObjects] = useState<DataObject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [status, setStatus] = useState("all");
  const [action, setAction] = useState("all");
  const [risk, setRisk] = useState("all");
  const [label, setLabel] = useState("all");
  const [sort, setSort] = useState("risk");
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      getProject(projectId),
      listObjects(projectId, {
        page,
        per_page: 100,
        search,
        status,
        label,
        sort_by: sort === "utility" ? "object_utility_score" : sort === "quality" ? "quality_score" : sort === "confidence" ? "" : "final_score",
      }),
    ])
      .then(([projectData, objectData]) => {
        setProject(projectData);
        setObjects(objectData.objects || []);
        setTotal(objectData.total || 0);
        setTotalPages(objectData.total_pages || 1);
        setError("");
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId, sort, page, search, status, label]);

  useEffect(() => {
    setPage(1);
  }, [sort, search, status, label]);

  const rows = useMemo(() => {
    const query = search.trim().toLowerCase();
    return objects
      .filter((object) => {
        const metric = object.metrics;
        const text = `${object.id} ${object.external_id || ""} ${object.file_path || ""} ${object.label || ""}`.toLowerCase();
        return (
          (action === "all" || recommendedAction(object, metric) === action) &&
          (risk === "all" || actionRisk(object, metric) === risk) &&
          (!query || text.includes(query))
        );
      })
      .sort((a, b) => compareObjects(a, b, sort));
  }, [objects, status, action, risk, label, sort, search]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading dataset explorer" />
      </>
    );
  }

  if (error || !project) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Project is not available"} />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Universal Dataset Explorer" description="Browse object_metrics with modality-specific columns, filters, search, and risk-aware sorting." />

      <div className="mb-4 grid gap-3 md:grid-cols-6">
        <div className="relative md:col-span-2">
          <Search className="pointer-events-none absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input className="pl-9" placeholder="Search id, file_path, label" value={search} onChange={(event) => setSearch(event.target.value)} />
        </div>
        <FilterSelect value={status} onChange={setStatus} label="All statuses" options={unique(objects.map((object) => object.status || "unknown"))} />
        <FilterSelect value={action} onChange={setAction} label="All actions" options={unique(objects.map((object) => recommendedAction(object, object.metrics)))} />
        <FilterSelect value={risk} onChange={setRisk} label="All risks" options={["high", "medium", "low"]} />
        <Select value={sort} onChange={(event) => setSort(event.target.value)}>
          <option value="risk">Sort by risk</option>
          <option value="status">Sort by status</option>
          <option value="utility">Sort by utility score</option>
          <option value="quality">Sort by quality score</option>
          <option value="confidence">Sort by confidence</option>
        </Select>
      </div>

      <div className="mb-4 max-w-xs">
        <FilterSelect value={label} onChange={setLabel} label="All classes" options={unique(objects.map((object) => object.label || "unlabeled"))} />
      </div>

      <Card>
        <CardContent className="overflow-x-auto p-0">
          <Table>
            <TableHeader>
              {isImageProject(project) ? <ImageHeader /> : null}
              {isTabularProject(project) ? <TabularHeader /> : null}
              {isYoloProject(project) ? <YoloHeader /> : null}
            </TableHeader>
            <TableBody>
              {rows.map((object) => (
                <ExplorerRow key={object.id} project={project} object={object} projectId={projectId} />
              ))}
            </TableBody>
          </Table>
          {!rows.length ? <div className="p-8 text-center text-sm text-muted-foreground">No objects match the current filters.</div> : null}
        </CardContent>
      </Card>

      <div className="mt-4 flex flex-wrap items-center justify-between gap-3 text-sm text-muted-foreground">
        <span>
          Showing {rows.length} of {total} objects, page {page} of {totalPages}
        </span>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" disabled={page <= 1 || loading} onClick={() => setPage((value) => Math.max(1, value - 1))}>
            Previous
          </Button>
          <Button variant="outline" size="sm" disabled={page >= totalPages || loading} onClick={() => setPage((value) => Math.min(totalPages, value + 1))}>
            Next
          </Button>
        </div>
      </div>
    </>
  );
}

function ExplorerRow({ project, object, projectId }: { project: Project; object: DataObject; projectId: string }) {
  const metric = object.metrics;
  const risk = actionRisk(object, metric);
  if (isImageProject(project)) {
    return (
      <TableRow>
        <TableCell className="font-medium">
          <Link className="hover:underline" href={`/projects/${projectId}/objects/${object.id}`}>
            {objectIdLabel(object)}
          </Link>
        </TableCell>
        <TableCell>
          <Preview projectId={projectId} object={object} />
        </TableCell>
        <TableCell>{object.label || "unlabeled"}</TableCell>
        <TableCell>{object.predicted_label || "-"}</TableCell>
        <TableCell>{formatPercent(object.confidence || 0)}</TableCell>
        <TableCell>{formatScore(metric?.entropy, 3)}</TableCell>
        <TableCell>{formatScore(metric?.quality_score, 2)}</TableCell>
        <TableCell>{formatScore(metric?.duplicate_score, 2)}</TableCell>
        <TableCell>{formatScore(metric?.label_error_probability, 2)}</TableCell>
        <TableCell>
          <StatusBadge status={object.status} />
        </TableCell>
        <TableCell>{recommendedAction(object, metric)}</TableCell>
        <TableCell>
          <RiskBadge risk={risk} />
        </TableCell>
        <TableCell>
          <OpenButton projectId={projectId} objectId={object.id} />
        </TableCell>
      </TableRow>
    );
  }
  if (isYoloProject(project)) {
    return (
      <TableRow>
        <TableCell className="font-medium">{objectIdLabel(object)}</TableCell>
        <TableCell className="max-w-64 truncate">{object.file_path || "-"}</TableCell>
        <TableCell>{formatScore(metric?.quality_score || metric?.final_score, 2)}</TableCell>
        <TableCell>
          <StatusBadge status={object.status} />
        </TableCell>
        <TableCell className="max-w-72">{(metric?.reasons || []).join(", ") || "-"}</TableCell>
        <TableCell>{metric?.recommendation || "-"}</TableCell>
        <TableCell>{recommendedAction(object, metric)}</TableCell>
        <TableCell>
          <RiskBadge risk={risk} />
        </TableCell>
        <TableCell>
          <OpenButton projectId={projectId} objectId={object.id} />
        </TableCell>
      </TableRow>
    );
  }
  return (
    <TableRow>
      <TableCell className="font-medium">{objectIdLabel(object)}</TableCell>
      <TableCell>{String(object.metadata?.row_number || "-")}</TableCell>
      <TableCell>
        <StatusBadge status={object.status} />
      </TableCell>
      <TableCell className="max-w-72">{objectReasons(object).join(", ") || "-"}</TableCell>
      <TableCell>{metric?.recommendation || "-"}</TableCell>
      <TableCell>{recommendedAction(object, metric)}</TableCell>
      <TableCell>
        <RiskBadge risk={risk} />
      </TableCell>
      <TableCell>
        <OpenButton projectId={projectId} objectId={object.id} />
      </TableCell>
    </TableRow>
  );
}

function ImageHeader() {
  return (
    <TableRow>
      {["id", "preview/file_path", "label", "predicted_label", "confidence", "entropy", "quality_score", "duplicate_score", "label_error_probability", "status", "recommended_action", "action_risk", ""].map((item) => (
        <TableHead key={item}>{item}</TableHead>
      ))}
    </TableRow>
  );
}

function TabularHeader() {
  return (
    <TableRow>
      {["id", "row_index", "status", "reasons", "recommendation", "recommended_action", "action_risk", ""].map((item) => (
        <TableHead key={item}>{item}</TableHead>
      ))}
    </TableRow>
  );
}

function YoloHeader() {
  return (
    <TableRow>
      {["id", "file_path", "cv_quality_score", "status", "reasons", "recommendation", "recommended_action", "action_risk", ""].map((item) => (
        <TableHead key={item}>{item}</TableHead>
      ))}
    </TableRow>
  );
}

function Preview({ projectId, object }: { projectId: string; object: DataObject }) {
  return (
    <div className="h-14 w-16 overflow-hidden rounded-md border bg-muted">
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={apiPath(`/api/projects/${projectId}/objects/${object.id}/file`)} alt={object.file_path || object.id} className="h-full w-full object-cover" onError={(event) => (event.currentTarget.style.display = "none")} />
    </div>
  );
}

function OpenButton({ projectId, objectId }: { projectId: string; objectId: string }) {
  return (
    <Button asChild size="sm" variant="outline">
      <Link href={`/projects/${projectId}/objects/${objectId}`}>
        <Eye />
        Open
      </Link>
    </Button>
  );
}

function RiskBadge({ risk }: { risk: string }) {
  return <Badge variant={risk === "high" ? "danger" : risk === "medium" ? "warning" : "muted"}>{risk}</Badge>;
}

function FilterSelect({ value, onChange, label, options }: { value: string; onChange: (value: string) => void; label: string; options: string[] }) {
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

function compareObjects(a: DataObject, b: DataObject, sort: string) {
  if (sort === "status") return (a.status || "").localeCompare(b.status || "");
  if (sort === "quality") return (b.metrics?.quality_score || 0) - (a.metrics?.quality_score || 0);
  if (sort === "confidence") return (b.confidence || 0) - (a.confidence || 0);
  if (sort === "utility") return (b.metrics?.object_utility_score || 0) - (a.metrics?.object_utility_score || 0);
  return riskRank(actionRisk(b, b.metrics)) - riskRank(actionRisk(a, a.metrics));
}

function objectReasons(object: DataObject) {
  const metricReasons = object.metrics?.reasons;
  const validation = object.metadata?.validation_errors;
  if (Array.isArray(metricReasons) && metricReasons.length) return metricReasons;
  if (Array.isArray(validation)) return validation.map(String);
  return [];
}

function riskRank(risk: string) {
  return risk === "high" ? 3 : risk === "medium" ? 2 : 1;
}
