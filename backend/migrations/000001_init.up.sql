CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    modality TEXT NOT NULL,
    task_type TEXT NOT NULL,
    classes JSONB,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dataset_versions (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    version_name TEXT NOT NULL,
    status TEXT NOT NULL,
    objects_count INT DEFAULT 0,
    readiness_score FLOAT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS data_objects (
    id UUID PRIMARY KEY,
    dataset_version_id UUID REFERENCES dataset_versions(id) ON DELETE CASCADE,
    external_id TEXT,
    file_path TEXT,
    text_content TEXT,
    label TEXT,
    predicted_label TEXT,
    confidence FLOAT,
    split TEXT,
    source TEXT,
    annotator TEXT,
    metadata JSONB,
    status TEXT DEFAULT 'ok',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS object_metrics (
    id UUID PRIMARY KEY,
    object_id UUID REFERENCES data_objects(id) ON DELETE CASCADE,
    entropy FLOAT,
    uncertainty_score FLOAT,
    label_error_probability FLOAT,
    duplicate_score FLOAT,
    rarity_score FLOAT,
    class_deficit_score FLOAT,
    quality_score FLOAT,
    novelty_score FLOAT,
    object_utility_score FLOAT,
    final_score FLOAT,
    reasons JSONB,
    recommendation TEXT,
    probabilities JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recommendations (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    dataset_version_id UUID REFERENCES dataset_versions(id) ON DELETE CASCADE,
    type TEXT,
    priority TEXT,
    title TEXT,
    description TEXT,
    affected_objects_count INT,
    object_ids JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS roadmap_items (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    dataset_version_id UUID REFERENCES dataset_versions(id) ON DELETE CASCADE,
    priority INT,
    title TEXT,
    description TEXT,
    action_type TEXT,
    expected_impact TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS analysis_jobs (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    dataset_version_id UUID REFERENCES dataset_versions(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    error_message TEXT,
    started_at TIMESTAMP DEFAULT NOW(),
    finished_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS exports (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    dataset_version_id UUID REFERENCES dataset_versions(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dataset_versions_project_id ON dataset_versions(project_id);
CREATE INDEX IF NOT EXISTS idx_data_objects_dataset_version_id ON data_objects(dataset_version_id);
CREATE INDEX IF NOT EXISTS idx_object_metrics_object_id ON object_metrics(object_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_project_id ON recommendations(project_id);
CREATE INDEX IF NOT EXISTS idx_roadmap_items_project_id ON roadmap_items(project_id);
