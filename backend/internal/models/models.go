package models

import (
	"time"

	"github.com/google/uuid"
)

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
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	DatasetVersionID uuid.UUID `json:"dataset_version_id"`
	Status           string    `json:"status"`
	ErrorMessage     string    `json:"error_message,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	FinishedAt       time.Time `json:"finished_at,omitempty"`
}

type Export struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	DatasetVersionID uuid.UUID `json:"dataset_version_id"`
	FilePath         string    `json:"file_path"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}
