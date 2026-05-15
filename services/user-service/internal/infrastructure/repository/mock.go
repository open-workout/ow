package repository

import (
	"context"
	"time"

	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type MockRepository struct {
	Called bool

	CreateUserFunc              func(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserFunc                 func(ctx context.Context, id int64) (*domain.User, error)
	GetByUsernameFunc           func(ctx context.Context, username string) (*domain.User, error)
	UpdateUserFunc              func(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUserFunc              func(ctx context.Context, id int64) error
	UpdateSplitFunc             func(ctx context.Context, userID int64, split domain.Split) (*domain.User, error)
	CreateRefreshTokenFunc      func(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error
	GetUserIDByRefreshTokenFunc func(ctx context.Context, tokenHash string) (int64, error)
	DeleteRefreshTokenFunc      func(ctx context.Context, tokenHash string) error
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.Called = true
	return m.CreateUserFunc(ctx, user)
}

func (m *MockRepository) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	m.Called = true
	return m.GetUserFunc(ctx, id)
}

func (m *MockRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	m.Called = true
	return m.GetByUsernameFunc(ctx, username)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.Called = true
	return m.UpdateUserFunc(ctx, user)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id int64) error {
	m.Called = true
	return m.DeleteUserFunc(ctx, id)
}

func (m *MockRepository) UpdateSplit(ctx context.Context, userID int64, split domain.Split) (*domain.User, error) {
	m.Called = true
	return m.UpdateSplitFunc(ctx, userID, split)
}

func (m *MockRepository) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) error {
	m.Called = true
	return m.CreateRefreshTokenFunc(ctx, userID, tokenHash, expiresAt)
}

func (m *MockRepository) GetUserIDByRefreshToken(ctx context.Context, tokenHash string) (int64, error) {
	m.Called = true
	return m.GetUserIDByRefreshTokenFunc(ctx, tokenHash)
}

func (m *MockRepository) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	m.Called = true
	return m.DeleteRefreshTokenFunc(ctx, tokenHash)
}
