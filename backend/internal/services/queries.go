package services

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func (s *Service) Dashboard(ctx context.Context, projectID uuid.UUID) (Dashboard, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return Dashboard{}, err
	}
	version, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return Dashboard{}, err
	}

	statusCounts := map[string]int{}
	labelCounts := map[string]int{}
	for _, object := range objects {
		statusCounts[object.Status]++
		label := object.Label
		if label == "" {
			label = "unlabeled"
		}
		labelCounts[label]++
	}
	summary := summarizeDataset(project, objects, metrics)
	return Dashboard{
		Project:                  project,
		DatasetVersion:           version,
		ObjectsCount:             len(objects),
		ReadinessScore:           version.ReadinessScore,
		StatusCounts:             statusCounts,
		LabelCounts:              labelCounts,
		ClassDistribution:        summary.ClassDistribution,
		ImbalanceIndex:           summary.ImbalanceIndex,
		AvgEntropy:               summary.AvgEntropy,
		AvgLabelErrorProbability: summary.AvgLabelErrorProbability,
		MissingFilesCount:        summary.MissingFilesCount,
		ReviewItems:              summary.ReviewItems,
		AnalysisSource:           summary.AnalysisSource,
	}, nil
}

func (s *Service) ProbabilisticAnalysis(ctx context.Context, projectID uuid.UUID) (ProbabilisticAnalysis, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return ProbabilisticAnalysis{}, err
	}
	version, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return ProbabilisticAnalysis{}, err
	}
	summary := summarizeDataset(project, objects, metrics)
	return ProbabilisticAnalysis{
		DatasetVersion:           version,
		Metrics:                  metrics,
		ClassDistribution:        summary.ClassDistribution,
		ImbalanceIndex:           summary.ImbalanceIndex,
		AvgEntropy:               summary.AvgEntropy,
		AvgLabelErrorProbability: summary.AvgLabelErrorProbability,
		MissingFilesCount:        summary.MissingFilesCount,
		ReviewItems:              summary.ReviewItems,
		AnalysisSource:           summary.AnalysisSource,
	}, nil
}

func (s *Service) ReviewQueue(ctx context.Context, projectID uuid.UUID) ([]ReviewQueueItem, error) {
	_, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return nil, err
	}
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
	return queue, nil
}

func (s *Service) Object(ctx context.Context, projectID, objectID uuid.UUID) (models.DataObject, error) {
	return s.repo.GetObject(ctx, projectID, objectID)
}

func (s *Service) ObjectFile(ctx context.Context, projectID, objectID uuid.UUID) (ObjectFile, error) {
	object, err := s.repo.GetObject(ctx, projectID, objectID)
	if err != nil {
		return ObjectFile{}, err
	}
	if object.FilePath == "" {
		return ObjectFile{}, badRequest("object has no file_path")
	}
	path, err := s.storage.SafeImagePath(projectID, object.FilePath)
	if err != nil {
		return ObjectFile{}, err
	}
	return ObjectFile{
		Path:        path,
		FileName:    filepath.Base(path),
		ContentType: s.storage.ImageContentType(path),
	}, nil
}

func (s *Service) GetJob(ctx context.Context, jobID uuid.UUID) (models.AnalysisJob, error) {
	return s.repo.GetJob(ctx, jobID)
}

func (s *Service) ListJobs(ctx context.Context, projectID uuid.UUID) ([]models.AnalysisJob, error) {
	return s.repo.ListJobsByProject(ctx, projectID)
}

func (s *Service) Recommendations(ctx context.Context, projectID uuid.UUID) ([]models.Recommendation, error) {
	version, err := s.repo.LatestDatasetVersion(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListRecommendationsForVersion(ctx, projectID, version.ID)
}

func (s *Service) Roadmap(ctx context.Context, projectID uuid.UUID) ([]models.RoadmapItem, error) {
	version, err := s.repo.LatestDatasetVersion(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return s.repo.ListRoadmapForVersion(ctx, projectID, version.ID)
}

func (s *Service) AgentSummary(ctx context.Context, projectID uuid.UUID) (map[string]any, error) {
	dashboard, err := s.Dashboard(ctx, projectID)
	if err != nil {
		return nil, err
	}
	recommendations, err := s.Recommendations(ctx, projectID)
	if err != nil {
		return nil, err
	}
	roadmap, err := s.Roadmap(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"summary":         agentSummaryText(dashboard),
		"dashboard":       dashboard,
		"recommendations": recommendations,
		"roadmap":         roadmap,
	}, nil
}

func (s *Service) latestObjectsAndMetrics(ctx context.Context, projectID uuid.UUID) (models.DatasetVersion, []models.DataObject, []models.ObjectMetric, error) {
	version, err := s.repo.LatestDatasetVersion(ctx, projectID)
	if err != nil {
		return models.DatasetVersion{}, nil, nil, err
	}
	objects, err := s.repo.ListObjects(ctx, version.ID)
	if err != nil {
		return models.DatasetVersion{}, nil, nil, err
	}
	metrics, err := s.repo.ListMetrics(ctx, version.ID)
	if err != nil {
		return models.DatasetVersion{}, nil, nil, err
	}
	return version, objects, metrics, nil
}

func summarizeDataset(project models.Project, objects []models.DataObject, metrics []models.ObjectMetric) datasetSummary {
	classes := classUniverse(project, objects)
	counts, labeledTotal := labelCounts(objects, classes)
	source := ""
	if len(metrics) > 0 {
		source = analysisSourceFromMetrics(metrics)
	}
	return datasetSummary{
		ClassDistribution:        classDistribution(counts, classes, labeledTotal),
		ImbalanceIndex:           imbalanceIndex(classDistribution(counts, classes, labeledTotal), classes, labeledTotal),
		AvgEntropy:               averageMetric(metrics, func(metric models.ObjectMetric) float64 { return metric.Entropy }),
		AvgLabelErrorProbability: averageMetric(metrics, func(metric models.ObjectMetric) float64 { return metric.LabelErrorProbability }),
		MissingFilesCount:        countObjects(objects, func(object models.DataObject) bool { return object.Status == statusMissingFile }),
		ReviewItems:              countReviewItems(objects, metrics),
		AnalysisSource:           source,
	}
}

func sortReviewQueue(queue []ReviewQueueItem) {
	sort.Slice(queue, func(i, j int) bool {
		return queue[i].Metric.FinalScore > queue[j].Metric.FinalScore
	})
}

func shouldReviewItem(object models.DataObject, metric models.ObjectMetric) bool {
	if object.Status != "" && object.Status != statusOK {
		return true
	}
	return shouldReviewMetric(metric)
}

func shouldReviewMetric(metric models.ObjectMetric) bool {
	if metric.FinalScore >= reviewThreshold {
		return true
	}
	recommendation := strings.TrimSpace(metric.Recommendation)
	return recommendation != "" && recommendation != "keep"
}

func countReviewItems(objects []models.DataObject, metrics []models.ObjectMetric) int {
	byID := map[uuid.UUID]models.DataObject{}
	for _, object := range objects {
		byID[object.ID] = object
	}
	count := 0
	for _, metric := range metrics {
		object := byID[metric.ObjectID]
		if shouldReviewItem(object, metric) {
			count++
		}
	}
	return count
}

func analysisSourceFromMetrics(metrics []models.ObjectMetric) string {
	for _, metric := range metrics {
		if source, ok := metric.Probabilities["analysis_source"].(string); ok && source != "" {
			return source
		}
	}
	return analysisSourceBackendLocal
}

func agentSummaryText(dashboard Dashboard) string {
	return fmt.Sprintf(
		"Dataset readiness is %.1f/100 with %d objects, %.2f imbalance index, %.2f average label error probability, and %d review items.",
		dashboard.ReadinessScore,
		dashboard.ObjectsCount,
		dashboard.ImbalanceIndex,
		dashboard.AvgLabelErrorProbability,
		dashboard.ReviewItems,
	)
}
