"use client";

import { useEffect, useState } from "react";
import { Sparkles } from "lucide-react";
import { getAgentSummary } from "@/lib/api";
import type { AgentSummary } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export function AISummary({ projectId }: { projectId: string }) {
  const [data, setData] = useState<AgentSummary | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    getAgentSummary(projectId)
      .then(setData)
      .catch((err: Error) => setError(err.message));
  }, [projectId]);

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-accent" />
          AI Summary
        </CardTitle>
      </CardHeader>
      <CardContent>
        {data ? (
          <p className="text-sm leading-6 text-muted-foreground">{data.summary}</p>
        ) : (
          <p className="text-sm text-muted-foreground">{error || "Summary will appear after analysis."}</p>
        )}
      </CardContent>
    </Card>
  );
}
