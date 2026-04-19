package service

import (
	"context"

	"github.com/open-workout/ow/internal/domain"
)

type Service struct {
	repo domain.WorkoutRepository
}

func NewService(repo domain.WorkoutRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateWorkout(ctx context.Context, workout *domain.WorkoutModel) (*domain.WorkoutModel, error) {
	w, err := s.repo.CreateWorkout(ctx, workout)
	if err != nil {
		return &domain.WorkoutModel{}, err
	}
	return w, nil
}
