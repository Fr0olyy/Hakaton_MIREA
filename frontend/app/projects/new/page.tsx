"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { Save } from "lucide-react";
import { createProject } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { PageHeader } from "@/components/page-header";

export default function CreateProjectPage() {
  const router = useRouter();
  const [name, setName] = useState("DataForge demo");
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
        modality: "image",
        task_type: "classification",
        classes: parsedClasses,
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
      <PageHeader title="Create Project" description="Level 1 supports image classification datasets." />
      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>Project settings</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(event) => setName(event.target.value)} required />
            </div>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="space-y-2">
                <Label>Modality</Label>
                <Input value="image" readOnly />
              </div>
              <div className="space-y-2">
                <Label>Task type</Label>
                <Input value="classification" readOnly />
              </div>
            </div>
            <div className="space-y-2">
              <Label htmlFor="classes">Classes</Label>
              <Textarea id="classes" value={classes} onChange={(event) => setClasses(event.target.value)} placeholder="cat&#10;dog&#10;butterfly" />
            </div>
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
