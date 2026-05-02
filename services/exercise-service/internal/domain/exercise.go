package domain

import (
	"context"
	"io"
)

type ExerciseModel struct {
	ExerciseID       int64    `json:"exercise_id"`
	Name             string   `json:"name"`
	ExerciseType     string   `json:"exercise_type"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Description      string   `json:"description"`
	UserID           int64    `json:"user_id"`
	IsPrivate        bool     `json:"is_private"`
}

type ExerciseMedia struct {
	ExerciseID int64  `json:"exercise_id"`
	UserID     int64  `json:"user_id"`
	URL        string `json:"url"`
}

type MuscleState struct {
	Muscles map[string]float64
	UserID  int64
}

type ExerciseService interface {
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)
	UpdateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)
	AddExerciseMedia(ctx context.Context, exerciseID int64, media *ExerciseMedia) error

	GetBestExercise(ctx context.Context, state MuscleState) (*ExerciseModel, error)
}

type ExerciseRepository interface {
	UpdateExercise(ctx context.Context, id int64, exerciseModel *ExerciseModel) (*ExerciseModel, error)
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)

	AddExerciseMedia(ctx context.Context, exerciseID int64, media *ExerciseMedia) error

	ListExercises(ctx context.Context, userID int64) ([]ExerciseModel, error)
	ListPublicExercises(ctx context.Context) ([]ExerciseModel, error)
	ListUserExercises(ctx context.Context, userID int64) ([]ExerciseModel, error)
}

type MediaStorage interface {
	Upload(ctx context.Context, file io.Reader, filename string) (string, error)
}
