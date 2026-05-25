package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAnnotator    UserRole = "annotator"
	RoleMLEngineer   UserRole = "ml_engineer"
	RoleDomainExpert UserRole = "domain_expert"
	RoleDataAnalyst  UserRole = "data_analyst"
	RoleAdmin        UserRole = "admin"
)

type ProjectMemberRole string

const (
	ProjectRoleAdmin        ProjectMemberRole = "admin"
	ProjectRoleMLEngineer   ProjectMemberRole = "ml_engineer"
	ProjectRoleAnnotator    ProjectMemberRole = "annotator"
	ProjectRoleDomainExpert ProjectMemberRole = "domain_expert"
	ProjectRoleDataAnalyst  ProjectMemberRole = "data_analyst"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Team struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedBy uuid.UUID `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

type TeamMember struct {
	ID        uuid.UUID `json:"id"`
	TeamID    uuid.UUID `json:"team_id"`
	UserID    uuid.UUID `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ProjectMember struct {
	ID        uuid.UUID         `json:"id"`
	ProjectID uuid.UUID         `json:"project_id"`
	UserID    uuid.UUID         `json:"user_id"`
	Role      ProjectMemberRole `json:"role"`
	CreatedAt time.Time         `json:"created_at"`
}

type Project struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Modality  string    `json:"modality"`
	TaskType  string    `json:"task_type"`
	Classes   []string  `json:"classes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DatasetVersion struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	VersionName    string    `json:"version_name"`
	Status         string    `json:"status"`
	ObjectsCount   int       `json:"objects_count"`
	ReadinessScore float64   `json:"readiness_score"`
	CreatedAt      time.Time `json:"created_at"`
}

type DataObject struct {
	ID               uuid.UUID      `json:"id"`
	DatasetVersionID uuid.UUID      `json:"dataset_version_id"`
	ExternalID       string         `json:"external_id,omitempty"`
	FilePath         string         `json:"file_path,omitempty"`
	TextContent      string         `json:"text_content,omitempty"`
	Label            string         `json:"label,omitempty"`
	PredictedLabel   string         `json:"predicted_label,omitempty"`
	Confidence       float64        `json:"confidence,omitempty"`
	Split            string         `json:"split,omitempty"`
	Source           string         `json:"source,omitempty"`
	Annotator        string         `json:"annotator,omitempty"`
	Metadata         map[string]any `json:"metadata,omitempty"`
	Status           string         `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	Metrics          *ObjectMetric  `json:"metrics,omitempty"`
}

type ObjectMetric struct {
	ID                    uuid.UUID      `json:"id"`
	ObjectID              uuid.UUID      `json:"object_id"`
	Entropy               float64        `json:"entropy"`
	UncertaintyScore      float64        `json:"uncertainty_score"`
	LabelErrorProbability float64        `json:"label_error_probability"`
	DuplicateScore        float64        `json:"duplicate_score"`
	RarityScore           float64        `json:"rarity_score"`
	ClassDeficitScore     float64        `json:"class_deficit_score"`
	QualityScore          float64        `json:"quality_score"`
	NoveltyScore          float64        `json:"novelty_score"`
	ObjectUtilityScore    float64        `json:"object_utility_score"`
	FinalScore            float64        `json:"final_score"`
	Reasons               []string       `json:"reasons,omitempty"`
	Recommendation        string         `json:"recommendation,omitempty"`
	Probabilities         map[string]any `json:"probabilities,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
}

type Recommendation struct {
	ID                   uuid.UUID   `json:"id"`
	ProjectID            uuid.UUID   `json:"project_id"`
	DatasetVersionID     uuid.UUID   `json:"dataset_version_id"`
	Type                 string      `json:"type"`
	Priority             string      `json:"priority"`
	Title                string      `json:"title"`
	Description          string      `json:"description"`
	AffectedObjectsCount int         `json:"affected_objects_count"`
	ObjectIDs            []uuid.UUID `json:"object_ids,omitempty"`
	CreatedAt            time.Time   `json:"created_at"`
}

type RoadmapItem struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	DatasetVersionID uuid.UUID `json:"dataset_version_id"`
	Priority         int       `json:"priority"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	ActionType       string    `json:"action_type"`
	ExpectedImpact   string    `json:"expected_impact"`
	CreatedAt        time.Time `json:"created_at"`
}

type AnalysisJob struct {
	ID               uuid.UUID          `json:"id"`
	ProjectID        uuid.UUID          `json:"project_id"`
	DatasetVersionID uuid.UUID          `json:"dataset_version_id"`
	Status           string             `json:"status"`
	ErrorMessage     string             `json:"error_message,omitempty"`
	ProgressPercent  int                `json:"progress_percent"`
	ProgressStage    string             `json:"progress_stage,omitempty"`
	OutputFiles      map[string]string  `json:"output_files,omitempty"`
	Files            map[string]string  `json:"files,omitempty"`
	Warnings         []string           `json:"warnings,omitempty"`
	Errors           []string           `json:"errors,omitempty"`
	StartedAt        time.Time          `json:"started_at"`
	FinishedAt       time.Time          `json:"finished_at,omitempty"`
}

type ObjectAction struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	ObjectID  uuid.UUID `json:"object_id"`
	UserID    uuid.UUID `json:"user_id"`
	Action    string    `json:"action"`
	OldValue  string    `json:"old_value,omitempty"`
	NewValue  string    `json:"new_value,omitempty"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ObjectComment struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	ObjectID  uuid.UUID `json:"object_id"`
	UserID    uuid.UUID `json:"user_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type ReviewAssignment struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	ObjectID  uuid.UUID `json:"object_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CollectionTask struct {
	ID          uuid.UUID `json:"id"`
	ProjectID   uuid.UUID `json:"project_id"`
	TargetClass string    `json:"target_class"`
	TargetCount int       `json:"target_count"`
	Priority    string    `json:"priority"`
	Risk        string    `json:"risk"`
	Status      string    `json:"status"`
	CreatedBy   uuid.UUID `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type SyntheticTask struct {
	ID             uuid.UUID `json:"id"`
	ProjectID      uuid.UUID `json:"project_id"`
	TargetClass    string    `json:"target_class"`
	TargetCount    int       `json:"target_count"`
	Prompt         string    `json:"prompt,omitempty"`
	NegativePrompt string    `json:"negative_prompt,omitempty"`
	Priority       string    `json:"priority"`
	Risk           string    `json:"risk"`
	Status         string    `json:"status"`
	CreatedBy      uuid.UUID `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
}

type ClassActionPlanItem struct {
	Class                     string  `json:"class"`
	Problem                   string  `json:"problem"`
	RecommendedAction         string  `json:"recommended_action"`
	RealCollectionPriority    float64 `json:"real_collection_priority"`
	SyntheticDataCandidateScore float64 `json:"synthetic_data_candidate_score"`
	Risk                      string  `json:"risk"`
	ExpectedImpact            string  `json:"expected_impact"`
}

type Export struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	DatasetVersionID uuid.UUID `json:"dataset_version_id"`
	FilePath         string    `json:"file_path"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}
