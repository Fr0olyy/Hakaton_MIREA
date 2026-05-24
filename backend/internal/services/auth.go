package services

import (
	"context"
	"strings"

	"hakaton/backend/internal/auth"
	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return AuthResponse{}, badRequest("email, password and name are required")
	}
	if len(req.Password) < 6 {
		return AuthResponse{}, badRequest("password must be at least 6 characters")
	}

	if s.repo != nil {
		existing, err := s.repo.GetUserByEmail(ctx, req.Email)
		if err == nil && existing.ID != uuid.Nil {
			return AuthResponse{}, badRequest("email already registered")
		}
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return AuthResponse{}, err
	}

	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hash,
		Name:         req.Name,
		Role:         models.RoleAnnotator,
	}
	if s.repo != nil {
		user, err = s.repo.CreateUser(ctx, user)
		if err != nil {
			return AuthResponse{}, err
		}
	}

	token, err := auth.GenerateToken(user, s.jwtSecret)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return AuthResponse{}, badRequest("email and password are required")
	}

	if s.repo == nil {
		return AuthResponse{}, badRequest("service not initialized")
	}

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return AuthResponse{}, badRequest("invalid email or password")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		return AuthResponse{}, badRequest("invalid email or password")
	}

	token, err := auth.GenerateToken(user, s.jwtSecret)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{Token: token, User: user}, nil
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (models.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}

func (s *Service) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]models.ProjectMember, error) {
	return s.repo.ListProjectMembers(ctx, projectID)
}

func (s *Service) RemoveProjectMember(ctx context.Context, projectID, userID uuid.UUID) error {
	return s.repo.RemoveProjectMember(ctx, projectID, userID)
}

func (s *Service) UpdateProjectRole(ctx context.Context, projectID, userID uuid.UUID, role models.ProjectMemberRole) error {
	if role != models.ProjectRoleAdmin &&
		role != models.ProjectRoleMLEngineer &&
		role != models.ProjectRoleAnnotator &&
		role != models.ProjectRoleDomainExpert &&
		role != models.ProjectRoleDataAnalyst {
		return badRequest("invalid role: " + string(role))
	}
	member := models.ProjectMember{
		ID:        uuid.New(),
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	_, err := s.repo.AddProjectMember(ctx, member)
	// allow re-adding with different role
	if err != nil && strings.Contains(err.Error(), "unique") {
		// remove and re-add
		_ = s.repo.RemoveProjectMember(ctx, projectID, userID)
		_, err = s.repo.AddProjectMember(ctx, member)
	}
	return err
}
