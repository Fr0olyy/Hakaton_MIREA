"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { BarChart3, Play, UploadCloud } from "lucide-react";
import { getProject, runAnalysis, uploadDataset } from "@/lib/api";
import type { AnalysisJob, Project, UploadResult } from "@/lib/types";
import { Button } from "@/components/ui/button";
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
  const [images, setImages] = useState<File | null>(null);
  const [uploadResult, setUploadResult] = useState<UploadResult | null>(null);
  const [analysisJob, setAnalysisJob] = useState<AnalysisJob | null>(null);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    getProject(projectId).then(setProject).catch((err: Error) => setError(err.message));
  }, [projectId]);

  async function onUpload() {
    if (!dataset) return;
    setBusy("upload");
    setError("");
    try {
      const result = await uploadDataset(projectId, dataset, images);
      setUploadResult(result);
      setAnalysisJob(null);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function onAnalyze() {
    setBusy("analysis");
    setError("");
    try {
      const job = await runAnalysis(projectId);
      setAnalysisJob(job);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy("");
    }
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        eyebrow={project?.name}
        title="Upload Dataset"
        description="Upload dataset.csv and images.zip, then run the Level 1 ML analysis pipeline."
        actions={
          analysisJob?.status === "done" ? (
            <Button asChild>
              <Link href={`/projects/${projectId}/dashboard`}>
                <BarChart3 />
                Open Dashboard
              </Link>
            </Button>
          ) : null
        }
      />

      <div className="grid gap-4 lg:grid-cols-[1fr_360px]">
        <Card>
          <CardHeader>
            <CardTitle>Dataset files</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="dataset">dataset.csv</Label>
              <Input id="dataset" type="file" accept=".csv,text/csv" onChange={(event) => setDataset(event.target.files?.[0] || null)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="images">images.zip</Label>
              <Input id="images" type="file" accept=".zip,application/zip" onChange={(event) => setImages(event.target.files?.[0] || null)} />
            </div>
            {error ? <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div> : null}
            <div className="flex flex-wrap gap-2">
              <Button onClick={onUpload} disabled={!dataset || busy !== ""}>
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
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Upload</span>
              {uploadResult ? <StatusBadge status={uploadResult.dataset_version.status} /> : <StatusBadge status="pending" />}
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Objects</span>
              <span className="font-medium">{uploadResult?.objects_count ?? 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Invalid rows</span>
              <span className="font-medium">{uploadResult?.invalid_objects ?? 0}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Analysis</span>
              <StatusBadge status={analysisJob?.status || (busy === "analysis" ? "running" : "not_started")} />
            </div>
          </CardContent>
        </Card>
      </div>
    </>
  );
}
