package services

import (
	"context"

	"hakaton/backend/internal/agentsvc"
	"hakaton/backend/internal/mlclient"
	"hakaton/backend/internal/repositories"
	filestorage "hakaton/backend/internal/storage"
)

type Service struct {
	repo      *repositories.Repository
	storage   *filestorage.Storage
	ml        *mlclient.Client
	agentsvc  *agentsvc.Client
	jwtSecret string
}

func New(repo *repositories.Repository, storage *filestorage.Storage, ml *mlclient.Client, agentClient *agentsvc.Client, jwtSecret string) *Service {
	return &Service{repo: repo, storage: storage, ml: ml, agentsvc: agentClient, jwtSecret: jwtSecret}
}

func (s *Service) Health(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		return err
	}
	return s.storage.Health()
}
