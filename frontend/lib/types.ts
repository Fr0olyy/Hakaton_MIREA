export type Project = {
  id: string;
  name: string;
  modality: Modality | string;
  task_type: string;
  classes?: string[];
  created_at: string;
  updated_at: string;
};

export type Modality = "image" | "image_classification" | "tabular_classification" | "image_detection_yolo";

export type ProjectRole = "admin" | "ml_engineer" | "annotator" | "domain_expert" | "data_analyst";

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
  progress_percent?: number;
  progress_stage?: string;
  output_files?: Record<string, string>;
  files?: Record<string, string>;
  warnings?: string[];
  errors?: string[];
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
  dashboard?: Dashboard;
  recommendations?: Recommendation[];
  roadmap?: RoadmapItem[];
  key_findings?: string[];
  risks?: string[];
  recommended_next_steps?: string[];
  used_context_fields?: string[];
};

export type ListObjectsResult = {
  objects: DataObject[];
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
};

export type ObjectAction = {
  id: string;
  project_id: string;
  object_id: string;
  user_id: string;
  action: string;
  old_value?: string;
  new_value?: string;
  comment?: string;
  created_at: string;
};

export type ObjectComment = {
  id: string;
  project_id: string;
  object_id: string;
  user_id: string;
  text: string;
  created_at: string;
};

export type ObjectDetail = DataObject & {
  actions?: ObjectAction[];
  comments?: ObjectComment[];
};

export type CollectionTask = {
  id: string;
  project_id: string;
  target_class?: string;
  target?: string;
  task_type?: string;
  target_count: number;
  priority: string;
  risk: string;
  status: string;
  reason?: string;
  expected_impact?: string;
  owner_role?: string;
  created_by?: string;
  created_at: string;
};

export type SyntheticTask = {
  id: string;
  project_id: string;
  target_class?: string;
  target?: string;
  task_type?: string;
  target_count: number;
  prompt?: string;
  negative_prompt?: string;
  expected_impact?: string;
  risk: string;
  synthetic_bias_risk?: string;
  requires_human_validation?: boolean;
  owner_role?: string;
  priority?: string;
  status: string;
  created_by?: string;
  created_at: string;
};

export type ClassActionPlanItem = {
  class?: string;
  target_class?: string;
  problem?: string;
  issue?: string;
  objects_count?: number;
  target_count?: number;
  class_deficit_score?: number;
  review_objects_count?: number;
  label_error_count?: number;
  hard_examples_count?: number;
  missing_values?: number;
  numeric_outliers?: number;
  high_cardinality_columns?: number;
  invalid_bboxes?: number;
  tiny_boxes?: number;
  affected_files_count?: number;
  recommended_action?: string;
  priority?: string | number;
  risk?: string;
  reason?: string;
  expected_impact?: string;
  real_collection_priority?: number;
  synthetic_data_candidate_score?: number;
};

export type AgentResponse = {
  summary: string;
  key_findings?: string[];
  risks?: string[];
  recommended_next_steps?: string[];
  used_context_fields?: string[];
};

export type DemoDataset = {
  modality: Modality | string;
  label: string;
  description: string;
};

export type DemoProjectResult = {
  project: Project;
  upload_result: UploadResult;
  analysis_job: AnalysisJob;
};
