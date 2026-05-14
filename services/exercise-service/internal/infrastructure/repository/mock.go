package repository

import (
	"context"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type MockRepository struct {
	Called bool

	CreateExerciseFunc   func(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error)
	AddExerciseMediaFunc func(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error
	GetExerciseMediaFunc func(ctx context.Context, exerciseID int64, callerUserID int64) ([]domain.ExerciseMedia, error)

	ListExercisesFunc       func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error)
	ListPublicExercisesFunc func(ctx context.Context) ([]domain.ExerciseModel, error)
	ListUserExercisesFunc   func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error)

	GetExerciseByIdFunc func(ctx context.Context, id int64, callerUserID int64) (*domain.ExerciseModel, error)

	UpdateExerciseFunc func(ctx context.Context, callerUserID int64, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error)
	DeleteExerciseFunc func(ctx context.Context, callerUserID int64, id int64) error
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

func (m *MockRepository) GetExerciseMedia(ctx context.Context, exerciseID int64, callerUserID int64) ([]domain.ExerciseMedia, error) {
	m.Called = true
	return m.GetExerciseMediaFunc(ctx, exerciseID, callerUserID)
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

func (m *MockRepository) GetExerciseById(ctx context.Context, id int64, callerUserID int64) (*domain.ExerciseModel, error) {
	m.Called = true
	return m.GetExerciseByIdFunc(ctx, id, callerUserID)
}

func (m *MockRepository) UpdateExercise(ctx context.Context, callerUserID int64, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	m.Called = true
	return m.UpdateExerciseFunc(ctx, callerUserID, exercise)
}

func (m *MockRepository) DeleteExercise(ctx context.Context, callerUserID int64, id int64) error {
	m.Called = true
	return m.DeleteExerciseFunc(ctx, callerUserID, id)
}
