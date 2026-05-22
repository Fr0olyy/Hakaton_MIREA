"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowRight, BarChart3, Plus, RefreshCw, Upload } from "lucide-react";
import { getDashboard, listProjects } from "@/lib/api";
import type { Dashboard, Project } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [snapshots, setSnapshots] = useState<Record<string, Dashboard | null>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  function load() {
    setLoading(true);
    setError("");
    listProjects()
      .then(async (items) => {
        setProjects(items);
        const loaded = await Promise.all(
          items.map(async (project) => {
            try {
              return [project.id, await getDashboard(project.id)] as const;
            } catch {
              return [project.id, null] as const;
            }
          }),
        );
        setSnapshots(Object.fromEntries(loaded));
      })
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  return (
    <>
      <PageHeader
        eyebrow="Level 1"
        title="Projects"
        description="Create an image classification dataset project, upload CSV/images, run ML analysis, and export a curated dataset."
        actions={
          <>
            <Button variant="outline" onClick={load}>
              <RefreshCw />
              Refresh
            </Button>
            <Button asChild>
              <Link href="/projects/new">
                <Plus />
                Create Project
              </Link>
            </Button>
          </>
        }
      />

      {loading ? <LoadingState label="Loading projects" /> : null}
      {error ? <ErrorState message={error} /> : null}

      {!loading && !error && (
        <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          {projects.map((project) => (
            <Card key={project.id} className="overflow-hidden">
              <CardHeader>
                <div className="flex items-start justify-between gap-3">
                  <CardTitle>{project.name}</CardTitle>
                  <Badge variant="success">{project.modality}</Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex flex-wrap gap-2">
                  <Badge variant="muted">{project.task_type}</Badge>
                  {(project.classes || []).slice(0, 5).map((label) => (
                    <Badge key={label} variant="outline">
                      {label}
                    </Badge>
                  ))}
                </div>
                <ProjectSnapshot snapshot={snapshots[project.id]} />
                <div className="text-xs text-muted-foreground">Created {new Date(project.created_at).toLocaleString()}</div>
                <div className="flex flex-wrap gap-2">
                  <Button asChild size="sm">
                    <Link href={`/projects/${project.id}/upload`}>
                      <Upload />
                      Upload
                    </Link>
                  </Button>
                  <Button asChild size="sm" variant="outline">
                    <Link href={`/projects/${project.id}/dashboard`}>
                      <BarChart3 />
                      Dashboard
                    </Link>
                  </Button>
                  <Button asChild size="sm" variant="ghost">
                    <Link href={`/projects/${project.id}/queue`}>
                      Queue
                      <ArrowRight />
                    </Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}

          {!projects.length ? (
            <Card className="md:col-span-2 xl:col-span-3">
              <CardContent className="flex flex-col items-center justify-center gap-3 p-10 text-center">
                <div className="text-base font-medium">No projects yet</div>
                <p className="max-w-md text-sm text-muted-foreground">Start with a Level 1 image classification project and upload dataset.csv plus images.zip.</p>
                <Button asChild>
                  <Link href="/projects/new">
                    <Plus />
                    Create Project
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ) : null}
        </div>
      )}
    </>
  );
}

function ProjectSnapshot({ snapshot }: { snapshot: Dashboard | null | undefined }) {
  if (snapshot === undefined) {
    return <div className="rounded-md bg-slate-50 p-3 text-sm text-muted-foreground">Checking dataset state...</div>;
  }
  if (snapshot === null) {
    return <div className="rounded-md border border-dashed p-3 text-sm text-muted-foreground">No analyzed dataset yet</div>;
  }
  return (
    <div className="grid grid-cols-3 gap-2 rounded-md bg-slate-50 p-3 text-sm">
      <SnapshotValue label="objects" value={String(snapshot.dataset_version.objects_count || snapshot.objects_count)} />
      <SnapshotValue label="review" value={String(snapshot.review_items)} />
      <SnapshotValue label="readiness" value={snapshot.readiness_score.toFixed(1)} />
    </div>
  );
}

function SnapshotValue({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="font-semibold text-slate-950">{value}</div>
      <div className="text-[11px] uppercase tracking-wide text-muted-foreground">{label}</div>
    </div>
  );
}
