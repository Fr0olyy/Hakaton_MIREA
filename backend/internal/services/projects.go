package services

import (
	"context"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func (s *Service) CreateProject(ctx context.Context, req CreateProjectRequest) (models.Project, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Modality = strings.ToLower(strings.TrimSpace(req.Modality))
	req.TaskType = strings.ToLower(strings.TrimSpace(req.TaskType))
	if req.Name == "" || req.Modality == "" || req.TaskType == "" {
		return models.Project{}, badRequest("name, modality and task_type are required")
	}
	if req.Modality != "image" || req.TaskType != "classification" {
		return models.Project{}, badRequest("Level 1 supports only modality=image and task_type=classification")
	}

	project := models.Project{
		ID:       uuid.New(),
		Name:     req.Name,
		Modality: req.Modality,
		TaskType: req.TaskType,
		Classes:  normalizeClasses(req.Classes),
	}
	created, err := s.repo.CreateProject(ctx, project)
	if err != nil {
		return created, err
	}
	return created, s.storage.EnsureProjectDirs(created.ID)
}

func (s *Service) ListProjects(ctx context.Context) ([]models.Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *Service) GetProject(ctx context.Context, id uuid.UUID) (models.Project, error) {
	return s.repo.GetProject(ctx, id)
}

func normalizeClasses(classes []string) []string {
	seen := map[string]struct{}{}
	normalized := make([]string, 0, len(classes))
	for _, class := range classes {
		class = strings.TrimSpace(class)
		if class == "" {
			continue
		}
		if _, ok := seen[class]; ok {
			continue
		}
		seen[class] = struct{}{}
		normalized = append(normalized, class)
	}
	return normalized
}
