package service_test

import (
	"context"
	"testing"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/storage"
	"github.com/open-workout/ow/services/exercise-service/internal/service"
)

func TestCreateExercise_Success(t *testing.T) {
	input := &domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           1,
		IsPrivate:        false,
	}
	want := &domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           1,
		IsPrivate:        false,
	}

	repo := &repository.MockRepository{
		CreateExerciseFunc: func(_ context.Context, e *domain.ExerciseModel) (*domain.ExerciseModel, error) {
			return want, nil
		},
	}

	svc := service.NewService(repo, storage.MockMediaStorage{})
	got, err := svc.CreateExercise(context.Background(), input)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != want.Name {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

//TODO: Test  all methods of the service
