package domain

import "context"

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
	URL        string `json:"url"`
}

type ExerciseService interface {
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)
	AddExerciseMedia(ctx context.Context, exerciseID int64, media *ExerciseMedia) error
}

type ExerciseRepository interface {
	UpdateExercise(ctx context.Context, id int64, exerciseModel *ExerciseModel) error
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)
}
