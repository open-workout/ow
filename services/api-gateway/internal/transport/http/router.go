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

	healthHandler  *handlers.HealthHandler
	workoutHandler *handlers.WorkoutHandler
}

func NewRouter(
	cfg *config.Config,
	healthHandler *handlers.HealthHandler,
	workoutHandler *handlers.WorkoutHandler,
) http.Handler {

	r := chi.NewRouter()

	router := &Router{
		cfg:            cfg,
		healthHandler:  healthHandler,
		workoutHandler: workoutHandler,
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

}
