package domain

import (
	"context"
	"errors"
)

var (
	ErrNotFound = errors.New("not found")
)

type User struct {
	UserId        string   `json:"user_id"`
	Email         string   `json:"email"`
	Username      string   `json:"username"`
	SportGoals    []string `json:"sport_goals"`
	Gender        string   `json:"gender"`
	Birthdate     string   `json:"birthdate"`
	ExerciseSplit Split    `json:"split"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateSplit(ctx context.Context, userID string, split Split) (*User, error)
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateSplit(ctx context.Context, userID string, split Split) (*User, error)
}
