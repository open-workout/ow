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
		RETURNING set_id
	`

	loggedAt := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		set.WorkoutID, set.ExerciseID, set.Reps, set.Difficulty, set.Weight, set.Unit, loggedAt,
	).Scan(&set.SetID)
	if err != nil {
		return nil, err
	}

	set.LoggedAt = loggedAt
	return set, nil
}

func (r *SqlRepository) SetWorkoutFinishTime(ctx context.Context, workoutId int64, finishedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE workouts SET finished_at = $1 WHERE workout_id = $2`, finishedAt, workoutId)
	return err
}

func (r *SqlRepository) GetSetsByWorkoutID(ctx context.Context, workoutId int64, userId string) ([]*domain.SetModel, error) {
	query := `
		SELECT ws.set_id, ws.workout_id, ws.exercise_id, ws.reps, ws.difficulty, ws.weight, ws.unit, ws.logged_at
		FROM workout_sets ws
		JOIN workouts w ON ws.workout_id = w.workout_id
		WHERE ws.workout_id = $1
		  AND w.user_id = $2
		ORDER BY ws.logged_at
	`

	rows, err := r.db.QueryContext(ctx, query, workoutId, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []*domain.SetModel
	for rows.Next() {
		s := &domain.SetModel{}
		if err := rows.Scan(&s.SetID, &s.WorkoutID, &s.ExerciseID, &s.Reps, &s.Difficulty, &s.Weight, &s.Unit, &s.LoggedAt); err != nil {
			return nil, err
		}
		sets = append(sets, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sets, nil
}

func (r *SqlRepository) GetLastTimeMaxSet(ctx context.Context, userId string, exerciseId int64) (*domain.SetModel, error) {
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
		&set.WorkoutID, &set.ExerciseID, &set.Reps, &set.Difficulty, &set.Weight, &set.Unit, &set.LoggedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.SetModel{}, nil
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

func (r *SqlRepository) DeleteWorkoutsByUserID(ctx context.Context, userId string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx,
		`DELETE FROM workout_sets WHERE workout_id IN (SELECT workout_id FROM workouts WHERE user_id = $1)`, userId,
	); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM workouts WHERE user_id = $1`, userId); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SqlRepository) UpdateSet(ctx context.Context, userId string, set *domain.SetModel) (*domain.SetModel, error) {
	query := `
		UPDATE workout_sets ws
		SET reps = $1, difficulty = $2, weight = $3, unit = $4
		FROM workouts w
		WHERE ws.set_id = $5
		  AND ws.workout_id = w.workout_id
		  AND w.user_id = $6
		RETURNING ws.set_id, ws.workout_id, ws.exercise_id, ws.reps, ws.difficulty, ws.weight, ws.unit, ws.logged_at
	`

	updated := &domain.SetModel{}
	err := r.db.QueryRowContext(ctx, query, set.Reps, set.Difficulty, set.Weight, set.Unit, set.SetID, userId).Scan(
		&updated.SetID, &updated.WorkoutID, &updated.ExerciseID,
		&updated.Reps, &updated.Difficulty, &updated.Weight, &updated.Unit, &updated.LoggedAt,
	)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (r *SqlRepository) DeleteSet(ctx context.Context, userId string, setId int64) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM workout_sets ws
		USING workouts w
		WHERE ws.set_id = $1
		  AND ws.workout_id = w.workout_id
		  AND w.user_id = $2
	`, setId, userId)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SqlRepository) GetWorkoutById(ctx context.Context, workoutId int64) (*domain.WorkoutModel, error) {
	query := `SELECT workout_id, user_id, started_at, finished_at, title FROM workouts WHERE workout_id = $1`

	workoutModel := &domain.WorkoutModel{}
	var finishedAt sql.NullTime
	var wTitle sql.NullString
	err := r.db.QueryRowContext(ctx, query, workoutId).Scan(
		&workoutModel.WorkoutID, &workoutModel.UserID,
		&workoutModel.StartedAt, &finishedAt, &wTitle,
	)
	if err != nil {
		return &domain.WorkoutModel{}, err
	}

	if finishedAt.Valid {
		workoutModel.FinishedAt = finishedAt.Time.UTC()
	}
	if wTitle.Valid {
		workoutModel.Title = wTitle.String
	} else {
		workoutModel.Title = "Workout"
	}

	return workoutModel, nil
}
