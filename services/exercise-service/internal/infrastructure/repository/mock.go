package repository

import (
	"context"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type MockRepository struct {
	Called bool

	CreateExerciseFunc   func(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error)
	AddExerciseMediaFunc func(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error

	ListExercisesFunc       func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error)
	ListPublicExercisesFunc func(ctx context.Context) ([]domain.ExerciseModel, error)
	ListUserExercisesFunc   func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error)
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Called: false,
	}
}

func (m *MockRepository) CreateExercise(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	m.Called = true
	return m.CreateExerciseFunc(ctx, exercise)
}

func (m *MockRepository) AddExerciseMedia(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error {
	m.Called = true
	return m.AddExerciseMediaFunc(ctx, exerciseID, media)
}

func (m *MockRepository) ListPublicExercises(ctx context.Context) ([]domain.ExerciseModel, error) {
	m.Called = true
	return m.ListPublicExercisesFunc(ctx)
}

func (m *MockRepository) ListUserExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
	m.Called = true
	return m.ListUserExercisesFunc(ctx, userID)
}

func (m *MockRepository) ListExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
	m.Called = true
	return m.ListExercisesFunc(ctx, userID)
}
