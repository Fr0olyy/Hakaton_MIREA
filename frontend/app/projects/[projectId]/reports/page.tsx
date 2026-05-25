"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getDashboard, getRecommendations, getRoadmap } from "@/lib/api";
import type { Dashboard, Recommendation, RoadmapItem } from "@/lib/types";
import { formatScore } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { LoadingState } from "@/components/state";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function ReportsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [roadmap, setRoadmap] = useState<RoadmapItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getDashboard(projectId), getRecommendations(projectId), getRoadmap(projectId)])
      .then(([dash, recs, road]) => {
        setDashboard(dash);
        setRecommendations(recs || []);
        setRoadmap(road || []);
      })
      .finally(() => setLoading(false));
  }, [projectId]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading reports" />
      </>
    );
  }

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Reports" description="Analyst view of dataset status, recommendations, roadmap, and version comparison." />
      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Summary</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <Info label="objects" value={String(dashboard?.objects_count || 0)} />
            <Info label="readiness" value={formatScore(dashboard?.readiness_score, 1)} />
            <Info label="review_queue" value={String(dashboard?.review_items || 0)} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Recommendations</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">{recommendations.slice(0, 5).map((item) => <div key={item.id}>{item.title}</div>)}</CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Version / Strategy Comparison</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">{roadmap.slice(0, 5).map((item) => <div key={item.id}>{item.title}</div>)}</CardContent>
        </Card>
      </div>
    </>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md bg-slate-50 p-2">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  );
}
