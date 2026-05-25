"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getRecommendations } from "@/lib/api";
import type { Recommendation } from "@/lib/types";
import { formatNumber } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { ErrorState, LoadingState } from "@/components/state";
import { AISummary } from "@/components/ai-summary";
import { PriorityBadge } from "@/components/status";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";

export default function RecommendationsPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [items, setItems] = useState<Recommendation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getRecommendations(projectId)
      .then(setItems)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader title="Recommendations" description="Actionable fixes and curation guidance generated from analysis." />
      {loading ? <LoadingState label="Loading recommendations" /> : null}
      {error ? <ErrorState message={error} /> : null}

      {!loading && !error ? (
        <div className="grid gap-4 lg:grid-cols-[1fr_380px]">
          <div className="space-y-3">
            {items.map((item) => (
              <Card key={item.id}>
                <CardHeader>
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <CardTitle>{item.title}</CardTitle>
                    <div className="flex gap-2">
                      <PriorityBadge priority={item.priority} />
                      <Badge variant="outline">{item.type}</Badge>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="space-y-3">
                  <p className="text-sm leading-6 text-muted-foreground">{item.description}</p>
                  <div className="text-sm">
                    <span className="font-medium">{formatNumber(item.affected_objects_count)}</span>{" "}
                    <span className="text-muted-foreground">affected objects</span>
                  </div>
                </CardContent>
              </Card>
            ))}
            {!items.length ? <Empty label="No recommendations yet. Run analysis first." /> : null}
          </div>
          <AISummary projectId={projectId} />
        </div>
      ) : null}
    </>
  );
}

function Empty({ label }: { label: string }) {
  return <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">{label}</div>;
}
