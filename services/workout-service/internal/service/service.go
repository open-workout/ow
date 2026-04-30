package service

import (
	"context"

	"github.com/open-workout/ow/services/workout-service/internal/domain"
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

func (s *Service) CreateSet(ctx context.Context, workoutSet *domain.SetModel) (*domain.SetModel, error) {
	w, err := s.repo.CreateSet(ctx, workoutSet)
	if err != nil {
		return &domain.SetModel{}, err
	}
	return w, nil
}
