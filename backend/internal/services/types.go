package services

import (
	"hakaton/backend/internal/models"
)

const (
	analysisSourceBackendLocal = "backend_local"
	reviewThreshold            = 0.45

	statusOK           = "ok"
	statusMissingFile  = "missing_file"
	statusMissingLabel = "missing_label"
	statusUnknownClass = "unknown_class"
	statusInvalid      = "invalid"
)

type CreateProjectRequest struct {
	Name     string   `json:"name"`
	Modality string   `json:"modality"`
	TaskType string   `json:"task_type"`
	Classes  []string `json:"classes"`
}

type UploadResult struct {
	DatasetVersion models.DatasetVersion `json:"dataset_version"`
	ObjectsCount   int                   `json:"objects_count"`
	InvalidObjects int                   `json:"invalid_objects"`
}

type Dashboard struct {
	Project                  models.Project        `json:"project"`
	DatasetVersion           models.DatasetVersion `json:"dataset_version"`
	ObjectsCount             int                   `json:"objects_count"`
	ReadinessScore           float64               `json:"readiness_score"`
	StatusCounts             map[string]int        `json:"status_counts"`
	LabelCounts              map[string]int        `json:"label_counts"`
	ClassDistribution        map[string]float64    `json:"class_distribution"`
	ImbalanceIndex           float64               `json:"imbalance_index"`
	AvgEntropy               float64               `json:"avg_entropy"`
	AvgLabelErrorProbability float64               `json:"avg_label_error_probability"`
	MissingFilesCount        int                   `json:"missing_files_count"`
	ReviewItems              int                   `json:"review_items"`
	AnalysisSource           string                `json:"analysis_source"`
}

type ProbabilisticAnalysis struct {
	DatasetVersion           models.DatasetVersion `json:"dataset_version"`
	Metrics                  []models.ObjectMetric `json:"metrics"`
	ClassDistribution        map[string]float64    `json:"class_distribution"`
	ImbalanceIndex           float64               `json:"imbalance_index"`
	AvgEntropy               float64               `json:"avg_entropy"`
	AvgLabelErrorProbability float64               `json:"avg_label_error_probability"`
	MissingFilesCount        int                   `json:"missing_files_count"`
	ReviewItems              int                   `json:"review_items"`
	AnalysisSource           string                `json:"analysis_source"`
}

type ReviewQueueItem struct {
	Object models.DataObject   `json:"object"`
	Metric models.ObjectMetric `json:"metric"`
}

type ObjectFile struct {
	Path        string `json:"-"`
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

type datasetSummary struct {
	ClassDistribution        map[string]float64 `json:"class_distribution"`
	ImbalanceIndex           float64            `json:"imbalance_index"`
	AvgEntropy               float64            `json:"avg_entropy"`
	AvgLabelErrorProbability float64            `json:"avg_label_error_probability"`
	MissingFilesCount        int                `json:"missing_files_count"`
	ReviewItems              int                `json:"review_items"`
	AnalysisSource           string             `json:"analysis_source"`
}
