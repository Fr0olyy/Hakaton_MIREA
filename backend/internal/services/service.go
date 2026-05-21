package services

import (
	"context"

	"hakaton/backend/internal/mlclient"
	"hakaton/backend/internal/repositories"
	filestorage "hakaton/backend/internal/storage"
)

type Service struct {
	repo    *repositories.Repository
	storage *filestorage.Storage
	ml      *mlclient.Client
}

func New(repo *repositories.Repository, storage *filestorage.Storage, ml *mlclient.Client) *Service {
	return &Service{repo: repo, storage: storage, ml: ml}
}

func (s *Service) Health(ctx context.Context) error {
	if err := s.repo.Ping(ctx); err != nil {
		return err
	}
	return s.storage.Health()
}
