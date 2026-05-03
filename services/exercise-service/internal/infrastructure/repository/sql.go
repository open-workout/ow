package repository

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
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
		pq.Array(exercise.SecondaryMuscles),
		exercise.Description,
		exercise.UserID,
		exercise.IsPrivate,
	).Scan(&exerciseId)

	if err != nil {
		return nil, err
	}

	exercise.ExerciseID = exerciseId

	return exercise, nil

}

func (r *SqlRepository) AddExerciseMedia(ctx context.Context, exerciseID int64, media *domain.ExerciseMedia) error {

	query := `
		INSERT INTO exercise_media (exercise_id, url, user_id) 
		VALUES ($1, $2, $3)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		media.ExerciseID,
		media.URL,
		media.UserID,
	)
	return err

}

func (r *SqlRepository) ListPublicExercises(ctx context.Context) ([]domain.ExerciseModel, error) {

	query := `
		SELECT *
		FROM exercises
		WHERE is_private = false
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var exercises []domain.ExerciseModel

	for rows.Next() {
		var ex domain.ExerciseModel

		if err := rows.Scan(
			&ex.ExerciseID,
			&ex.Name,
			&ex.ExerciseType,
			&ex.PrimaryMuscle,
			pq.Array(&ex.SecondaryMuscles),
			&ex.Description,
			&ex.UserID,
			&ex.IsPrivate,
		); err != nil {
			return nil, err
		}

		exercises = append(exercises, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return exercises, nil
}

func (r *SqlRepository) ListUserExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {

	query := `
		SELECT *
		FROM exercises
		WHERE is_private = true
		  AND user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var exercises []domain.ExerciseModel

	for rows.Next() {
		var ex domain.ExerciseModel

		if err := rows.Scan(
			&ex.ExerciseID,
			&ex.Name,
			&ex.ExerciseType,
			&ex.PrimaryMuscle,
			pq.Array(&ex.SecondaryMuscles),
			&ex.Description,
			&ex.UserID,
			&ex.IsPrivate,
		); err != nil {
			return nil, err
		}

		exercises = append(exercises, ex)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return exercises, nil
}

func (r *SqlRepository) ListExercises(ctx context.Context, userID int64) ([]domain.ExerciseModel, error) {
	public, err := r.ListPublicExercises(ctx)
	if err != nil {
		return nil, err
	}

	private, err := r.ListUserExercises(ctx, userID)
	if err != nil {
		return nil, err
	}

	return append(public, private...), nil

}
