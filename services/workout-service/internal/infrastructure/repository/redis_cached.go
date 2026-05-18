package repository

import (
	"context"
	"time"

	"github.com/open-workout/ow/services/workout-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCachedRepository struct {
	rdb  *redis.Client
	repo domain.WorkoutRepository
}

func NewRedisCachedRepository(rdb *redis.Client, repo domain.WorkoutRepository) *RedisCachedRepository {
	return &RedisCachedRepository{
		rdb:  rdb,
		repo: repo,
	}
}

func (r *RedisCachedRepository) CreateWorkout(ctx context.Context, workout *domain.WorkoutModel) (*domain.WorkoutModel, error) {
	return r.repo.CreateWorkout(ctx, workout)
}

func (r *RedisCachedRepository) SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error {
	return r.repo.SetWorkoutFinishTime(ctx, workoutId, finishedAt)
}

func (r *RedisCachedRepository) GetWorkoutById(ctx context.Context, workoutId int64) (*domain.WorkoutModel, error) {
	return r.repo.GetWorkoutById(ctx, workoutId)
}

func (r *RedisCachedRepository) DeleteWorkout(ctx context.Context, workoutId int64) error {
	return r.repo.DeleteWorkout(ctx, workoutId)
}

func (r *RedisCachedRepository) DeleteWorkoutsByUserID(ctx context.Context, userId string) error {
	return r.repo.DeleteWorkoutsByUserID(ctx, userId)
}

func (r *RedisCachedRepository) CreateSet(ctx context.Context, set *domain.SetModel) (*domain.SetModel, error) {
	return r.repo.CreateSet(ctx, set)
}

func (r *RedisCachedRepository) UpdateSet(ctx context.Context, userId string, set *domain.SetModel) (*domain.SetModel, error) {
	return r.repo.UpdateSet(ctx, userId, set)
}

func (r *RedisCachedRepository) DeleteSet(ctx context.Context, userId string, setId int64) error {
	return r.repo.DeleteSet(ctx, userId, setId)
}

func (r *RedisCachedRepository) GetSetsByWorkoutID(ctx context.Context, workoutId int64, userId string) ([]*domain.SetModel, error) {
	return r.repo.GetSetsByWorkoutID(ctx, workoutId, userId)
}

func (r *RedisCachedRepository) GetLastTimeMaxSet(ctx context.Context, userId string, exerciseId int64) (*domain.SetModel, error) {
	return r.repo.GetLastTimeMaxSet(ctx, userId, exerciseId)
}
