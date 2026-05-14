package service

import (
	"context"
	"time"

	"github.com/open-workout/ow/services/workout-service/internal/domain"
)

type Service struct {
	repo domain.WorkoutRepository
}

func NewService(repo domain.WorkoutRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateWorkout(ctx context.Context, workout *domain.WorkoutModel) (*domain.WorkoutModel, error) {
	return s.repo.CreateWorkout(ctx, workout)
}

func (s *Service) SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error {
	return s.repo.SetWorkoutFinishTime(ctx, workoutId, finishedAt)
}

func (s *Service) GetWorkoutById(ctx context.Context, workoutId int64) (*domain.WorkoutModel, error) {
	return s.repo.GetWorkoutById(ctx, workoutId)
}

func (s *Service) DeleteWorkout(ctx context.Context, workoutId int64) error {
	return s.repo.DeleteWorkout(ctx, workoutId)
}

func (s *Service) DeleteWorkoutsByUserID(ctx context.Context, userId int64) error {
	return s.repo.DeleteWorkoutsByUserID(ctx, userId)
}

func (s *Service) CreateSet(ctx context.Context, workoutSet *domain.SetModel) (*domain.SetModel, error) {
	return s.repo.CreateSet(ctx, workoutSet)
}

func (s *Service) DeleteSet(ctx context.Context, workoutId int64, exerciseId int64) error {
	return s.repo.DeleteSet(ctx, workoutId, exerciseId)
}

func (s *Service) GetLastTimeMaxSet(ctx context.Context, userId int64, exerciseId int64) (*domain.SetModel, error) {
	return s.repo.GetLastTimeMaxSet(ctx, userId, exerciseId)
}
