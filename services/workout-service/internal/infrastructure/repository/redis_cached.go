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

func (r *RedisCachedRepository) DeleteWorkoutsByUserID(ctx context.Context, userId int64) error {
	return r.repo.DeleteWorkoutsByUserID(ctx, userId)
}

func (r *RedisCachedRepository) CreateSet(ctx context.Context, set *domain.SetModel) (*domain.SetModel, error) {
	return r.repo.CreateSet(ctx, set)
}

func (r *RedisCachedRepository) DeleteSet(ctx context.Context, workoutId int64, exerciseId int64) error {
	return r.repo.DeleteSet(ctx, workoutId, exerciseId)
}

func (r *RedisCachedRepository) GetLastTimeMaxSet(ctx context.Context, userId int64, exerciseId int64) (*domain.SetModel, error) {
	return r.repo.GetLastTimeMaxSet(ctx, userId, exerciseId)
}
