package repository

import (
	"context"
	"database/sql"

	"github.com/open-workout/ow/internal/domain"
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
	// TODO
	return workout, nil
}
