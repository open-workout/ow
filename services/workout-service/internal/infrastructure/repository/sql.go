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
		INSERT INTO workout_sets (workout_id, exercise_id, reps, difficulty, weight, logged_at)
		VALUES ($1, $2, $3, $4, $5, $6)
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
	SELECT s.workout_id, s.exercise_id, s.reps, s.difficulty, s.weight, s.logged_at
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

func (r *SqlRepository) GetWorkoutById(ctx context.Context, workoutId int64) (*domain.WorkoutModel, error) {
	query := `
	SELECT * FROM workouts WHERE workout_id = $1
	`

	workoutModel := &domain.WorkoutModel{}
	var finishedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, workoutId).Scan(
		&workoutModel.WorkoutID,
		&workoutModel.UserID,
		&workoutModel.StartedAt,
		&finishedAt,
	)

	if err != nil {
		return &domain.WorkoutModel{}, err
	}

	if finishedAt.Valid {
		workoutModel.FinishedAt = finishedAt.Time.UTC()
	} else {
		workoutModel.FinishedAt = time.Time{}
	}

	return workoutModel, nil
}
