package services

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hakaton/backend/internal/mlclient"
	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

const analysisSourceMLService = "ml_service"

type mlRecommendation struct {
	Type                 string   `json:"type"`
	Priority             string   `json:"priority"`
	Title                string   `json:"title"`
	Description          string   `json:"description"`
	AffectedObjectsCount int      `json:"affected_objects_count"`
	ObjectIDs            []string `json:"object_ids"`
}

type mlRoadmapItem struct {
	Priority       int    `json:"priority"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	ActionType     string `json:"action_type"`
	ExpectedImpact string `json:"expected_impact"`
}

func (s *Service) analyzeWithMLService(ctx context.Context, project models.Project, version models.DatasetVersion, objects []models.DataObject) (localAnalysis, error) {
	if s.ml == nil {
		return localAnalysis{}, errors.New("ml-service client is not configured")
	}
	if project.Modality != "image" || project.TaskType != "classification" {
		return localAnalysis{}, errors.New("ml-service supports only image classification projects")
	}

	datasetPath := s.storage.DatasetPath(project.ID)
	if _, err := os.Stat(datasetPath); err != nil {
		return localAnalysis{}, fmt.Errorf("dataset.csv is not available for ml-service: %w", err)
	}
	imagesDir := s.storage.ImagesDir(project.ID)
	if _, err := os.Stat(imagesDir); err != nil {
		return localAnalysis{}, fmt.Errorf("images directory is not available for ml-service: %w", err)
	}

	outputDir := filepath.Join(s.storage.AnalysisDir(project.ID), version.ID.String())
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return localAnalysis{}, err
	}

	result, err := s.ml.AnalyzeDataset(ctx, mlclient.DatasetAnalysisRequest{
		DatasetPath: datasetPath,
		ImagesDir:   imagesDir,
		OutputDir:   outputDir,
	})
	if err != nil {
		return localAnalysis{}, err
	}

	metrics, statuses, err := metricsFromMLResults(result.Files["results_csv"], objects)
	if err != nil {
		return localAnalysis{}, err
	}
	recommendations, err := recommendationsFromMLFile(result.Files["recommendations_json"], project, version, objects)
	if err != nil {
		return localAnalysis{}, err
	}
	roadmap, err := roadmapFromMLFile(result.Files["roadmap_json"], project, version)
	if err != nil {
		return localAnalysis{}, err
	}

	objectsWithStatuses := applyObjectStatuses(objects, statuses)
	summary := summarizeDataset(project, objectsWithStatuses, metrics)
	summary.AnalysisSource = analysisSourceMLService

	return localAnalysis{
		readiness:       result.DatasetReadinessScore,
		metrics:         metrics,
		recommendations: recommendations,
		roadmap:         roadmap,
		summary:         summary,
		objectStatuses:  statuses,
	}, nil
}

func metricsFromMLResults(path string, objects []models.DataObject) ([]models.ObjectMetric, map[uuid.UUID]string, error) {
	rows, err := readCSVRows(path)
	if err != nil {
		return nil, nil, err
	}
	matcher := newMLObjectMatcher(objects)
	metrics := make([]models.ObjectMetric, 0, len(rows))
	statuses := map[uuid.UUID]string{}

	for index, row := range rows {
		object, ok := matcher.match(row, index)
		if !ok {
			return nil, nil, fmt.Errorf("ml-service result row %d cannot be mapped to a backend object", index+2)
		}
		status := strings.TrimSpace(row["status"])
		if shouldApplyMLStatus(object.Status, status) {
			statuses[object.ID] = status
			object.Status = status
		}

		duplicateScore := csvFloat(row, "duplicate_score")
		qualityScore := csvFloat(row, "quality_score")
		metric := models.ObjectMetric{
			ID:                    uuid.New(),
			ObjectID:              object.ID,
			Entropy:               csvFloat(row, "entropy"),
			UncertaintyScore:      csvFloat(row, "uncertainty_score"),
			LabelErrorProbability: csvFloat(row, "label_error_probability"),
			DuplicateScore:        duplicateScore,
			RarityScore:           csvFloat(row, "class_deficit_score"),
			ClassDeficitScore:     csvFloat(row, "class_deficit_score"),
			QualityScore:          qualityScore,
			NoveltyScore:          clamp(1-duplicateScore, 0, 1),
			ObjectUtilityScore:    csvFloat(row, "object_utility_score"),
			FinalScore:            csvFloat(row, "final_score"),
			Reasons:               splitMLReasons(row["reasons"]),
			Recommendation:        strings.TrimSpace(row["recommendation"]),
			Probabilities:         mlProbabilities(row, status),
		}
		metrics = append(metrics, metric)
	}

	if len(metrics) == 0 {
		return nil, nil, errors.New("ml-service results.csv contains no metrics")
	}
	return metrics, statuses, nil
}

func recommendationsFromMLFile(path string, project models.Project, version models.DatasetVersion, objects []models.DataObject) ([]models.Recommendation, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("ml-service response is missing recommendations_json")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var raw []mlRecommendation
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return nil, err
	}

	mapped := make([]models.Recommendation, 0, len(raw))
	for _, item := range raw {
		objectIDs := mapExternalIDsToObjects(item.ObjectIDs, objects)
		affected := item.AffectedObjectsCount
		if affected == 0 {
			affected = len(objectIDs)
		}
		mapped = append(mapped, models.Recommendation{
			ID:                   uuid.New(),
			ProjectID:            project.ID,
			DatasetVersionID:     version.ID,
			Type:                 strings.TrimSpace(item.Type),
			Priority:             strings.TrimSpace(item.Priority),
			Title:                strings.TrimSpace(item.Title),
			Description:          strings.TrimSpace(item.Description),
			AffectedObjectsCount: affected,
			ObjectIDs:            objectIDs,
		})
	}
	return mapped, nil
}

func roadmapFromMLFile(path string, project models.Project, version models.DatasetVersion) ([]models.RoadmapItem, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("ml-service response is missing roadmap_json")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var raw []mlRoadmapItem
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return nil, err
	}

	mapped := make([]models.RoadmapItem, 0, len(raw))
	for _, item := range raw {
		mapped = append(mapped, models.RoadmapItem{
			ID:               uuid.New(),
			ProjectID:        project.ID,
			DatasetVersionID: version.ID,
			Priority:         item.Priority,
			Title:            strings.TrimSpace(item.Title),
			Description:      strings.TrimSpace(item.Description),
			ActionType:       strings.TrimSpace(item.ActionType),
			ExpectedImpact:   strings.TrimSpace(item.ExpectedImpact),
		})
	}
	return mapped, nil
}

func readCSVRows(path string) ([]map[string]string, error) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.New("empty csv path")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	rows := make([]map[string]string, 0, len(records))
	for _, record := range records {
		row := map[string]string{}
		for i, header := range headers {
			if i < len(record) {
				row[header] = strings.TrimSpace(record[i])
			}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

type mlObjectMatcher struct {
	objects    []models.DataObject
	byExternal map[string]int
	byFile     map[string][]int
	used       map[int]struct{}
}

func newMLObjectMatcher(objects []models.DataObject) *mlObjectMatcher {
	matcher := &mlObjectMatcher{
		objects:    objects,
		byExternal: map[string]int{},
		byFile:     map[string][]int{},
		used:       map[int]struct{}{},
	}
	for index, object := range objects {
		if object.ExternalID != "" {
			matcher.byExternal[object.ExternalID] = index
		}
		if object.FilePath != "" {
			key := normalizedRelativePath(object.FilePath)
			matcher.byFile[key] = append(matcher.byFile[key], index)
		}
	}
	return matcher
}

func (m *mlObjectMatcher) match(row map[string]string, rowIndex int) (models.DataObject, bool) {
	if index, ok := m.byExternal[strings.TrimSpace(row["id"])]; ok {
		if _, used := m.used[index]; !used {
			m.used[index] = struct{}{}
			return m.objects[index], true
		}
	}

	fileKey := normalizedRelativePath(row["file_path"])
	for _, index := range m.byFile[fileKey] {
		if _, used := m.used[index]; !used {
			m.used[index] = struct{}{}
			return m.objects[index], true
		}
	}

	if rowIndex >= 0 && rowIndex < len(m.objects) {
		if _, used := m.used[rowIndex]; !used {
			m.used[rowIndex] = struct{}{}
			return m.objects[rowIndex], true
		}
	}
	return models.DataObject{}, false
}

func mapExternalIDsToObjects(externalIDs []string, objects []models.DataObject) []uuid.UUID {
	byExternal := map[string]uuid.UUID{}
	for _, object := range objects {
		if object.ExternalID != "" {
			byExternal[object.ExternalID] = object.ID
		}
	}
	result := make([]uuid.UUID, 0, len(externalIDs))
	seen := map[uuid.UUID]struct{}{}
	for _, externalID := range externalIDs {
		id, ok := byExternal[strings.TrimSpace(externalID)]
		if !ok || id == uuid.Nil {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizedRelativePath(value string) string {
	value = filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
	value = strings.TrimPrefix(value, "./")
	return value
}

func csvFloat(row map[string]string, key string) float64 {
	value := strings.TrimSpace(row[key])
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func splitMLReasons(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ";")
	reasons := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			reasons = append(reasons, part)
		}
	}
	return reasons
}

func mlProbabilities(row map[string]string, status string) map[string]any {
	payload := map[string]any{
		"analysis_source": analysisSourceMLService,
	}
	if status != "" {
		payload["ml_status"] = status
	}
	for key := range row {
		if !strings.HasPrefix(key, "prob_") {
			continue
		}
		class := strings.TrimPrefix(key, "prob_")
		if class == "" {
			continue
		}
		payload[class] = csvFloat(row, key)
	}
	return payload
}

func shouldApplyMLStatus(currentStatus, mlStatus string) bool {
	mlStatus = strings.TrimSpace(mlStatus)
	if mlStatus == "" {
		return false
	}
	return currentStatus == "" || currentStatus == statusOK || isMLCurationStatus(currentStatus)
}

func isMLCurationStatus(status string) bool {
	switch status {
	case "ok", "hard_example", "rare_class_candidate", "suspected_label_error", "duplicate", "bad_quality", "exclude_candidate":
		return true
	default:
		return false
	}
}

func applyObjectStatuses(objects []models.DataObject, statuses map[uuid.UUID]string) []models.DataObject {
	result := make([]models.DataObject, len(objects))
	copy(result, objects)
	for index := range result {
		if status := statuses[result[index].ID]; status != "" {
			result[index].Status = status
		}
	}
	return result
}
