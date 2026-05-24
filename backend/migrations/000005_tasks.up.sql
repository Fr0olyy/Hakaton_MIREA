CREATE TABLE IF NOT EXISTS collection_tasks (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    target_class TEXT NOT NULL,
    target_count INT NOT NULL,
    priority TEXT DEFAULT 'medium',
    risk TEXT DEFAULT 'low',
    status TEXT DEFAULT 'draft',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS synthetic_tasks (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    target_class TEXT NOT NULL,
    target_count INT NOT NULL,
    prompt TEXT,
    negative_prompt TEXT,
    priority TEXT DEFAULT 'medium',
    risk TEXT DEFAULT 'low',
    status TEXT DEFAULT 'draft',
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_collection_tasks_project_id ON collection_tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_synthetic_tasks_project_id ON synthetic_tasks(project_id);
