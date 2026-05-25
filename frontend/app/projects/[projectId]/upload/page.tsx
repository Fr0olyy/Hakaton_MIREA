"use client";

import type { ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import JSZip from "jszip";
import { AlertTriangle, BarChart3, FileArchive, FileSpreadsheet, Play, UploadCloud } from "lucide-react";
import { getProject, runAnalysis, uploadDataset } from "@/lib/api";
import { isImageProject, isTabularProject, isYoloProject, modalityLabel } from "@/lib/level2";
import type { AnalysisJob, Project, UploadResult } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { StatusBadge } from "@/components/status";

export default function UploadDatasetPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [project, setProject] = useState<Project | null>(null);
  const [dataset, setDataset] = useState<File | null>(null);
  const [assets, setAssets] = useState<File | null>(null);
  const [uploadResult, setUploadResult] = useState<UploadResult | null>(null);
  const [analysisJob, setAnalysisJob] = useState<AnalysisJob | null>(null);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");
  const [warnings, setWarnings] = useState<string[]>([]);

  useEffect(() => {
    getProject(projectId).then(setProject).catch((err: Error) => setError(err.message));
  }, [projectId]);

  const uploadReady = useMemo(() => {
    if (!project) return false;
    if (isImageProject(project)) return Boolean(dataset && assets);
    if (isTabularProject(project)) return Boolean(dataset);
    if (isYoloProject(project)) return Boolean(dataset || assets);
    return Boolean(dataset);
  }, [project, dataset, assets]);

  async function onUpload() {
    if (!project) return;
    setBusy("upload");
    setError("");
    setWarnings([]);
    try {
      let datasetFile = dataset;
      if (isYoloProject(project) && !datasetFile && assets) {
        datasetFile = await buildYoloManifest(assets);
        setWarnings(["YOLO manifest was generated locally from label files because no manifest CSV was provided."]);
      }
      if (!datasetFile) throw new Error("dataset.csv or source file is required");
      const result = await uploadDataset(projectId, datasetFile, assets);
      setUploadResult(result);
      setAnalysisJob(null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function onAnalyze() {
    if (!project) return;
    setBusy("analysis");
    setError("");
    try {
      const job = await runAnalysis(projectId, project);
      setAnalysisJob(job);
      setWarnings(job.warnings || []);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy("");
    }
  }

  const files = analysisJob?.files || analysisJob?.output_files || {};
  const jobErrors = analysisJob?.errors || (analysisJob?.error_message ? [analysisJob.error_message] : []);

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        eyebrow={project?.name}
        title="Analyze Dataset"
        description={`Upload files for ${modalityLabel(project?.modality)} and run the analysis pipeline.`}
        actions={
          analysisJob?.status === "completed" ? (
            <Button asChild>
              <Link href={`/projects/${projectId}/dashboard`}>
                <BarChart3 />
                Open Dashboard
              </Link>
            </Button>
          ) : null
        }
      />

      <div className="grid gap-4 lg:grid-cols-[1fr_380px]">
        <Card>
          <CardHeader>
            <div className="flex items-center justify-between gap-3">
              <CardTitle>Dataset files</CardTitle>
              <Badge variant="muted">{modalityLabel(project?.modality)}</Badge>
            </div>
          </CardHeader>
          <CardContent className="space-y-5">
            {project && isImageProject(project) ? <ImageUploadForm onDataset={setDataset} onAssets={setAssets} /> : null}
            {project && isTabularProject(project) ? <TabularUploadForm onDataset={setDataset} /> : null}
            {project && isYoloProject(project) ? <YoloUploadForm onDataset={setDataset} onAssets={setAssets} /> : null}

            {warnings.length ? <MessageList tone="warning" title="Warnings" items={warnings} /> : null}
            {jobErrors.length ? <MessageList tone="danger" title="Errors" items={jobErrors} /> : null}
            {error ? <MessageList tone="danger" title="Request failed" items={[error]} /> : null}

            <div className="flex flex-wrap gap-2">
              <Button onClick={onUpload} disabled={!uploadReady || busy !== ""}>
                <UploadCloud />
                {busy === "upload" ? "Uploading" : "Upload Dataset"}
              </Button>
              <Button variant="secondary" onClick={onAnalyze} disabled={busy !== ""}>
                <Play />
                {busy === "analysis" ? "Running" : "Run Analysis"}
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Analysis status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <StatusLine label="Modality" value={modalityLabel(project?.modality)} />
            <StatusLine label="Task type" value={project?.task_type || "-"} />
            <StatusLine label="Upload" badge={uploadResult?.dataset_version.status || "pending"} />
            <StatusLine label="Objects" value={String(uploadResult?.objects_count ?? 0)} />
            <StatusLine label="Invalid rows" value={String(uploadResult?.invalid_objects ?? 0)} />
            <StatusLine label="Analysis" badge={analysisJob?.status || (busy === "analysis" ? "running" : "not_started")} />
            <StatusLine label="Progress" value={`${analysisJob?.progress_percent ?? (analysisJob?.status === "completed" ? 100 : 0)}%`} />
            {Object.keys(files).length ? (
              <div className="rounded-md border bg-slate-50 p-3">
                <div className="mb-2 font-medium">Files from analysis</div>
                <div className="space-y-1 text-xs text-muted-foreground">
                  {Object.entries(files).map(([name, path]) => (
                    <div key={name} className="truncate" title={path}>
                      {name}: {path}
                    </div>
                  ))}
                </div>
              </div>
            ) : null}
          </CardContent>
        </Card>
      </div>
    </>
  );
}

function ImageUploadForm({ onDataset, onAssets }: { onDataset: (file: File | null) => void; onAssets: (file: File | null) => void }) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <FileInput icon={<FileSpreadsheet className="h-4 w-4" />} id="dataset" label="dataset.csv" accept=".csv,text/csv" onChange={onDataset} />
      <FileInput icon={<FileArchive className="h-4 w-4" />} id="images" label="images.zip" accept=".zip,application/zip" onChange={onAssets} />
    </div>
  );
}

function TabularUploadForm({ onDataset }: { onDataset: (file: File | null) => void }) {
  return <FileInput icon={<FileSpreadsheet className="h-4 w-4" />} id="tabular" label="tabular.csv" accept=".csv,text/csv" onChange={onDataset} />;
}

function YoloUploadForm({ onDataset, onAssets }: { onDataset: (file: File | null) => void; onAssets: (file: File | null) => void }) {
  return (
    <div className="grid gap-4 md:grid-cols-2">
      <FileInput icon={<FileSpreadsheet className="h-4 w-4" />} id="manifest" label="manifest.csv (optional)" accept=".csv,text/csv" onChange={onDataset} />
      <FileInput icon={<FileArchive className="h-4 w-4" />} id="yolo" label="YOLO dataset / labels.zip" accept=".zip,application/zip" onChange={onAssets} />
    </div>
  );
}

function FileInput({ id, label, accept, icon, onChange }: { id: string; label: string; accept: string; icon: ReactNode; onChange: (file: File | null) => void }) {
  return (
    <div className="space-y-2 rounded-md border bg-slate-50 p-4">
      <Label htmlFor={id} className="flex items-center gap-2">
        {icon}
        {label}
      </Label>
      <Input id={id} type="file" accept={accept} onChange={(event) => onChange(event.target.files?.[0] || null)} />
    </div>
  );
}

function StatusLine({ label, value, badge }: { label: string; value?: string; badge?: string }) {
  return (
    <div className="flex items-center justify-between gap-3">
      <span className="text-muted-foreground">{label}</span>
      {badge ? <StatusBadge status={badge} /> : <span className="text-right font-medium">{value}</span>}
    </div>
  );
}

function MessageList({ tone, title, items }: { tone: "warning" | "danger"; title: string; items: string[] }) {
  const classes = tone === "warning" ? "border-amber-200 bg-amber-50 text-amber-900" : "border-red-200 bg-red-50 text-red-900";
  return (
    <div className={`rounded-md border p-3 text-sm ${classes}`}>
      <div className="mb-1 flex items-center gap-2 font-medium">
        <AlertTriangle className="h-4 w-4" />
        {title}
      </div>
      <ul className="list-inside list-disc space-y-1">
        {items.map((item) => (
          <li key={item}>{item}</li>
        ))}
      </ul>
    </div>
  );
}

async function buildYoloManifest(archive: File) {
  const zip = await JSZip.loadAsync(archive);
  const rows = ["id,file_path,label"];
  let index = 0;
  for (const [path, entry] of Object.entries(zip.files)) {
    if (entry.dir || !path.toLowerCase().endsWith(".txt")) continue;
    const content = (await entry.async("string")).trim();
    const firstClass = content.split(/\s+/)[0] || "unlabeled";
    rows.push(`${csvEscape(`label_${index + 1}`)},${csvEscape(path)},${csvEscape(firstClass)}`);
    index++;
  }
  if (index === 0) {
    rows.push("empty_yolo_dataset,,unlabeled");
  }
  return new File([rows.join("\n")], "dataset.csv", { type: "text/csv" });
}

function csvEscape(value: string) {
  if (/[",\n]/.test(value)) return `"${value.replace(/"/g, '""')}"`;
  return value;
}
