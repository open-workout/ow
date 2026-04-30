package domain

import (
	"context"
	"time"
)

type SetModel struct {
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id"`
	Reps       int       `json:"reps"`
	Difficulty int       `json:"difficulty"`
	Weight     float64   `json:"weight"`
	LoggedAt   time.Time `json:"logged_at"`
}

type WorkoutModel struct {
	WorkoutID  int64     `json:"workout_id"`
	UserID     int64     `json:"user_id"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

type WorkoutRepository interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
}

type WorkoutService interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
}
