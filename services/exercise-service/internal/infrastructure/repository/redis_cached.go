package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCachedRepository struct {
	redis *redis.Client
	repo  domain.ExerciseRepository
}

func NewRedisCachedRepository(redis *redis.Client, repo domain.ExerciseRepository) *RedisCachedRepository {
	return &RedisCachedRepository{
		redis: redis,
		repo:  repo,
	}
}

func (r *RedisCachedRepository) CreateExercise(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	return r.repo.CreateExercise(ctx, exercise)
}

func (r *RedisCachedRepository) AddExerciseMedia(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error {
	return r.repo.AddExerciseMedia(ctx, exerciseID, media)
}

func (r *RedisCachedRepository) ListPublicExercises(ctx context.Context) ([]domain.ExerciseModel, error) {
	const cacheKey = "public_exercises"

	cached, err := r.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		var exercises []domain.ExerciseModel
		if json.Unmarshal([]byte(cached), &exercises) == nil {
			return exercises, nil
		}
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	exercises, err := r.repo.ListPublicExercises(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(exercises)
	if err == nil {
		_ = r.redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return exercises, nil
}

func (r *RedisCachedRepository) ListUserExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
	return r.repo.ListUserExercises(ctx, userID)
}

func (r *RedisCachedRepository) GetExerciseById(ctx context.Context, id int64) (*domain.ExerciseModel, error) {
	cacheKey := fmt.Sprintf("exercise:%d", id)

	cached, err := r.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		var ex domain.ExerciseModel
		if json.Unmarshal([]byte(cached), &ex) == nil {
			return &ex, nil
		}
	}

	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	ex, err := r.repo.GetExerciseById(ctx, id)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(ex)
	if err == nil {
		_ = r.redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return ex, nil
}

func (r *RedisCachedRepository) UpdateExercise(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	updated, err := r.repo.UpdateExercise(ctx, exercise)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("exercise:%d", exercise.ExerciseID)
	data, err := json.Marshal(updated)
	if err == nil {
		_ = r.redis.Set(ctx, cacheKey, data, 10*time.Minute).Err()
	}

	return updated, nil
}

func (r *RedisCachedRepository) DeleteExercise(ctx context.Context, id int64) error {
	if err := r.repo.DeleteExercise(ctx, id); err != nil {
		return err
	}

	_ = r.redis.Del(ctx, fmt.Sprintf("exercise:%d", id)).Err()

	return nil
}

func (r *RedisCachedRepository) ListExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
	public, err := r.ListPublicExercises(ctx)
	if err != nil {
		return nil, err
	}

	private, err := r.ListUserExercises(ctx, userID)
	if err != nil {
		return nil, err
	}

	return append(public, private...), nil
}
