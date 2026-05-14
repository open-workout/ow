package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/exercise-service/internal/infrastructure/storage"
	"github.com/open-workout/ow/services/exercise-service/internal/service"
	handlers "github.com/open-workout/ow/services/exercise-service/internal/transport/http/handlers"
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

	h := handlers.NewExerciseHandler(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /exercises", h.CreateExercise)
	mux.HandleFunc("GET /exercises", h.ListExercises)
	mux.HandleFunc("POST /exercises/recommendations", h.GetTopExercises)
	mux.HandleFunc("GET /exercises/{id}", h.GetExerciseById)
	mux.HandleFunc("PUT /exercises/{id}", h.UpdateExercise)
	mux.HandleFunc("DELETE /exercises/{id}", h.DeleteExercise)
	mux.HandleFunc("POST /exercises/{id}/media", h.AddExerciseMedia)
	mux.HandleFunc("GET /exercises/{id}/media", h.GetExerciseMedia)

	port := env.GetInt("EXERCISE_SERVICE_PORT", 8083)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Printf("exercise-service listening on :%d", port)
	log.Fatal(srv.ListenAndServe())

}
