package services

import (
	"archive/zip"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type zipSource struct {
	Name string
	Path string
}

func (s *Service) Export(ctx context.Context, projectID uuid.UUID) (models.Export, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return models.Export{}, err
	}
	version, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return models.Export{}, err
	}
	recommendations, err := s.repo.ListRecommendationsForVersion(ctx, projectID, version.ID)
	if err != nil {
		return models.Export{}, err
	}
	roadmap, err := s.repo.ListRoadmapForVersion(ctx, projectID, version.ID)
	if err != nil {
		return models.Export{}, err
	}

	exportID := uuid.New()
	exportDir := filepath.Join(s.storage.ExportsDir(projectID), exportID.String())
	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return models.Export{}, err
	}

	metricsByObject := metricsByObjectID(metrics)
	csvPath := filepath.Join(exportDir, "dataset_v2.csv")
	if err := writeDatasetCSV(csvPath, objects, metricsByObject); err != nil {
		return models.Export{}, err
	}

	summary := summarizeDataset(project, objects, metrics)
	reportPath := filepath.Join(exportDir, "dataset_report.json")
	if err := writeJSONFile(reportPath, map[string]any{
		"project":         project,
		"dataset_version": version,
		"summary":         summary,
		"objects_count":   len(objects),
		"metrics_count":   len(metrics),
	}); err != nil {
		return models.Export{}, err
	}

	reviewQueue := reviewQueueFrom(objects, metrics)
	reviewQueuePath := filepath.Join(exportDir, "review_queue.csv")
	if err := writeReviewQueueCSV(reviewQueuePath, reviewQueue); err != nil {
		return models.Export{}, err
	}

	recommendationsPath := filepath.Join(exportDir, "recommendations.json")
	if err := writeJSONFile(recommendationsPath, recommendations); err != nil {
		return models.Export{}, err
	}
	roadmapPath := filepath.Join(exportDir, "roadmap.json")
	if err := writeJSONFile(roadmapPath, roadmap); err != nil {
		return models.Export{}, err
	}

	sources := []zipSource{
		{Name: "dataset_v2.csv", Path: csvPath},
		{Name: "dataset_report.json", Path: reportPath},
		{Name: "review_queue.csv", Path: reviewQueuePath},
		{Name: "recommendations.json", Path: recommendationsPath},
		{Name: "roadmap.json", Path: roadmapPath},
	}
	usedImageNames := map[string]struct{}{}
	for _, object := range objects {
		if !shouldExportObject(object) || object.FilePath == "" {
			continue
		}
		path, err := s.storage.SafeImagePath(projectID, object.FilePath)
		if err != nil {
			continue
		}
		name := exportImageName(object.FilePath)
		if _, ok := usedImageNames[name]; ok {
			continue
		}
		usedImageNames[name] = struct{}{}
		sources = append(sources, zipSource{Name: name, Path: path})
	}

	zipPath := filepath.Join(exportDir, "export.zip")
	if err := zipFiles(zipPath, sources); err != nil {
		return models.Export{}, err
	}
	return s.repo.CreateExport(ctx, models.Export{
		ID:               exportID,
		ProjectID:        projectID,
		DatasetVersionID: version.ID,
		FilePath:         zipPath,
		Status:           "ready",
	})
}

func (s *Service) GetExport(ctx context.Context, projectID, exportID uuid.UUID) (models.Export, error) {
	return s.repo.GetExport(ctx, projectID, exportID)
}

func writeDatasetCSV(path string, objects []models.DataObject, metrics map[uuid.UUID]models.ObjectMetric) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"external_id",
		"file_path",
		"text_content",
		"label",
		"predicted_label",
		"confidence",
		"split",
		"source",
		"annotator",
		"status",
		"entropy",
		"uncertainty_score",
		"label_error_probability",
		"class_deficit_score",
		"novelty_score",
		"object_utility_score",
		"final_score",
		"review_recommendation",
	}
	if err := writer.Write(headers); err != nil {
		return err
	}
	for _, object := range objects {
		if !shouldExportObject(object) {
			continue
		}
		metric := metrics[object.ID]
		if err := writer.Write([]string{
			object.ExternalID,
			object.FilePath,
			object.TextContent,
			object.Label,
			object.PredictedLabel,
			formatFloat(object.Confidence),
			object.Split,
			object.Source,
			object.Annotator,
			object.Status,
			formatFloat(metric.Entropy),
			formatFloat(metric.UncertaintyScore),
			formatFloat(metric.LabelErrorProbability),
			formatFloat(metric.ClassDeficitScore),
			formatFloat(metric.NoveltyScore),
			formatFloat(metric.ObjectUtilityScore),
			formatFloat(metric.FinalScore),
			metric.Recommendation,
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeReviewQueueCSV(path string, queue []ReviewQueueItem) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write([]string{"object_id", "external_id", "file_path", "label", "predicted_label", "final_score", "recommendation", "reasons"}); err != nil {
		return err
	}
	for _, item := range queue {
		if err := writer.Write([]string{
			item.Object.ID.String(),
			item.Object.ExternalID,
			item.Object.FilePath,
			item.Object.Label,
			item.Object.PredictedLabel,
			formatFloat(item.Metric.FinalScore),
			item.Metric.Recommendation,
			strings.Join(item.Metric.Reasons, "; "),
		}); err != nil {
			return err
		}
	}
	return writer.Error()
}

func writeJSONFile(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func zipFiles(target string, sources []zipSource) error {
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].Name < sources[j].Name
	})

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := zip.NewWriter(out)
	defer writer.Close()

	seen := map[string]struct{}{}
	for _, source := range sources {
		name := filepath.ToSlash(filepath.Clean(source.Name))
		if strings.HasPrefix(name, "../") || strings.HasPrefix(name, "/") || name == "." {
			return fmt.Errorf("unsafe export path: %s", source.Name)
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		src, err := os.Open(source.Path)
		if err != nil {
			return err
		}
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		dst, err := writer.CreateHeader(header)
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := src.Close()
		if copyErr != nil || closeErr != nil {
			return errors.Join(copyErr, closeErr)
		}
	}
	return nil
}

func reviewQueueFrom(objects []models.DataObject, metrics []models.ObjectMetric) []ReviewQueueItem {
	byID := map[uuid.UUID]models.DataObject{}
	for _, object := range objects {
		byID[object.ID] = object
	}
	queue := make([]ReviewQueueItem, 0, len(metrics))
	for _, metric := range metrics {
		if object, ok := byID[metric.ObjectID]; ok && shouldReviewItem(object, metric) {
			queue = append(queue, ReviewQueueItem{Object: object, Metric: metric})
		}
	}
	sortReviewQueue(queue)
	return queue
}

func shouldExportObject(object models.DataObject) bool {
	switch object.Status {
	case statusMissingFile, statusInvalid, "duplicate", "bad_quality", "suspected_label_error", "exclude_candidate":
		return false
	default:
		return true
	}
}

func metricsByObjectID(metrics []models.ObjectMetric) map[uuid.UUID]models.ObjectMetric {
	result := map[uuid.UUID]models.ObjectMetric{}
	for _, metric := range metrics {
		result[metric.ObjectID] = metric
	}
	return result
}

func exportImageName(filePath string) string {
	clean := filepath.ToSlash(filepath.Clean(filePath))
	clean = strings.TrimPrefix(clean, "./")
	if strings.HasPrefix(clean, "images/") {
		return clean
	}
	return "images/" + clean
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.6f", value)
}
