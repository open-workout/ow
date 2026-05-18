package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type SqlRepository struct {
	db *sql.DB
}

func NewSqlRepository(db *sql.DB) *SqlRepository {
	return &SqlRepository{db: db}
}

func (r *SqlRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	splitJSON, err := json.Marshal(user.ExerciseSplit)
	if err != nil {
		return nil, err
	}

	goals := user.SportGoals
	if goals == nil {
		goals = []string{}
	}

	query := `
		INSERT INTO users (user_id, email, username, sport_goals, gender, birthdate, split)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING user_id, email, username, sport_goals, gender, birthdate, split
	`
	return scanUser(r.db.QueryRowContext(ctx, query,
		user.UserId,
		user.Email,
		user.Username,
		pq.Array(goals),
		user.Gender,
		user.Birthdate,
		splitJSON,
	))
}

func (r *SqlRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT user_id, email, username, sport_goals, gender, birthdate, split FROM users WHERE user_id = $1`
	return scanUser(r.db.QueryRowContext(ctx, query, id))
}

func (r *SqlRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	goals := user.SportGoals
	if goals == nil {
		goals = []string{}
	}

	query := `
		UPDATE users
		SET email = $1, username = $2, sport_goals = $3, gender = $4, birthdate = $5
		WHERE user_id = $6
		RETURNING user_id, email, username, sport_goals, gender, birthdate, split
	`
	return scanUser(r.db.QueryRowContext(ctx, query,
		user.Email,
		user.Username,
		pq.Array(goals),
		user.Gender,
		user.Birthdate,
		user.UserId,
	))
}

func (r *SqlRepository) DeleteUser(ctx context.Context, id string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE user_id = $1`, id)
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

func (r *SqlRepository) UpdateSplit(ctx context.Context, userID string, split domain.Split) (*domain.User, error) {
	splitJSON, err := json.Marshal(split)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE users
		SET split = $1
		WHERE user_id = $2
		RETURNING user_id, email, username, sport_goals, gender, birthdate, split
	`
	return scanUser(r.db.QueryRowContext(ctx, query, splitJSON, userID))
}

func scanUser(row *sql.Row) (*domain.User, error) {
	var u domain.User
	var splitJSON []byte

	err := row.Scan(
		&u.UserId,
		&u.Email,
		&u.Username,
		pq.Array(&u.SportGoals),
		&u.Gender,
		&u.Birthdate,
		&splitJSON,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(splitJSON, &u.ExerciseSplit); err != nil {
		return nil, err
	}

	return &u, nil
}
