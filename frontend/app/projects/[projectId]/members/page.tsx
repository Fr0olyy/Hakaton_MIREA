"use client";

import { useParams } from "next/navigation";
import { roleOptions } from "@/lib/level2";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function MembersPage() {
  const { projectId } = useParams<{ projectId: string }>();
  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Members" description="Role-based Level 2 workspace view for the hackathon demo." />
      <Card>
        <CardHeader>
          <CardTitle>Project roles</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-3 md:grid-cols-2">
          {roleOptions.map((role) => (
            <div key={role.value} className="flex items-center justify-between rounded-md border bg-card p-3">
              <span className="font-medium">{role.label}</span>
              <Badge variant="muted">{role.value}</Badge>
            </div>
          ))}
        </CardContent>
      </Card>
    </>
  );
}
