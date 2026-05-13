package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/open-workout/ow/services/api-gateway/internal/clients/exerciseclient"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/userclient"
	"github.com/open-workout/ow/services/api-gateway/internal/config"
	transport "github.com/open-workout/ow/services/api-gateway/internal/transport/http"
	"github.com/open-workout/ow/services/api-gateway/internal/transport/http/handlers"
)

func main() {
	cfg := config.Load()

	userClient := userclient.New(cfg.UserServiceURL)
	exerciseClient := exerciseclient.New(cfg.ExerciseServiceURL)

	healthHandler := handlers.NewHealthHandler()
	userHandler := handlers.NewUserHandler(userClient)
	exerciseHandler := handlers.NewExerciseHandler(exerciseClient)
	workoutHandler := &handlers.WorkoutHandler{}

	h := transport.NewRouter(cfg, healthHandler, workoutHandler, userHandler, exerciseHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      h,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	log.Printf("api-gateway listening on :%d", cfg.Port)
	log.Fatal(srv.ListenAndServe())
}
