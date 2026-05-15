package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type User struct {
	UserId        int64    `json:"user_id"`
	Email         string   `json:"email"`
	Username      string   `json:"username"`
	PasswordHash  string   `json:"-"`
	SportGoals    []string `json:"sport_goals"`
	Gender        string   `json:"gender"`
	Birthdate     string   `json:"birthdate"`
	ExerciseSplit Split    `json:"split"`
}

type LoginResult struct {
	UserID       int64
	RefreshToken string
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateSplit(ctx context.Context, userID int64, split Split) (*User, error)
	CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetUserIDByRefreshToken(ctx context.Context, tokenHash string) (int64, error)
	DeleteRefreshToken(ctx context.Context, tokenHash string) error
}

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, id int64) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
	UpdateSplit(ctx context.Context, userID int64, split Split) (*User, error)
	Login(ctx context.Context, username, password string) (*LoginResult, error)
	Refresh(ctx context.Context, refreshToken string) (int64, error)
	Logout(ctx context.Context, refreshToken string) error
}
