package domain

import (
	"context"
	"time"
)

type WorkoutModel struct {
	WorkoutID int `json:"workout_id"`
	UserID    int `json:"user_id"`
}

type SetModel struct {
	WorkoutID  int       `json:"workout_id"`
	ExerciseID int       `json:"exercise_id"`
	Reps       int       `json:"reps"`
	Difficulty int       `json:"difficulty"`
	Weight     float64   `json:"weight"`
	LoggedAt   time.Time `json:"logged_at"`
}

type WorkoutRepository interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
}

type WorkoutService interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
}
