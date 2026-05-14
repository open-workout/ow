package domain

import (
	"context"
	"time"
)

type SetModel struct {
	SetID      int64     `json:"set_id"`
	WorkoutID  int64     `json:"workout_id"`
	ExerciseID int64     `json:"exercise_id"`
	Reps       int       `json:"reps"`
	Difficulty int       `json:"difficulty"`
	Weight     float64   `json:"weight"`
	Unit       string    `json:"unit"`
	LoggedAt   time.Time `json:"logged_at"`
}

type WorkoutModel struct {
	WorkoutID  int64     `json:"workout_id"`
	UserID     int64     `json:"user_id"`
	Title      string    `json:"title,omitempty"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at,omitempty"`
}

type WorkoutRepository interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)
	SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error
	GetWorkoutById(ctx context.Context, workoutId int64) (*WorkoutModel, error)
	DeleteWorkout(ctx context.Context, workoutId int64) error
	DeleteWorkoutsByUserID(ctx context.Context, userId int64) error

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
	UpdateSet(ctx context.Context, userId int64, set *SetModel) (*SetModel, error)
	DeleteSet(ctx context.Context, userId int64, setId int64) error
	GetSetsByWorkoutID(ctx context.Context, workoutId int64, userId int64) ([]*SetModel, error)

	GetLastTimeMaxSet(ctx context.Context, userId int64, exerciseId int64) (*SetModel, error)
}

type WorkoutService interface {
	CreateWorkout(ctx context.Context, workout *WorkoutModel) (*WorkoutModel, error)
	SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error
	GetWorkoutById(ctx context.Context, workoutId int64) (*WorkoutModel, error)
	DeleteWorkout(ctx context.Context, workoutId int64) error
	DeleteWorkoutsByUserID(ctx context.Context, userId int64) error

	CreateSet(ctx context.Context, set *SetModel) (*SetModel, error)
	UpdateSet(ctx context.Context, userId int64, set *SetModel) (*SetModel, error)
	DeleteSet(ctx context.Context, userId int64, setId int64) error
	GetSetsByWorkoutID(ctx context.Context, workoutId int64, userId int64) ([]*SetModel, error)

	GetLastTimeMaxSet(ctx context.Context, userId int64, exerciseId int64) (*SetModel, error)
}
