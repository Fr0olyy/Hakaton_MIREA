"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import JSZip from "jszip";
import { Archive, Download, FileJson, FileSpreadsheet } from "lucide-react";
import { createExport } from "@/lib/api";
import type { ExportArtifact } from "@/lib/types";
import { apiPath, downloadBlob } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { StatusBadge } from "@/components/status";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const files = [
  { label: "dataset_v2.csv", zipName: "dataset_v2.csv", downloadName: "dataset_v2.csv", icon: FileSpreadsheet },
  { label: "review_queue.csv", zipName: "review_queue.csv", downloadName: "review_queue.csv", icon: FileSpreadsheet },
  { label: "recommendations.json", zipName: "recommendations.json", downloadName: "recommendations.json", icon: FileJson },
  { label: "roadmap.json", zipName: "roadmap.json", downloadName: "roadmap.json", icon: FileJson },
  { label: "report.json", zipName: "dataset_report.json", downloadName: "report.json", icon: FileJson },
];

export default function ExportPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [artifact, setArtifact] = useState<ExportArtifact | null>(null);
  const [busy, setBusy] = useState("");
  const [error, setError] = useState("");

  async function ensureExport(showBusy = true) {
    if (artifact) return artifact;
    if (showBusy) setBusy("generate");
    setError("");
    try {
      const created = await createExport(projectId);
      setArtifact(created);
      return created;
    } catch (err) {
      setError((err as Error).message);
      throw err;
    } finally {
      if (showBusy) setBusy("");
    }
  }

  async function downloadZip() {
    setBusy("zip");
    setError("");
    try {
      const current = await ensureExport(false);
      const response = await fetch(apiPath(`/api/projects/${projectId}/exports/${current.id}/download`));
      if (!response.ok) throw new Error("export download failed");
      downloadBlob(await response.blob(), "dataforge_export.zip");
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setBusy("");
    }
  }

  async function downloadSingle(zipName: string, downloadName: string) {
    const current = await ensureExport(false);
    setBusy(zipName);
    setError("");
    try {
      const response = await fetch(apiPath(`/api/projects/${projectId}/exports/${current.id}/download`));
      if (!response.ok) throw new Error("export download failed");
      const zip = await JSZip.loadAsync(await response.blob());
      const file = zip.file(zipName);
      if (!file) throw new Error(`${zipName} is missing from export.zip`);
      downloadBlob(await file.async("blob"), downloadName);
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
        title="Export"
        description="Generate Dataset v2 and download the analysis artifacts."
        actions={
          <Button onClick={() => void ensureExport()} disabled={busy !== ""}>
            <Archive />
            {busy === "generate" ? "Generating" : "Generate Dataset v2"}
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[360px_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>Export status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">Artifact</span>
              <StatusBadge status={artifact?.status || "not_generated"} />
            </div>
            {artifact ? <div className="text-xs text-muted-foreground">Created {new Date(artifact.created_at).toLocaleString()}</div> : null}
            {error ? <div className="rounded-md border border-red-200 bg-red-50 p-3 text-red-800">{error}</div> : null}
            <Button variant="outline" onClick={downloadZip} disabled={busy !== ""}>
              <Download />
              {busy === "zip" ? "Preparing" : "Download export.zip"}
            </Button>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Artifact files</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-2 md:grid-cols-2">
            {files.map((file) => {
              const Icon = file.icon;
              return (
                <Button key={file.label} variant="outline" className="justify-start" onClick={() => downloadSingle(file.zipName, file.downloadName)} disabled={busy !== ""}>
                  <Icon />
                  {busy === file.zipName ? "Preparing" : file.label}
                </Button>
              );
            })}
          </CardContent>
        </Card>
      </div>
    </>
  );
}
