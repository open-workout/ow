package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/storage"
	"github.com/open-workout/ow/services/exercise-service/internal/service"
	"github.com/open-workout/ow/shared/env"
	"github.com/redis/go-redis/v9"
)

func main() {

	ctx := context.Background()

	// PostgreSQL
	db, err := sql.Open("postgres", env.GetString("POSTGRES_DSN", ""))
	if err != nil {
		log.Fatalf("failed to open postgres connection: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("postgres ping failed: %v", err)
	}

	//Redis
	redisDSN := env.GetString("REDIS_DSN", "")
	redisOpts, err := redis.ParseURL(redisDSN)

	if err != nil {
		log.Fatalf("failed to parse redis DSN: %v", err)
	}

	redisClient := redis.NewClient(redisOpts)
	defer redisClient.Close()
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis ping failed: %v", err)
	}

	// Repository
	pgRepo := repository.NewSqlRepository(db)
	repo := repository.NewRedisCachedRepository(redisClient, pgRepo)

	// Media storage
	mediaRoot := env.GetString("MEDIA_BASE_ROOT", "/uploads")
	mediaDir := env.GetString("MEDIA_BASE_URL", "")
	mediaStorage, err := storage.NewLocalMediaStorage(mediaRoot, mediaDir)

	if err != nil {
		log.Fatalf("failed to create local media storage: %v", err)
	}

	svc := service.NewService(repo, mediaStorage)

	_ = svc

	log.Printf("exercise service initialised")

}
