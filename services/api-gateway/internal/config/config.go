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
	JWTSecret string
	JWTIssuer string

	//Services (gRPC targets)
	WorkoutServiceURL string

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
		JWTSecret: env.GetString("JWT_SECRET", "dev-secret-change-me"),
		JWTIssuer: env.GetString("JWT_ISSUER", "open-workout"),

		// Services (gRPC)
		WorkoutServiceURL: env.GetString("WORKOUT_SERVICE_URL", "localhost:50051"),

		// Observability
		LogLevel: env.GetString("LOG_LEVEL", "info"),

		// Rate limiting
		RateLimitEnabled: env.GetBool("RATE_LIMIT_ENABLED", true),
		RateLimitRPS:     env.GetInt("RATE_LIMIT_RPS", 100),
	}
}
