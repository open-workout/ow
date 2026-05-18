package service

import (
	"context"

	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type Service struct {
	repo domain.UserRepository
}

func NewService(repo domain.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return s.repo.UpdateUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *Service) UpdateSplit(ctx context.Context, userID string, split domain.Split) (*domain.User, error) {
	return s.repo.UpdateSplit(ctx, userID, split)
}
