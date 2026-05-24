package services

import (
	"context"
	"math"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type CreateCollectionTaskRequest struct {
	TargetClass string `json:"target_class"`
	TargetCount int    `json:"target_count"`
	Priority    string `json:"priority"`
	Risk        string `json:"risk"`
}

type CreateSyntheticTaskRequest struct {
	TargetClass    string `json:"target_class"`
	TargetCount    int    `json:"target_count"`
	Prompt         string `json:"prompt"`
	NegativePrompt string `json:"negative_prompt"`
	Priority       string `json:"priority"`
	Risk           string `json:"risk"`
}

func (s *Service) CreateCollectionTask(ctx context.Context, projectID, userID uuid.UUID, req CreateCollectionTaskRequest) (models.CollectionTask, error) {
	req.TargetClass = strings.TrimSpace(req.TargetClass)
	if req.TargetClass == "" || req.TargetCount <= 0 {
		return models.CollectionTask{}, badRequest("target_class and target_count are required")
	}
	task := models.CollectionTask{
		ID:          uuid.New(),
		ProjectID:   projectID,
		TargetClass: req.TargetClass,
		TargetCount: req.TargetCount,
		Priority:    nonEmpty(req.Priority, "medium"),
		Risk:        nonEmpty(req.Risk, "low"),
		Status:      "draft",
		CreatedBy:   userID,
	}
	return s.repo.CreateCollectionTask(ctx, task)
}

func (s *Service) ListCollectionTasks(ctx context.Context, projectID uuid.UUID) ([]models.CollectionTask, error) {
	return s.repo.ListCollectionTasks(ctx, projectID)
}

func (s *Service) UpdateCollectionTask(ctx context.Context, projectID, taskID uuid.UUID, updates map[string]any) (models.CollectionTask, error) {
	if err := s.repo.UpdateCollectionTask(ctx, taskID, updates); err != nil {
		return models.CollectionTask{}, err
	}
	return models.CollectionTask{}, nil
}

func (s *Service) CreateSyntheticTask(ctx context.Context, projectID, userID uuid.UUID, req CreateSyntheticTaskRequest) (models.SyntheticTask, error) {
	req.TargetClass = strings.TrimSpace(req.TargetClass)
	if req.TargetClass == "" || req.TargetCount <= 0 {
		return models.SyntheticTask{}, badRequest("target_class and target_count are required")
	}
	task := models.SyntheticTask{
		ID:             uuid.New(),
		ProjectID:      projectID,
		TargetClass:    req.TargetClass,
		TargetCount:    req.TargetCount,
		Prompt:         strings.TrimSpace(req.Prompt),
		NegativePrompt: strings.TrimSpace(req.NegativePrompt),
		Priority:       nonEmpty(req.Priority, "medium"),
		Risk:           nonEmpty(req.Risk, "low"),
		Status:         "draft",
		CreatedBy:      userID,
	}
	return s.repo.CreateSyntheticTask(ctx, task)
}

func (s *Service) ListSyntheticTasks(ctx context.Context, projectID uuid.UUID) ([]models.SyntheticTask, error) {
	return s.repo.ListSyntheticTasks(ctx, projectID)
}

func (s *Service) UpdateSyntheticTask(ctx context.Context, projectID, taskID uuid.UUID, updates map[string]any) (models.SyntheticTask, error) {
	if err := s.repo.UpdateSyntheticTask(ctx, taskID, updates); err != nil {
		return models.SyntheticTask{}, err
	}
	return models.SyntheticTask{}, nil
}

func (s *Service) ClassActionPlan(ctx context.Context, projectID uuid.UUID) ([]models.ClassActionPlanItem, error) {
	project, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	_, objects, metrics, err := s.latestObjectsAndMetrics(ctx, projectID)
	if err != nil {
		return nil, err
	}

	classes := classUniverse(project, objects)
	counts, labeledTotal := labelCounts(objects, classes)
	if labeledTotal == 0 {
		return nil, nil
	}

	expected := float64(labeledTotal) / float64(len(classes))
	metricsByObject := mapObjectsMetrics(metrics)
	classObjects := mapClassObjects(objects)

	var items []models.ClassActionPlanItem
	for _, class := range classes {
		count := counts[class]
		deficit := expected - float64(count)

		problem := "ok"
		var recommendedAction string
		var risk string
		var impact string
		realPriority := 0.0
		syntheticScore := 0.0

		if deficit >= expected*0.5 {
			problem = "critical_deficit"
			recommendedAction = "collect_real"
			risk = "high"
			impact = "Adding real samples significantly improves class balance and model robustness"
			realPriority = clamp(deficit/expected, 0, 1)
			syntheticScore = clamp(0.3+(deficit/expected)*0.4, 0, 1)
		} else if deficit > 0 {
			problem = "moderate_deficit"
			recommendedAction = "generate_synthetic"
			risk = "medium"
			impact = "Synthetic data can help balance the class with lower collection cost"
			realPriority = clamp(deficit/expected, 0, 1)
			syntheticScore = clamp(0.5+(deficit/expected)*0.3, 0, 1)
		} else {
			recommendedAction = "none"
			risk = "low"
			impact = "Class is adequately represented"
		}

		// Check for label errors in this class
		classObjectsList := classObjects[class]
		var labelErrors int
		for _, obj := range classObjectsList {
			if metric, ok := metricsByObject[obj.ID]; ok && metric.LabelErrorProbability >= 0.55 {
				labelErrors++
			}
		}
		if labelErrors > 0 && problem == "ok" {
			problem = "label_errors"
			recommendedAction = "review_labels"
			risk = "medium"
			impact = "Review and fix potential label errors before training"
		}

		items = append(items, models.ClassActionPlanItem{
			Class:                     class,
			Problem:                   problem,
			RecommendedAction:         recommendedAction,
			RealCollectionPriority:    math.Round(realPriority*100) / 100,
			SyntheticDataCandidateScore: math.Round(syntheticScore*100) / 100,
			Risk:                      risk,
			ExpectedImpact:            impact,
		})
	}
	return items, nil
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func mapObjectsMetrics(metrics []models.ObjectMetric) map[uuid.UUID]models.ObjectMetric {
	result := map[uuid.UUID]models.ObjectMetric{}
	for _, m := range metrics {
		result[m.ObjectID] = m
	}
	return result
}

func mapClassObjects(objects []models.DataObject) map[string][]models.DataObject {
	result := map[string][]models.DataObject{}
	for _, obj := range objects {
		label := obj.Label
		if label == "" {
			label = "unlabeled"
		}
		result[label] = append(result[label], obj)
	}
	return result
}
