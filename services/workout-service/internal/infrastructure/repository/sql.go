package repository

import (
	"context"
	"database/sql"
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
		INSERT INTO workouts (user_id, started_at, finished_at)
		VALUES ($1, $2, $3)
		RETURNING workout_id
	`

	var workoutId int64

	err := r.db.QueryRowContext(ctx, query, workout.UserID, workout.StartedAt, sql.NullTime{}).Scan(&workoutId)
	if err != nil {
		return nil, err
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
