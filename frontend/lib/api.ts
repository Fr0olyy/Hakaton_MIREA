import { apiPath } from "@/lib/utils";
import type {
  AgentSummary,
  AnalysisJob,
  AgentResponse,
  ClassActionPlanItem,
  CollectionTask,
  Dashboard,
  DemoDataset,
  DemoProjectResult,
  ExportArtifact,
  ListObjectsResult,
  ObjectComment,
  ObjectDetail,
  ProbabilisticAnalysis,
  Project,
  Recommendation,
  ReviewQueueItem,
  RoadmapItem,
  SyntheticTask,
  UploadResult,
} from "@/lib/types";

async function parseError(response: Response) {
  if (response.status === 404) {
    return "not found";
  }

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

  const data = await response.json();

  return data as T;
}

export function listProjects() {
  return requestJSON<Project[]>("/api/projects");
}

export function listDemoDatasets() {
  return requestJSON<DemoDataset[]>("/api/demo-datasets");
}

export function createDemoProject(modality: string) {
  return requestJSON<DemoProjectResult>(`/api/demo-datasets/${modality}/project`, { method: "POST" });
}

export async function getProject(projectId: string) {
  try {
    const project = await requestJSON<Project>(`/api/projects/${projectId}`);
    return normalizeProject(project);
  } catch (error) {
    const projects = await listProjects().catch(() => []);
    const project = projects.find((item) => item.id === projectId);

    if (project) {
      return normalizeProject(project);
    }

    throw error;
  }
}

function normalizeProject(project: Project): Project {
  return {
    ...project,
    classes: Array.isArray(project.classes) ? project.classes : [],
  };
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

export function runAnalysis(projectId: string, project?: Project) {
  const payload = project
    ? {
        modality: project.modality,
        task_type: project.task_type,
        dataset_path: "storage/projects/" + project.id + "/raw/dataset.csv",
        data_dir: "storage/projects/" + project.id + "/raw/images",
        output_dir: "storage/projects/" + project.id + "/analysis/latest",
      }
    : undefined;
  return requestJSON<AnalysisJob>(`/api/projects/${projectId}/analyze`, {
    method: "POST",
    body: payload ? JSON.stringify(payload) : undefined,
  });
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

export function listObjects(projectId: string, query: Record<string, string | number | undefined> = {}) {
  const params = new URLSearchParams();
  Object.entries(query).forEach(([key, value]) => {
    if (value !== undefined && value !== "" && value !== "all") params.set(key, String(value));
  });
  const suffix = params.toString() ? `?${params}` : "";
  return requestJSON<ListObjectsResult>(`/api/projects/${projectId}/objects${suffix}`);
}

export async function listAllObjects(projectId: string, query: Record<string, string | number | undefined> = {}) {
  const first = await listObjects(projectId, { ...query, page: 1, per_page: 100 });
  if (first.total_pages <= 1) return first;
  const pages = await Promise.all(
    Array.from({ length: first.total_pages - 1 }, (_, index) => listObjects(projectId, { ...query, page: index + 2, per_page: 100 })),
  );
  return {
    ...first,
    objects: [first.objects, ...pages.map((page) => page.objects)].flat(),
  };
}

export function getObject(projectId: string, objectId: string) {
  return requestJSON<ObjectDetail>(`/api/projects/${projectId}/objects/${objectId}`);
}

export function createObjectAction(projectId: string, objectId: string, payload: { action: string; old_value?: string; new_value?: string; comment?: string }) {
  return requestJSON(`/api/projects/${projectId}/objects/${objectId}/action`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export function addObjectComment(projectId: string, objectId: string, text: string) {
  return requestJSON<ObjectComment>(`/api/projects/${projectId}/objects/${objectId}/comments`, {
    method: "POST",
    body: JSON.stringify({ text }),
  });
}

export function getClassActionPlan(projectId: string) {
  return requestJSON<ClassActionPlanItem[]>(`/api/projects/${projectId}/class-action-plan`);
}

export function getCollectionTasks(projectId: string) {
  return requestJSON<CollectionTask[]>(`/api/projects/${projectId}/collection-tasks`);
}

export function updateCollectionTask(projectId: string, taskId: string, updates: Record<string, unknown>) {
  return requestJSON<CollectionTask>(`/api/projects/${projectId}/collection-tasks/${taskId}`, {
    method: "PATCH",
    body: JSON.stringify(updates),
  });
}

export function getSyntheticTasks(projectId: string) {
  return requestJSON<SyntheticTask[]>(`/api/projects/${projectId}/synthetic-tasks`);
}

export function updateSyntheticTask(projectId: string, taskId: string, updates: Record<string, unknown>) {
  return requestJSON<SyntheticTask>(`/api/projects/${projectId}/synthetic-tasks/${taskId}`, {
    method: "PATCH",
    body: JSON.stringify(updates),
  });
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

export function postAgentEndpoint(projectId: string, endpoint: "summary" | "collection-plan" | "synthetic-plan" | "dataset-summary") {
  return requestJSON<AgentResponse>(`/api/projects/${projectId}/agent/${endpoint}`, { method: "POST" });
}

export function safeArray<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}