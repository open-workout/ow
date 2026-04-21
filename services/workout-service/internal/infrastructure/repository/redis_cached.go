package repository

import (
	"context"

	"github.com/open-workout/ow/internal/domain"
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
