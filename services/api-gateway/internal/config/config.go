package config

import (
	"time"

	"github.com/open-workout/ow/shared/env"
)

type Config struct {
	//Server
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	//Auth
	Auth0IssuerURL string
	Auth0Audience  string

	//Services
	UserServiceURL     string
	ExerciseServiceURL string
	WorkoutServiceURL  string

	//Observability
	LogLevel string

	//Rate limiting
	RateLimitEnabled bool
	RateLimitRPS     int
}

func Load() *Config {
	return &Config{
		//Server
		Port: env.GetInt("PORT", 8080),

		ReadTimeout:  time.Duration(env.GetInt("READ_TIMEOUT_MS", 5000)) * time.Millisecond,
		WriteTimeout: time.Duration(env.GetInt("WRITE_TIMEOUT_MS", 5000)) * time.Millisecond,

		// Auth
		Auth0IssuerURL: "https://" + env.GetString("AUTH0_DOMAIN", "") + "/",
		Auth0Audience:  env.GetString("AUTH0_AUDIENCE", ""),

		// Services
		UserServiceURL:     env.GetString("USER_SERVICE_URL", "http://localhost:8081"),
		ExerciseServiceURL: env.GetString("EXERCISE_SERVICE_URL", "http://localhost:8083"),
		WorkoutServiceURL:  env.GetString("WORKOUT_SERVICE_URL", "http://localhost:8082"),

		// Observability
		LogLevel: env.GetString("LOG_LEVEL", "info"),

		// Rate limiting
		RateLimitEnabled: env.GetBool("RATE_LIMIT_ENABLED", true),
		RateLimitRPS:     env.GetInt("RATE_LIMIT_RPS", 100),
	}
}
