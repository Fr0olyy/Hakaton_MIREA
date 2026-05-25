"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { Database, FileSpreadsheet, Save, ScanSearch } from "lucide-react";
import { createProject } from "@/lib/api";
import { modalityOptions, taskTypeForModality } from "@/lib/level2";
import type { Modality } from "@/lib/types";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { PageHeader } from "@/components/page-header";

export default function CreateProjectPage() {
  const router = useRouter();
  const [name, setName] = useState("DataForge demo");
  const [modality, setModality] = useState<Modality>("image_classification");
  const [classes, setClasses] = useState("butterfly\ncat\nchicken\ncow\ndog\nelephant\nhorse\nsheep\nspider\nsquirrel");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  async function onSubmit(event: FormEvent) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    const parsedClasses = classes
      .split(/[\n,;]/)
      .map((item) => item.trim())
      .filter(Boolean);
    try {
      const project = await createProject({
        name,
        modality,
        task_type: taskTypeForModality(modality),
        classes: modality === "image_classification" ? parsedClasses : [],
      });
      router.push(`/projects/${project.id}/upload`);
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <PageHeader title="Create Project" description="Choose the dataset modality. The UI adapts upload, analysis, dashboards, queues, and action plans to that modality." />
      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>Project settings</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(event) => setName(event.target.value)} required />
            </div>
            <div className="grid gap-4 md:grid-cols-[1fr_220px]">
              <div className="space-y-2">
                <Label>Modality</Label>
                <Select value={modality} onChange={(event) => setModality(event.target.value as Modality)}>
                  {modalityOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Task type</Label>
                <Input value={taskTypeForModality(modality)} readOnly />
              </div>
            </div>
            <div className="grid gap-3 md:grid-cols-3">
              {modalityOptions.map((option) => {
                const Icon = option.value === "tabular_classification" ? FileSpreadsheet : option.value === "image_detection_yolo" ? ScanSearch : Database;
                return (
                  <button
                    key={option.value}
                    type="button"
                    onClick={() => setModality(option.value)}
                    className={`rounded-md border p-4 text-left transition hover:border-primary ${modality === option.value ? "border-primary bg-primary/5" : "bg-card"}`}
                  >
                    <div className="flex items-center justify-between gap-2">
                      <Icon className="h-5 w-5 text-primary" />
                      {modality === option.value ? <Badge variant="success">selected</Badge> : null}
                    </div>
                    <div className="mt-3 font-medium">{option.label}</div>
                    <div className="mt-1 text-sm text-muted-foreground">{option.description}</div>
                  </button>
                );
              })}
            </div>
            {modality === "image_classification" ? (
              <div className="space-y-2">
                <Label htmlFor="classes">Classes</Label>
                <Textarea id="classes" value={classes} onChange={(event) => setClasses(event.target.value)} placeholder="cat&#10;dog&#10;butterfly" />
              </div>
            ) : null}
            {error ? <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">{error}</div> : null}
            <Button type="submit" disabled={submitting || !name.trim()}>
              <Save />
              {submitting ? "Creating" : "Create Project"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </>
  );
}
