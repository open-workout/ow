package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/open-workout/ow/services/user-service/internal/domain"
	"github.com/open-workout/ow/services/user-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/user-service/internal/service"
)

const testUserID = "auth0|svc-1"

func TestService_CreateUser_Success(t *testing.T) {
	mock := repository.NewMockRepository()
	want := &domain.User{UserId: testUserID, Email: "a@example.com", Username: "alice"}
	mock.CreateUserFunc = func(_ context.Context, u *domain.User) (*domain.User, error) {
		return want, nil
	}
	svc := service.NewService(mock)

	got, err := svc.CreateUser(context.Background(), &domain.User{UserId: testUserID, Email: "a@example.com", Username: "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserId != testUserID {
		t.Errorf("expected UserId %s, got %s", testUserID, got.UserId)
	}
}

func TestService_GetUser_NotFound(t *testing.T) {
	mock := repository.NewMockRepository()
	mock.GetUserFunc = func(_ context.Context, _ string) (*domain.User, error) {
		return nil, sql.ErrNoRows
	}
	svc := service.NewService(mock)

	_, err := svc.GetUser(context.Background(), "auth0|nobody")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestService_DeleteUser_Success(t *testing.T) {
	deleted := false
	mock := repository.NewMockRepository()
	mock.DeleteUserFunc = func(_ context.Context, id string) error {
		deleted = true
		return nil
	}
	svc := service.NewService(mock)

	if err := svc.DeleteUser(context.Background(), testUserID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DeleteUser to be called")
	}
}
