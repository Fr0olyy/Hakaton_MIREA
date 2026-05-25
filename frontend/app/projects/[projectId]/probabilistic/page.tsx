"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { getProbabilisticAnalysis } from "@/lib/api";
import type { ProbabilisticAnalysis } from "@/lib/types";
import { maxProbability } from "@/lib/utils";
import { ProjectNav } from "@/components/project-nav";
import { PageHeader } from "@/components/page-header";
import { LoadingState, ErrorState } from "@/components/state";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { BarChart, DistributionSummary, Histogram } from "@/components/charts";

export default function ProbabilisticPage() {
  const { projectId } = useParams<{ projectId: string }>();
  const [data, setData] = useState<ProbabilisticAnalysis | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getProbabilisticAnalysis(projectId)
      .then(setData)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [projectId]);

  if (loading) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <LoadingState label="Loading probabilistic analysis" />
      </>
    );
  }

  if (error || !data) {
    return (
      <>
        <ProjectNav projectId={projectId} />
        <ErrorState message={error || "Probabilistic analysis is not available."} />
      </>
    );
  }

  const entropy = data.metrics.map((metric) => metric.entropy);
  const confidence = data.metrics.map((metric) => maxProbability(metric.probabilities));
  const uncertainty = data.metrics.map((metric) => metric.uncertainty_score);
  const labelRisk = data.metrics.map((metric) => metric.label_error_probability);
  const classDeficit = Object.fromEntries(
    data.metrics
      .slice()
      .sort((a, b) => b.class_deficit_score - a.class_deficit_score)
      .slice(0, 15)
      .map((metric, index) => [`object ${index + 1}`, metric.class_deficit_score]),
  );

  return (
    <>
      <ProjectNav projectId={projectId} />
      <PageHeader
        title="Probabilistic Analysis"
        description={`Metrics count: ${data.metrics.length}. Entropy, confidence, uncertainty, label risk, and class deficit are normalized for Level 2 review.`}
      />

      <div className="grid gap-4 lg:grid-cols-2">
        <MetricDistribution title="entropy distribution" values={entropy} />
        <MetricDistribution title="confidence distribution" values={confidence} />
        <MetricDistribution title="uncertainty distribution" values={uncertainty} />
        <MetricDistribution title="label_error_probability distribution" values={labelRisk} />
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>class_deficit chart</CardTitle>
          </CardHeader>
          <CardContent>
            <BarChart data={classDeficit} valueKind="score" maxItems={15} />
          </CardContent>
        </Card>
      </div>
    </>
  );
}

function MetricDistribution({ title, values }: { title: string; values: number[] }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <Histogram values={values} />
        <DistributionSummary values={values} />
      </CardContent>
    </Card>
  );
}
