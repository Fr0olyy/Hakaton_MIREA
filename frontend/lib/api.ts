import { apiPath } from "@/lib/utils";
import type {
  AgentSummary,
  AnalysisJob,
  Dashboard,
  ExportArtifact,
  ProbabilisticAnalysis,
  Project,
  Recommendation,
  ReviewQueueItem,
  RoadmapItem,
  UploadResult,
} from "@/lib/types";

async function parseError(response: Response) {
  try {
    const data = await response.json();
    return data.error || data.message || response.statusText;
  } catch {
    return response.statusText;
  }
}

export async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(apiPath(path), {
    ...init,
    cache: "no-store",
    headers: {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...init?.headers,
    },
  });
  if (!response.ok) {
    throw new Error(await parseError(response));
  }
  return response.json() as Promise<T>;
}

export function listProjects() {
  return requestJSON<Project[]>("/api/projects");
}

export function getProject(projectId: string) {
  return requestJSON<Project>(`/api/projects/${projectId}`);
}

export function createProject(payload: { name: string; modality: string; task_type: string; classes: string[] }) {
  return requestJSON<Project>("/api/projects", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function uploadDataset(projectId: string, dataset: File, images?: File | null) {
  const formData = new FormData();
  formData.append("dataset", dataset);
  if (images) formData.append("images", images);
  return requestJSON<UploadResult>(`/api/projects/${projectId}/upload`, {
    method: "POST",
    body: formData,
  });
}

export function runAnalysis(projectId: string) {
  return requestJSON<AnalysisJob>(`/api/projects/${projectId}/analyze`, { method: "POST" });
}

export function getDashboard(projectId: string) {
  return requestJSON<Dashboard>(`/api/projects/${projectId}/dashboard`);
}

export function getProbabilisticAnalysis(projectId: string) {
  return requestJSON<ProbabilisticAnalysis>(`/api/projects/${projectId}/probabilistic-analysis`);
}

export function getReviewQueue(projectId: string) {
  return requestJSON<ReviewQueueItem[]>(`/api/projects/${projectId}/review-queue`);
}

export function getRecommendations(projectId: string) {
  return requestJSON<Recommendation[]>(`/api/projects/${projectId}/recommendations`);
}

export function getRoadmap(projectId: string) {
  return requestJSON<RoadmapItem[]>(`/api/projects/${projectId}/roadmap`);
}

export function createExport(projectId: string) {
  return requestJSON<ExportArtifact>(`/api/projects/${projectId}/export`, { method: "POST" });
}

export function getAgentSummary(projectId: string) {
  return requestJSON<AgentSummary>(`/api/projects/${projectId}/agent/summary`, { method: "POST" });
}
