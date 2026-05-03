package service_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/storage"
	"github.com/open-workout/ow/services/exercise-service/internal/service"
)

var errRepo = errors.New("repo error")
var errStorage = errors.New("storage error")

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

func TestCreateExercise_RepoError(t *testing.T) {
	mockRepo := &repository.MockRepository{
		CreateExerciseFunc: func(_ context.Context, _ *domain.ExerciseModel) (*domain.ExerciseModel, error) {
			return nil, errRepo
		},
	}

	ex := domain.ExerciseModel{
		Name:             "Pull Up",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"biceps"},
		Description:      "pull movement",
		UserID:           1,
		IsPrivate:        false,
	}

	svc := service.NewService(mockRepo, storage.MockMediaStorage{})
	_, err := svc.CreateExercise(context.Background(), &ex)

	if !errors.Is(err, errRepo) {
		t.Errorf("got %v, want %v", err, errRepo)
	}

}

func TestAddExerciseMedia_Success(t *testing.T) {
	uploadCalled := false

	mockRepo := &repository.MockRepository{
		AddExerciseMediaFunc: func(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error {
			return nil
		},
	}

	mockMedia := &storage.MockMediaStorage{
		UploadFunc: func(_ context.Context, file *domain.ExerciseMediaUpload) (string, error) {
			uploadCalled = true
			return "https://example.com/media/1.jpg", nil
		},
	}

	svc := service.NewService(mockRepo, mockMedia)
	err := svc.AddExerciseMedia(
		context.Background(),
		1,
		&domain.ExerciseMedia{ExerciseID: 1, UserID: 1},
		&domain.ExerciseMediaUpload{Filename: "1.jpg", File: strings.NewReader("data")},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uploadCalled {
		t.Error("expected upload to be called")
	}
}

func TestAddExerciseMedia_RepoError_SkipsUpload(t *testing.T) {
	uploadCalled := false

	mockRepo := &repository.MockRepository{
		AddExerciseMediaFunc: func(_ context.Context, _ int64, _ *domain.ExerciseMedia) error {
			return errRepo
		},
	}

	mockMedia := &storage.MockMediaStorage{
		UploadFunc: func(_ context.Context, file *domain.ExerciseMediaUpload) (string, error) {
			uploadCalled = true
			return "", nil
		},
	}

	svc := service.NewService(mockRepo, mockMedia)

	err := svc.AddExerciseMedia(context.Background(), 1, &domain.ExerciseMedia{}, &domain.ExerciseMediaUpload{})
	if !errors.Is(err, errRepo) {
		t.Errorf("expected repo error, got %v", err)
	}
	if uploadCalled {
		t.Error("Upload should not be called when repo fails")
	}
}

func TestAddExerciseMedia_StorageError(t *testing.T) {

	mockRepo := &repository.MockRepository{
		AddExerciseMediaFunc: func(_ context.Context, _ int64, _ *domain.ExerciseMedia) error {
			return nil
		},
	}

	mockMedia := &storage.MockMediaStorage{
		UploadFunc: func(_ context.Context, file *domain.ExerciseMediaUpload) (string, error) {
			return "", errStorage
		},
	}

	svc := service.NewService(mockRepo, mockMedia)
	err := svc.AddExerciseMedia(context.Background(), 1, &domain.ExerciseMedia{}, &domain.ExerciseMediaUpload{})
	if !errors.Is(err, errStorage) {
		t.Errorf("expected storage error, got %v", err)
	}
}

func TestListExercises_Success(t *testing.T) {
	want := []domain.ExerciseModel{
		{
			Name:             "Pull Up",
			ExerciseType:     "compound",
			PrimaryMuscle:    "back",
			SecondaryMuscles: []string{"biceps"},
			Description:      "pull movement",
			UserID:           1,
			IsPrivate:        false,
		},
		{
			Name:             "Deadlift",
			ExerciseType:     "compound",
			PrimaryMuscle:    "back",
			SecondaryMuscles: []string{"legs"},
			Description:      "pull movement",
			UserID:           1,
			IsPrivate:        false,
		},
	}

	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
			if userID != 1 {
				t.Errorf("got user_id %d, want 1", userID)
			}
			return want, nil
		},
	}

	mockMedia := &storage.MockMediaStorage{
		UploadFunc: func(_ context.Context, file *domain.ExerciseMediaUpload) (string, error) {
			return "", errStorage
		},
	}
	svc := service.NewService(mockRepo, mockMedia)

	got, err := svc.ListExercises(context.Background(), 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func buildExercise(id int64, primary string, secondary ...string) domain.ExerciseModel {
	return domain.ExerciseModel{
		ExerciseID:       id,
		Name:             primary + "-exercise",
		PrimaryMuscle:    primary,
		SecondaryMuscles: secondary,
		Description:      "exercise",
		UserID:           1,
	}
}
func muscleState(userID int64, muscles map[string]float64) domain.MuscleState {
	return domain.MuscleState{
		UserID:  userID,
		Muscles: muscles,
	}
}

func TestGetTopExercises_DefaultLimitWhenMinusOne(t *testing.T) {

	exercises := make([]domain.ExerciseModel, 15)
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
			return exercises, nil
		},
	}

	mockMedia := &storage.MockMediaStorage{}
	svc := service.NewService(mockRepo, mockMedia)
	// 15 exercises that all score > 0; with limit=-1 only 10 should be returned.

	for i := range exercises {
		exercises[i] = buildExercise(int64(i+1), "chest")
	}

	state := muscleState(1, map[string]float64{"chest": 0.8})
	got, err := svc.GetTopExercises(context.Background(), state, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 10 {
		t.Errorf("got %d exercises, want 10", len(got))
	}
}

func TestGetTopExercises_LimitCappedToAvailable(t *testing.T) {
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(_ context.Context, _ int64) ([]domain.ExerciseModel, error) {
			return []domain.ExerciseModel{
				buildExercise(1, "chest"),
				buildExercise(2, "chest"),
			}, nil
		},
	}
	svc := service.NewService(mockRepo, &storage.MockMediaStorage{})

	got, err := svc.GetTopExercises(context.Background(), muscleState(1, map[string]float64{"chest": 0.5}), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("got %d exercises, want 2", len(got))
	}
}

func TestGetTopExercises_ZeroScoreExercisesExcluded(t *testing.T) {
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(_ context.Context, _ int64) ([]domain.ExerciseModel, error) {
			return []domain.ExerciseModel{
				buildExercise(1, "chest"), // matches state
				buildExercise(2, "back"),  // no match → score 0, excluded
			}, nil
		},
	}
	svc := service.NewService(mockRepo, &storage.MockMediaStorage{})

	got, err := svc.GetTopExercises(context.Background(), muscleState(1, map[string]float64{"chest": 1.0}), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d exercises, want 1", len(got))
	}
	if got[0].PrimaryMuscle != "chest" {
		t.Errorf("expected chest exercise, got %s", got[0].PrimaryMuscle)
	}
}

func TestGetTopExercises_SortedByScoreDescending(t *testing.T) {
	// ex1: primary only        → score = (0.9 * 1.2) / 1.2        = 0.9
	// ex2: primary + secondary → score = (0.9*1.2 + 0.8*1) / 2.2 ≈ 0.854
	// ex1 scores higher because adding a lower-valued secondary muscle dilutes the average.
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(_ context.Context, _ int64) ([]domain.ExerciseModel, error) {
			return []domain.ExerciseModel{
				buildExercise(1, "chest"),
				buildExercise(2, "chest", "triceps"),
			}, nil
		},
	}
	svc := service.NewService(mockRepo, &storage.MockMediaStorage{})

	got, err := svc.GetTopExercises(context.Background(), muscleState(1, map[string]float64{"chest": 0.9, "triceps": 0.8}), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got[0].ExerciseID != 1 {
		t.Errorf("expected ex1 first (higher score), got ExerciseID=%d", got[0].ExerciseID)
	}
}

func TestGetTopExercises_RepoError(t *testing.T) {
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(_ context.Context, _ int64) ([]domain.ExerciseModel, error) {
			return nil, errRepo
		},
	}
	svc := service.NewService(mockRepo, &storage.MockMediaStorage{})

	_, err := svc.GetTopExercises(context.Background(), domain.MuscleState{UserID: 1}, 5)
	if !errors.Is(err, errRepo) {
		t.Errorf("expected repo error, got %v", err)
	}
}

func TestGetTopExercises_EmptyList(t *testing.T) {
	mockRepo := &repository.MockRepository{
		ListExercisesFunc: func(_ context.Context, _ int64) ([]domain.ExerciseModel, error) {
			return []domain.ExerciseModel{}, nil
		},
	}
	svc := service.NewService(mockRepo, &storage.MockMediaStorage{})

	got, err := svc.GetTopExercises(context.Background(), muscleState(1, map[string]float64{"chest": 1}), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}
