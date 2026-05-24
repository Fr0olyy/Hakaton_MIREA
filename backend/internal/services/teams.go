package services

import (
	"context"
	"strings"

	"hakaton/backend/internal/models"

	"github.com/google/uuid"
)

type CreateTeamRequest struct {
	Name string `json:"name"`
}

type TeamWithMembers struct {
	models.Team
	Members []models.TeamMember `json:"members,omitempty"`
}

func (s *Service) CreateTeam(ctx context.Context, req CreateTeamRequest, createdBy uuid.UUID) (models.Team, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return models.Team{}, badRequest("team name is required")
	}
	team := models.Team{
		ID:        uuid.New(),
		Name:      req.Name,
		CreatedBy: createdBy,
	}
	team, err := s.repo.CreateTeam(ctx, team)
	if err != nil {
		return team, err
	}
	member := models.TeamMember{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: createdBy,
		Role:   "admin",
	}
	_, err = s.repo.AddTeamMember(ctx, member)
	return team, err
}

func (s *Service) ListTeams(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	return s.repo.ListTeamsByUser(ctx, userID)
}

func (s *Service) GetTeam(ctx context.Context, teamID uuid.UUID, userID uuid.UUID) (TeamWithMembers, error) {
	team, err := s.repo.GetTeam(ctx, teamID)
	if err != nil {
		return TeamWithMembers{}, err
	}
	members, err := s.repo.ListTeamMembers(ctx, teamID)
	if err != nil {
		return TeamWithMembers{}, err
	}
	return TeamWithMembers{Team: team, Members: members}, nil
}

func (s *Service) AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (models.TeamMember, error) {
	member := models.TeamMember{
		ID:     uuid.New(),
		TeamID: teamID,
		UserID: userID,
		Role:   "member",
	}
	return s.repo.AddTeamMember(ctx, member)
}

func (s *Service) RemoveTeamMember(ctx context.Context, teamID, userID uuid.UUID) error {
	return s.repo.RemoveTeamMember(ctx, teamID, userID)
}
