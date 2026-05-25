package services

import (
	"context"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

func (s *Service) CreateProject(ctx context.Context, req CreateProjectRequest, creatorID uuid.UUID) (models.Project, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Modality = strings.ToLower(strings.TrimSpace(req.Modality))
	req.TaskType = strings.ToLower(strings.TrimSpace(req.TaskType))
	if req.Name == "" || req.Modality == "" || req.TaskType == "" {
		return models.Project{}, badRequest("name, modality and task_type are required")
	}
	if !isSupportedProjectType(req.Modality, req.TaskType) {
		return models.Project{}, badRequest("unsupported modality/task_type combination")
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
	if err := s.storage.EnsureProjectDirs(created.ID); err != nil {
		return created, err
	}
	if creatorID != uuid.Nil {
		member := models.ProjectMember{
			ID:        uuid.New(),
			ProjectID: created.ID,
			UserID:    creatorID,
			Role:      models.ProjectRoleAdmin,
		}
		if _, err := s.repo.AddProjectMember(ctx, member); err != nil {
			return created, err
		}
	}
	return created, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]models.Project, error) {
	return s.repo.ListProjects(ctx)
}

func (s *Service) GetProject(ctx context.Context, projectID uuid.UUID) (models.Project, error) {
	return s.repo.GetProject(ctx, projectID)
}

func isSupportedProjectType(modality, taskType string) bool {
	switch modality {
	case "image", "image_classification":
		return taskType == "classification"
	case "tabular_classification":
		return taskType == "classification"
	case "image_detection_yolo":
		return taskType == "detection" || taskType == "object_detection"
	default:
		return false
	}
}

func isImageClassificationProject(project models.Project) bool {
	return (project.Modality == "image" || project.Modality == "image_classification") && project.TaskType == "classification"
}

func isImageLikeProject(project models.Project) bool {
	return project.Modality == "image" || project.Modality == "image_classification" || project.Modality == "image_detection_yolo"
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
