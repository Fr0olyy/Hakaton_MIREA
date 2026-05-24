CREATE TABLE IF NOT EXISTS object_actions (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID REFERENCES data_objects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    action TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT,
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS object_comments (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID REFERENCES data_objects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS review_assignments (
    id UUID PRIMARY KEY,
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID REFERENCES data_objects(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    status TEXT DEFAULT 'assigned',
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, object_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_object_actions_project_id ON object_actions(project_id);
CREATE INDEX IF NOT EXISTS idx_object_actions_object_id ON object_actions(object_id);
CREATE INDEX IF NOT EXISTS idx_object_comments_object_id ON object_comments(object_id);
CREATE INDEX IF NOT EXISTS idx_review_assignments_user_id ON review_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_review_assignments_project_id ON review_assignments(project_id);
