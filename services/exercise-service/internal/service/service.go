package service

import (
	"context"
	"sort"

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

func (s *Service) AddExerciseMedia(ctx context.Context, exerciseID int64, callerUserID string, media *domain.ExerciseMedia, file *domain.ExerciseMediaUpload) error {
	ex, err := s.repo.GetExerciseById(ctx, exerciseID, callerUserID)
	if err != nil {
		return err
	}
	if callerUserID != "" && ex.UserID != callerUserID {
		return domain.ErrForbidden
	}

	url, err := s.mediaStorage.Upload(ctx, file)
	if err != nil {
		return err
	}

	media.URL = url
	return s.repo.AddExerciseMedia(ctx, exerciseID, media)
}

func (s *Service) GetExerciseMedia(ctx context.Context, exerciseID int64, callerUserID string) ([]domain.ExerciseMedia, error) {
	return s.repo.GetExerciseMedia(ctx, exerciseID, callerUserID)
}

func (s *Service) ListExercises(ctx context.Context, userID string) ([]domain.ExerciseModel, error) {
	return s.repo.ListExercises(ctx, userID)
}

func (s *Service) GetExerciseById(ctx context.Context, id int64, callerUserID string) (*domain.ExerciseModel, error) {
	return s.repo.GetExerciseById(ctx, id, callerUserID)
}

func (s *Service) UpdateExercise(ctx context.Context, callerUserID string, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	return s.repo.UpdateExercise(ctx, callerUserID, exercise)
}

func (s *Service) DeleteExercise(ctx context.Context, callerUserID string, id int64) error {
	return s.repo.DeleteExercise(ctx, callerUserID, id)
}

func (s *Service) GetTopExercises(
	ctx context.Context,
	state domain.MuscleState,
	limit int,
) ([]domain.ExerciseModel, error) {

	if limit == -1 {
		limit = 10
	}

	exercises, err := s.ListExercises(ctx, state.UserID)
	if err != nil {
		return nil, err
	}

	type scoredExercise struct {
		ex    domain.ExerciseModel
		score float64
	}

	var scored []scoredExercise

	for _, exercise := range exercises {
		score := scoreExercise(&exercise, state)

		if score == 0 {
			continue
		}

		scored = append(scored, scoredExercise{
			ex:    exercise,
			score: score,
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	if len(scored) < limit {
		limit = len(scored)
	}

	result := make([]domain.ExerciseModel, limit)
	for i := 0; i < limit; i++ {
		result[i] = scored[i].ex
	}

	return result, nil
}
