package repository

import (
	"github.com/open-workout/ow/services/exercise-service/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisCachedRepository struct {
	redis *redis.Client
	repo  domain.ExerciseRepository
}
