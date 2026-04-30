package repository

import (
	"context"
	"database/sql"

	"github.com/open-workout/ow/services/exercise-service/internal/domain"
)

type SqlRepository struct {
	db *sql.DB
}

func NewSqlRepository(db *sql.DB) *SqlRepository {
	return &SqlRepository{
		db: db,
	}
}

func (r *SqlRepository) CreateExercise(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {

	query := `
	INSERT INTO exercises  (name, exercise_type, primary_muscle, secondary_muscles, description, user_id, is_private)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING exercise_id
`

	var exerciseId int64

	err := r.db.QueryRowContext(
		ctx,
		query,
		exercise.Name,
		exercise.ExerciseType,
		exercise.PrimaryMuscle,
		exercise.SecondaryMuscles,
		exercise.Description,
		exercise.UserID,
		exercise.IsPrivate,
	).Scan(&exerciseId)

	if err != nil {
		return nil, err
	}

	exercise.ExerciseID = int(exerciseId)

	return exercise, nil

}

func (r *SqlRepository) UpdateExercise(ctx context.Context, exercise *domain.ExerciseModel) (*domain.ExerciseModel, error) {
	query := `
	UPDATE exercises
	SET
		name = COALESCE(name, $1),
		exercise_type = COALESCE(exercise_type, $2),
		primary_muscle = COALESCE(primary_muscle, $3),
		secondary_muscles = COALESCE(secondary_muscles, $4),
		description = COALESCE(description, $5)
	WHERE exercise_id = $6
	RETURNING exercise_id
	`

	var exerciseId int64

}
