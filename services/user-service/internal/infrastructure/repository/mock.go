package repository

import (
	"context"

	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type MockRepository struct {
	Called bool

	CreateUserFunc  func(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserFunc     func(ctx context.Context, id string) (*domain.User, error)
	UpdateUserFunc  func(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUserFunc  func(ctx context.Context, id string) error
	UpdateSplitFunc func(ctx context.Context, userID string, split domain.Split) (*domain.User, error)
}

func NewMockRepository() *MockRepository {
	return &MockRepository{}
}

func (m *MockRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.Called = true
	return m.CreateUserFunc(ctx, user)
}

func (m *MockRepository) GetUser(ctx context.Context, id string) (*domain.User, error) {
	m.Called = true
	return m.GetUserFunc(ctx, id)
}

func (m *MockRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	m.Called = true
	return m.UpdateUserFunc(ctx, user)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id string) error {
	m.Called = true
	return m.DeleteUserFunc(ctx, id)
}

func (m *MockRepository) UpdateSplit(ctx context.Context, userID string, split domain.Split) (*domain.User, error) {
	m.Called = true
	return m.UpdateSplitFunc(ctx, userID, split)
}
