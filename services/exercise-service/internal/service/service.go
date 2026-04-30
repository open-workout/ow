package service

import (
	"context"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type Service struct {
	repo         domain.ExerciseRepository
	mediaStorage domain.MediaStorage
}

func NewService(repo domain.ExerciseRepository, ms domain.MediaStorage) *Service {
	return &Service{
		repo:         repo,
		mediaStorage: ms,
	}
}

func (s *Service) CreateExercise(ctx context.Context, exercise *domain.ExerciseModel) (model *domain.ExerciseModel, err error) {
	return s.repo.CreateExercise(ctx, exercise)
}

func (s *Service) UpdateExercise(ctx context.Context, exercise *domain.ExerciseModel) (model *domain.ExerciseModel, err error) {
	return s.repo.UpdateExercise(ctx, exercise.ExerciseID, exercise)
}

func (s *Service) AddExerciseMedia(ctx context.Context, exerciseId int64, media *domain.ExerciseMedia) error {
	err := s.repo.AddExerciseMedia(ctx, exerciseId, media)
	if err != nil {
		return err
	}

	//TODO: add logic for storing media on the server
	// by calling mediaStorage.Upload(something)

	return nil
}
