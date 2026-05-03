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
			}
			ex2 := domain.ExerciseModel{
				ExerciseID:       2,
				Name:             "Exercise2",
				ExerciseType:     "compound",
				PrimaryMuscle:    "legs",
				SecondaryMuscles: []string{"abs"},
				Description:      "Exercise2",
				UserID:           1,
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
