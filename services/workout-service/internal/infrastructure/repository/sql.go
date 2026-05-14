package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/open-workout/ow/services/workout-service/internal/domain"
)

type SqlRepository struct {
	db *sql.DB
}

func NewSqlRepository(db *sql.DB) *SqlRepository {
	return &SqlRepository{
		db: db,
	}
}

func (r *SqlRepository) CreateWorkout(ctx context.Context, workout *domain.WorkoutModel) (*domain.WorkoutModel, error) {

	query := `
		INSERT INTO workouts (user_id, started_at)
		VALUES ($1, $2)
		RETURNING workout_id
	`

	var workoutId int64

	err := r.db.QueryRowContext(ctx, query, workout.UserID, workout.StartedAt).Scan(&workoutId)
	if err != nil {
		return &domain.WorkoutModel{}, err
	}

	workout.WorkoutID = workoutId
	return workout, nil

}

func (r *SqlRepository) CreateSet(ctx context.Context, set *domain.SetModel) (*domain.SetModel, error) {

	query := `
		INSERT INTO workout_sets (workout_id, exercise_id, reps, difficulty, weight, unit, logged_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	loggedAt := time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		set.WorkoutID,
		set.ExerciseID,
		set.Reps,
		set.Difficulty,
		set.Weight,
		set.Unit,
		loggedAt,
	)

	if err != nil {
		return nil, err
	}

	set.LoggedAt = loggedAt

	return set, nil
}

func (r *SqlRepository) SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error {
	query := `
	UPDATE workouts SET finished_at = $1 WHERE workout_id = $2
`
	_, err := r.db.ExecContext(ctx, query, finishedAt, workoutId)
	if err != nil {
		return err
	}
	return nil
}

func (r *SqlRepository) GetLastTimeMaxSet(ctx context.Context, userId int64, exerciseId int64) (*domain.SetModel, error) {

	query := `
	WITH latest_workout AS (
		SELECT w.workout_id
		FROM workouts w
		JOIN workout_sets s ON s.workout_id = w.workout_id
		WHERE w.user_id = $1
		  AND s.exercise_id = $2
		  AND w.finished_at IS NOT NULL
		AND w.finished_at > w.started_at
		ORDER BY w.finished_at DESC
		LIMIT 1
	)
	SELECT s.workout_id, s.exercise_id, s.reps, s.difficulty, s.weight, s.unit, s.logged_at
	FROM workout_sets s
	JOIN latest_workout lw ON lw.workout_id = s.workout_id
	WHERE s.exercise_id = $2
	ORDER BY s.weight DESC, s.logged_at DESC
	LIMIT 1;
`

	var set domain.SetModel

	err := r.db.QueryRowContext(ctx, query, userId, exerciseId).Scan(
		&set.WorkoutID,
		&set.ExerciseID,
		&set.Reps,
		&set.Difficulty,
		&set.Weight,
		&set.Unit,
		&set.LoggedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.SetModel{}, nil // no previous set found
		}
		return &domain.SetModel{}, err
	}

	return &set, nil
}

func (r *SqlRepository) DeleteWorkout(ctx context.Context, workoutId int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM workout_sets WHERE workout_id = $1`, workoutId); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM workouts WHERE workout_id = $1`, workoutId); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SqlRepository) DeleteWorkoutsByUserID(ctx context.Context, userId int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM workout_sets WHERE workout_id IN (SELECT workout_id FROM workouts WHERE user_id = $1)`, userId); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM workouts WHERE user_id = $1`, userId); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SqlRepository) DeleteSet(ctx context.Context, workoutId int64, exerciseId int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM workout_sets WHERE workout_id = $1 AND exercise_id = $2`, workoutId, exerciseId)
	return err
}

func (r *SqlRepository) GetWorkoutById(ctx context.Context, workoutId int64) (*domain.WorkoutModel, error) {
	query := `
	SELECT workout_id, user_id, started_at, finished_at, title FROM workouts WHERE workout_id = $1
	`

	workoutModel := &domain.WorkoutModel{}
	var finishedAt sql.NullTime
	var wTitle sql.NullString
	err := r.db.QueryRowContext(ctx, query, workoutId).Scan(
		&workoutModel.WorkoutID,
		&workoutModel.UserID,
		&workoutModel.StartedAt,
		&finishedAt,
		&wTitle,
	)

	if err != nil {
		return &domain.WorkoutModel{}, err
	}

	if finishedAt.Valid {
		workoutModel.FinishedAt = finishedAt.Time.UTC()
	} else {
		workoutModel.FinishedAt = time.Time{}
	}

	if wTitle.Valid {
		workoutModel.Title = wTitle.String
	} else {
		workoutModel.Title = "Workout"
	}

	return workoutModel, nil
}
