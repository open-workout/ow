package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-chi/chi/v5/middleware"

	"github.com/open-workout/ow/services/api-gateway/internal/config"
	"github.com/open-workout/ow/services/api-gateway/internal/transport/http/handlers"

	appmw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

type Router struct {
	cfg *config.Config

	healthHandler   *handlers.HealthHandler
	workoutHandler  *handlers.WorkoutHandler
	userHandler     *handlers.UserHandler
	exerciseHandler *handlers.ExerciseHandler
	authHandler     *handlers.AuthHandler
}

func NewRouter(
	cfg *config.Config,
	healthHandler *handlers.HealthHandler,
	workoutHandler *handlers.WorkoutHandler,
	userHandler *handlers.UserHandler,
	exerciseHandler *handlers.ExerciseHandler,
	authHandler *handlers.AuthHandler,
) http.Handler {

	r := chi.NewRouter()

	router := &Router{
		cfg:             cfg,
		healthHandler:   healthHandler,
		workoutHandler:  workoutHandler,
		userHandler:     userHandler,
		exerciseHandler: exerciseHandler,
		authHandler:     authHandler,
	}

	router.register(r)

	return r
}

func (rt *Router) register(r chi.Router) {

	// =====================
	// Global middleware
	// =====================
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Custom middleware
	r.Use(appmw.CORS())
	r.Use(appmw.RateLimiter(rt.cfg))
	r.Use(appmw.Logging())

	// =====================
	// Health
	// =====================
	r.Get("/health", rt.healthHandler.Check)

	// =====================
	// Auth
	// =====================
	r.Post("/auth/login", rt.authHandler.Login)
	r.Post("/auth/refresh", rt.authHandler.Refresh)
	r.Post("/auth/logout", rt.authHandler.Logout)

	// =====================
	// Users
	// =====================
	r.Post("/users", rt.userHandler.CreateUser)
	r.Route("/users", func(r chi.Router) {
		r.Use(appmw.Auth(rt.cfg.JWTSecret))
		r.Get("/{id}", rt.userHandler.GetUser)
		r.Put("/{id}", rt.userHandler.UpdateUser)
		r.Delete("/{id}", rt.userHandler.DeleteUser)
		r.Put("/{id}/split", rt.userHandler.UpdateSplit)
	})

	// =====================
	// Exercises
	// =====================
	r.Route("/exercises", func(r chi.Router) {
		r.Use(appmw.Auth(rt.cfg.JWTSecret))
		r.Get("/", rt.exerciseHandler.ListExercises)
		r.Post("/", rt.exerciseHandler.CreateExercise)
		r.Post("/recommendations", rt.exerciseHandler.GetTopExercises)
		r.Get("/{id}", rt.exerciseHandler.GetExerciseById)
		r.Post("/{id}/media", rt.exerciseHandler.AddExerciseMedia)
		r.Get("/{id}/media", rt.exerciseHandler.GetExerciseMedia)
	})

	// =====================
	// Workouts & Sets
	// =====================
	r.Route("/workouts", func(r chi.Router) {
		r.Use(appmw.Auth(rt.cfg.JWTSecret))
		r.Get("/{workout_id}", rt.workoutHandler.GetWorkout)
		r.Get("/{workout_id}/sets", rt.workoutHandler.GetSets)
	})

	r.Route("/sets", func(r chi.Router) {
		r.Use(appmw.Auth(rt.cfg.JWTSecret))
		r.Put("/{set_id}", rt.workoutHandler.UpdateSet)
		r.Delete("/{set_id}", rt.workoutHandler.DeleteSet)
	})

}
