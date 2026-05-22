export type Project = {
  id: string;
  name: string;
  modality: string;
  task_type: string;
  classes?: string[];
  created_at: string;
  updated_at: string;
};

export type DatasetVersion = {
  id: string;
  project_id: string;
  version_name: string;
  status: string;
  objects_count: number;
  readiness_score: number;
  created_at: string;
};

export type DataObject = {
  id: string;
  dataset_version_id: string;
  external_id?: string;
  file_path?: string;
  text_content?: string;
  label?: string;
  predicted_label?: string;
  confidence?: number;
  split?: string;
  source?: string;
  annotator?: string;
  metadata?: Record<string, unknown>;
  status: string;
  created_at: string;
  metrics?: ObjectMetric;
};

export type ObjectMetric = {
  id: string;
  object_id: string;
  entropy: number;
  uncertainty_score: number;
  label_error_probability: number;
  duplicate_score: number;
  rarity_score: number;
  class_deficit_score: number;
  quality_score: number;
  novelty_score: number;
  object_utility_score: number;
  final_score: number;
  reasons?: string[];
  recommendation?: string;
  probabilities?: Record<string, unknown>;
  created_at: string;
};

export type Dashboard = {
  project: Project;
  dataset_version: DatasetVersion;
  objects_count: number;
  readiness_score: number;
  status_counts: Record<string, number>;
  label_counts: Record<string, number>;
  class_distribution: Record<string, number>;
  imbalance_index: number;
  avg_entropy: number;
  avg_label_error_probability: number;
  missing_files_count: number;
  review_items: number;
  analysis_source: string;
};

export type ProbabilisticAnalysis = {
  dataset_version: DatasetVersion;
  metrics: ObjectMetric[];
  class_distribution: Record<string, number>;
  imbalance_index: number;
  avg_entropy: number;
  avg_label_error_probability: number;
  missing_files_count: number;
  review_items: number;
  analysis_source: string;
};

export type ReviewQueueItem = {
  object: DataObject;
  metric: ObjectMetric;
};

export type Recommendation = {
  id: string;
  project_id: string;
  dataset_version_id: string;
  type: string;
  priority: string;
  title: string;
  description: string;
  affected_objects_count: number;
  object_ids?: string[];
  created_at: string;
};

export type RoadmapItem = {
  id: string;
  project_id: string;
  dataset_version_id: string;
  priority: number;
  title: string;
  description: string;
  action_type: string;
  expected_impact: string;
  created_at: string;
};

export type AnalysisJob = {
  id: string;
  project_id: string;
  dataset_version_id: string;
  status: string;
  error_message?: string;
  started_at: string;
  finished_at?: string;
};

export type UploadResult = {
  dataset_version: DatasetVersion;
  objects_count: number;
  invalid_objects: number;
};

export type ExportArtifact = {
  id: string;
  project_id: string;
  dataset_version_id: string;
  file_path: string;
  status: string;
  created_at: string;
};

export type AgentSummary = {
  summary: string;
  dashboard: Dashboard;
  recommendations: Recommendation[];
  roadmap: RoadmapItem[];
};
