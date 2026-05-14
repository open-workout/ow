package repository_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/redis/go-redis/v9"
)

func TestListPublicExercises_CacheHit(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	exercises := []domain.ExerciseModel{
		{
			ExerciseID:       1,
			Name:             "Exercise1",
			ExerciseType:     "compound",
			PrimaryMuscle:    "legs",
			SecondaryMuscles: []string{"abs"},
			Description:      "Exercise1",
			UserID:           1,
			IsPrivate:        false,
			WeightDirection:  1,
		},
		{
			ExerciseID:       2,
			Name:             "Exercise2",
			ExerciseType:     "compound",
			PrimaryMuscle:    "legs",
			SecondaryMuscles: []string{"abs"},
			Description:      "Exercise2",
			UserID:           1,
			IsPrivate:        false,
			WeightDirection:  1,
		},
	}

	data, _ := json.Marshal(exercises)
	err = s.Set("public_exercises", string(data))
	if err != nil {
		t.Fatal(err)
	}

	mockRepo := repository.NewMockRepository()

	cache := repository.NewRedisCachedRepository(rdb, mockRepo)
	result, err := cache.ListPublicExercises(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if mockRepo.Called == true {
		t.Fatal("mockRepo was called, but shouldn't have")
	}

	if len(result) != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}

	// verify cache was written
	val, err := s.Get("public_exercises")
	if err != nil {
		t.Fatal(err)
	}

	if val == "" {
		t.Fatal("empty cache value")
	}
}

func TestUpdateExercise_UpdatesCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})

	updated := &domain.ExerciseModel{
		ExerciseID:      5,
		Name:            "Updated Name",
		ExerciseType:    "isolation",
		PrimaryMuscle:   "chest",
		UserID:          1,
		IsPrivate:       false,
		WeightDirection: 1,
	}

	repo := &repository.MockRepository{
		UpdateExerciseFunc: func(_ context.Context, _ int64, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
			return updated, nil
		},
	}

	cache := repository.NewRedisCachedRepository(rdb, repo)

	result, err := cache.UpdateExercise(context.Background(), 1, &domain.ExerciseModel{ExerciseID: 5})
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Updated Name" {
		t.Errorf("unexpected result: %+v", result)
	}

	val, err := s.Get("exercise:5")
	if err != nil {
		t.Fatal("expected cache to be updated after UpdateExercise")
	}
	if val == "" {
		t.Fatal("expected non-empty cache value")
	}
}

func TestDeleteExercise_InvalidatesCache(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})

	ex := domain.ExerciseModel{ExerciseID: 3, Name: "Deadlift"}
	data, _ := json.Marshal(ex)
	if err := s.Set("exercise:3", string(data)); err != nil {
		t.Fatal(err)
	}

	repo := &repository.MockRepository{
		DeleteExerciseFunc: func(_ context.Context, _ int64, id int64) error {
			return nil
		},
	}

	cache := repository.NewRedisCachedRepository(rdb, repo)

	if err := cache.DeleteExercise(context.Background(), 1, 3); err != nil {
		t.Fatal(err)
	}

	if _, err := s.Get("exercise:3"); err == nil {
		t.Fatal("expected cache key to be evicted after DeleteExercise")
	}
}

func TestGetExerciseById_CacheHit(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})

	ex := domain.ExerciseModel{
		ExerciseID:       42,
		Name:             "Deadlift",
		ExerciseType:     "compound",
		PrimaryMuscle:    "back",
		SecondaryMuscles: []string{"hamstrings"},
		Description:      "hip hinge",
		UserID:           1,
		IsPrivate:        false,
		WeightDirection:  1,
	}

	data, _ := json.Marshal(ex)
	if err := s.Set("exercise:42", string(data)); err != nil {
		t.Fatal(err)
	}

	mockRepo := repository.NewMockRepository()
	cache := repository.NewRedisCachedRepository(rdb, mockRepo)

	result, err := cache.GetExerciseById(context.Background(), 42, 0)
	if err != nil {
		t.Fatal(err)
	}

	if mockRepo.Called {
		t.Fatal("mockRepo was called, but shouldn't have been on cache hit")
	}

	if result.ExerciseID != 42 || result.Name != "Deadlift" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestGetExerciseById_CacheMiss(t *testing.T) {
	s, _ := miniredis.Run()
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})

	ex := &domain.ExerciseModel{
		ExerciseID:       7,
		Name:             "Squat",
		ExerciseType:     "compound",
		PrimaryMuscle:    "legs",
		SecondaryMuscles: []string{"glutes"},
		Description:      "knee dominant",
		UserID:           1,
		IsPrivate:        false,
		WeightDirection:  1,
	}

	repo := &repository.MockRepository{
		GetExerciseByIdFunc: func(_ context.Context, id int64, _ int64) (*domain.ExerciseModel, error) {
			return ex, nil
		},
	}

	cache := repository.NewRedisCachedRepository(rdb, repo)

	result, err := cache.GetExerciseById(context.Background(), 7, 0)
	if err != nil {
		t.Fatal(err)
	}

	if !repo.Called {
		t.Fatal("repo was not called, but should have been on cache miss")
	}

	if result.ExerciseID != 7 || result.Name != "Squat" {
		t.Errorf("unexpected result: %+v", result)
	}

	val, err := s.Get("exercise:7")
	if err != nil {
		t.Fatal("expected cache to be populated after miss")
	}
	if val == "" {
		t.Fatal("expected non-empty cache value")
	}
}

func TestListPublicExercises_CacheMiss(t *testing.T) {
	s, _ := miniredis.Run()
	defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	repo := &repository.MockRepository{
		Called: false,
		ListPublicExercisesFunc: func(_ context.Context) ([]domain.ExerciseModel, error) {
			ex1 := domain.ExerciseModel{
				ExerciseID:       1,
				Name:             "Exercise1",
				ExerciseType:     "compound",
				PrimaryMuscle:    "legs",
				SecondaryMuscles: []string{"abs"},
				Description:      "Exercise1",
				UserID:           1,
				WeightDirection:  1,
			}
			ex2 := domain.ExerciseModel{
				ExerciseID:       2,
				Name:             "Exercise2",
				ExerciseType:     "compound",
				PrimaryMuscle:    "legs",
				SecondaryMuscles: []string{"abs"},
				Description:      "Exercise2",
				UserID:           1,
				WeightDirection:  1,
			}
			return []domain.ExerciseModel{ex1, ex2}, nil
		},
	}

	cache := repository.NewRedisCachedRepository(rdb, repo)

	_, err := cache.ListPublicExercises(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if repo.Called == false {
		t.Fatal("repo was not called, but should have")
	}

	// verify cache was written
	val, err := s.Get("public_exercises")

	if err != nil {
		t.Fatal(err)
	}

	if val == "" {
		t.Fatal("expected set value, got empty cache value")
	}
	print(val)
}
