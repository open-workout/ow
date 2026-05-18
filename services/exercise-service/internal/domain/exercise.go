package domain

import (
	"context"
	"errors"
	"io"
)

var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

type ExerciseModel struct {
	ExerciseID       int64    `json:"exercise_id"`
	Name             string   `json:"name"`
	ExerciseType     string   `json:"exercise_type"`
	PrimaryMuscle    string   `json:"primary_muscle"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	Description      string   `json:"description"`
	UserID           string   `json:"user_id"`
	IsPrivate        bool     `json:"is_private"`
	WeightDirection  int64    `json:"weight_direction"`
}

type ExerciseMedia struct {
	ExerciseID int64     `json:"exercise_id"`
	UserID     string    `json:"user_id"`
	URL        string    `json:"url"`
	File       io.Reader `json:"-"`
}

type ExerciseMediaUpload struct {
	ExerciseID int64  `json:"exercise_id"`
	UserID     string `json:"user_id"`
	File       io.Reader
	Filename   string `json:"filename"`
	MimeType   string `json:"mime_type"`
}

type MuscleState struct {
	Muscles map[string]float64
	UserID  string
}

type ExerciseService interface {
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)
	AddExerciseMedia(ctx context.Context, exerciseID int64, callerUserID string, media *ExerciseMedia, mediaFile *ExerciseMediaUpload) error
	GetExerciseMedia(ctx context.Context, exerciseID int64, callerUserID string) ([]ExerciseMedia, error)

	GetTopExercises(ctx context.Context, state MuscleState, limit int) ([]ExerciseModel, error)

	ListExercises(ctx context.Context, userID string) ([]ExerciseModel, error)

	GetExerciseById(ctx context.Context, id int64, callerUserID string) (*ExerciseModel, error)

	UpdateExercise(ctx context.Context, callerUserID string, exercise *ExerciseModel) (*ExerciseModel, error)
	DeleteExercise(ctx context.Context, callerUserID string, id int64) error
}

type ExerciseRepository interface {
	CreateExercise(ctx context.Context, exercise *ExerciseModel) (*ExerciseModel, error)

	AddExerciseMedia(ctx context.Context, exerciseID int64, media *ExerciseMedia) error
	GetExerciseMedia(ctx context.Context, exerciseID int64, callerUserID string) ([]ExerciseMedia, error)

	ListExercises(ctx context.Context, userID string) ([]ExerciseModel, error)
	ListPublicExercises(ctx context.Context) ([]ExerciseModel, error)
	ListUserExercises(ctx context.Context, userID string) ([]ExerciseModel, error)

	GetExerciseById(ctx context.Context, id int64, callerUserID string) (*ExerciseModel, error)

	UpdateExercise(ctx context.Context, callerUserID string, exercise *ExerciseModel) (*ExerciseModel, error)
	DeleteExercise(ctx context.Context, callerUserID string, id int64) error
}

type MediaStorage interface {
	Upload(ctx context.Context, file *ExerciseMediaUpload) (string, error)
}
