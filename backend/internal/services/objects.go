package services

import (
	"context"
	"math"
	"strings"

	"hakaton/backend/internal/models"
	"hakaton/backend/internal/repositories"

	"github.com/google/uuid"
)

type ListObjectsResult struct {
	Objects    []models.DataObject `json:"objects"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

type ObjectDetail struct {
	models.DataObject
	Metrics     *models.ObjectMetric   `json:"metrics,omitempty"`
	Actions     []models.ObjectAction  `json:"actions,omitempty"`
	Comments    []models.ObjectComment `json:"comments,omitempty"`
	Assignments []models.ReviewAssignment `json:"assignments,omitempty"`
}

func (s *Service) ListObjects(ctx context.Context, projectID uuid.UUID, page, perPage int, status, label, reason, recommendation, sortBy, sortOrder, search string, scoreMin, scoreMax float64) (ListObjectsResult, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	version, err := s.repo.LatestDatasetVersion(ctx, projectID)
	if err != nil {
		return ListObjectsResult{}, err
	}

	sortOrder = strings.ToUpper(sortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	filter := repositories.ListObjectsFilter{
		DatasetVersionID: version.ID,
		Status:           status,
		Label:            label,
		Recommendation:   reason,
		ScoreMin:         scoreMin,
		ScoreMax:         scoreMax,
		Search:           search,
		SortBy:           sortBy,
		SortOrder:        sortOrder,
		Offset:           (page - 1) * perPage,
		Limit:            perPage,
	}

	objects, total, err := s.repo.ListObjectsPaginated(ctx, filter)
	if err != nil {
		return ListObjectsResult{}, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return ListObjectsResult{
		Objects:    objects,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *Service) GetObjectDetail(ctx context.Context, projectID, objectID uuid.UUID) (ObjectDetail, error) {
	object, err := s.repo.GetObject(ctx, projectID, objectID)
	if err != nil {
		return ObjectDetail{}, err
	}

	actions, _ := s.repo.ListObjectActions(ctx, projectID, objectID)
	comments, _ := s.repo.ListObjectComments(ctx, projectID, objectID)

	return ObjectDetail{
		DataObject: object,
		Metrics:    object.Metrics,
		Actions:    actions,
		Comments:   comments,
	}, nil
}

func (s *Service) PerformObjectAction(ctx context.Context, projectID, objectID, userID uuid.UUID, action, oldValue, newValue, comment string) (models.ObjectAction, error) {
	validActions := map[string]bool{
		"approve": true, "exclude": true, "relabel": true,
		"mark_duplicate": true, "send_to_expert": true, "add_to_next_train": true,
		"expert_approve": true, "expert_reject": true,
	}
	if !validActions[action] {
		return models.ObjectAction{}, badRequest("invalid action: " + action)
	}

	a := models.ObjectAction{
		ID:        uuid.New(),
		ProjectID: projectID,
		ObjectID:  objectID,
		UserID:    userID,
		Action:    action,
		OldValue:  oldValue,
		NewValue:  newValue,
		Comment:   comment,
	}
	return s.repo.CreateObjectAction(ctx, a)
}

func (s *Service) AddObjectComment(ctx context.Context, projectID, objectID, userID uuid.UUID, text string) (models.ObjectComment, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return models.ObjectComment{}, badRequest("comment text is required")
	}
	c := models.ObjectComment{
		ID:        uuid.New(),
		ProjectID: projectID,
		ObjectID:  objectID,
		UserID:    userID,
		Text:      text,
	}
	return s.repo.CreateObjectComment(ctx, c)
}

func (s *Service) ListObjectComments(ctx context.Context, projectID, objectID uuid.UUID) ([]models.ObjectComment, error) {
	return s.repo.ListObjectComments(ctx, projectID, objectID)
}

func (s *Service) AssignReview(ctx context.Context, projectID, objectID, userID uuid.UUID) (models.ReviewAssignment, error) {
	a := models.ReviewAssignment{
		ID:        uuid.New(),
		ProjectID: projectID,
		ObjectID:  objectID,
		UserID:    userID,
		Status:    "assigned",
	}
	return s.repo.CreateReviewAssignment(ctx, a)
}

func (s *Service) AssignedToMe(ctx context.Context, projectID, userID uuid.UUID) ([]models.ReviewAssignment, error) {
	return s.repo.ListReviewAssignmentsByUser(ctx, projectID, userID)
}
