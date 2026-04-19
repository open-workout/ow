package domain

import "context"

type WorkoutModel struct {
	ID     string
	UserID string
}

type SetModel struct {
	WorkoutID  string
	ExerciseID string
	Reps       int
	Difficulty int
	Weight     float64
}

type WorkoutRepository interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)
}

type WorkoutService interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)
}
