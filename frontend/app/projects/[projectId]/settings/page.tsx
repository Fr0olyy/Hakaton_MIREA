"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getProject } from "@/lib/api";
import { modalityLabel } from "@/lib/level2";
import type { Project } from "@/lib/types";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function SettingsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [project, setProject] = useState<Project | null>(null);

  useEffect(() => {
    getProject(projectId).then(setProject).catch(() => undefined);
  }, [projectId]);

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Settings" description="Read-only project settings for the Level 2 demo." />
      <Card>
        <CardHeader>
          <CardTitle>{project?.name || "Project"}</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 text-sm md:grid-cols-3">
          <Info label="modality" value={modalityLabel(project?.modality)} />
          <Info label="task_type" value={project?.task_type || "-"} />
          <Info label="classes" value={(project?.classes || []).join(", ") || "-"} />
        </CardContent>
      </Card>
    </>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border bg-card p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 break-words font-medium">{value}</div>
    </div>
  );
}
