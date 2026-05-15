package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/open-workout/ow/services/user-service/internal/domain"
	"github.com/open-workout/ow/services/user-service/internal/infrastructure/repository"
	"github.com/open-workout/ow/services/user-service/internal/service"
)

func hashPwd(t *testing.T, pw string) string {
	t.Helper()
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(b)
}

// --- Login ---

func TestService_Login_Success(t *testing.T) {
	hash := hashPwd(t, "correct")
	var storedUserID int64
	mock := repository.NewMockRepository()
	mock.GetByUsernameFunc = func(_ context.Context, username string) (*domain.User, error) {
		return &domain.User{UserId: 42, Username: username, PasswordHash: hash}, nil
	}
	mock.CreateRefreshTokenFunc = func(_ context.Context, userID int64, _ string, expiresAt time.Time) error {
		storedUserID = userID
		if expiresAt.Before(time.Now().Add(89 * 24 * time.Hour)) {
			t.Error("expected expiry ~90 days from now")
		}
		return nil
	}
	svc := service.NewService(mock)

	result, err := svc.Login(context.Background(), "john", "correct")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", result.UserID)
	}
	if len(result.RefreshToken) != 64 {
		t.Errorf("expected 64-char hex token, got %d chars", len(result.RefreshToken))
	}
	if storedUserID != 42 {
		t.Errorf("expected refresh token stored for userID 42, got %d", storedUserID)
	}
}

func TestService_Login_UserNotFound(t *testing.T) {
	mock := repository.NewMockRepository()
	mock.GetByUsernameFunc = func(_ context.Context, _ string) (*domain.User, error) {
		return nil, sql.ErrNoRows
	}
	svc := service.NewService(mock)

	_, err := svc.Login(context.Background(), "nobody", "pw")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Login_WrongPassword(t *testing.T) {
	hash := hashPwd(t, "correct")
	mock := repository.NewMockRepository()
	mock.GetByUsernameFunc = func(_ context.Context, _ string) (*domain.User, error) {
		return &domain.User{UserId: 1, Username: "john", PasswordHash: hash}, nil
	}
	svc := service.NewService(mock)

	_, err := svc.Login(context.Background(), "john", "wrong")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Login_TokenStorageError(t *testing.T) {
	hash := hashPwd(t, "pw")
	mock := repository.NewMockRepository()
	mock.GetByUsernameFunc = func(_ context.Context, _ string) (*domain.User, error) {
		return &domain.User{UserId: 1, Username: "john", PasswordHash: hash}, nil
	}
	mock.CreateRefreshTokenFunc = func(_ context.Context, _ int64, _ string, _ time.Time) error {
		return errors.New("db down")
	}
	svc := service.NewService(mock)

	_, err := svc.Login(context.Background(), "john", "pw")
	if err == nil {
		t.Fatal("expected error when token storage fails")
	}
}

// --- Refresh ---

func TestService_Refresh_Success(t *testing.T) {
	mock := repository.NewMockRepository()
	mock.GetUserIDByRefreshTokenFunc = func(_ context.Context, _ string) (int64, error) {
		return 42, nil
	}
	svc := service.NewService(mock)

	userID, err := svc.Refresh(context.Background(), "any-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != 42 {
		t.Errorf("expected userID 42, got %d", userID)
	}
}

func TestService_Refresh_TokenNotFound(t *testing.T) {
	mock := repository.NewMockRepository()
	mock.GetUserIDByRefreshTokenFunc = func(_ context.Context, _ string) (int64, error) {
		return 0, sql.ErrNoRows
	}
	svc := service.NewService(mock)

	_, err := svc.Refresh(context.Background(), "bad-token")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// --- Logout ---

func TestService_Logout_Success(t *testing.T) {
	deleted := false
	mock := repository.NewMockRepository()
	mock.DeleteRefreshTokenFunc = func(_ context.Context, _ string) error {
		deleted = true
		return nil
	}
	svc := service.NewService(mock)

	if err := svc.Logout(context.Background(), "some-token"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Error("expected DeleteRefreshToken to be called")
	}
}

func TestService_Logout_Error(t *testing.T) {
	mock := repository.NewMockRepository()
	mock.DeleteRefreshTokenFunc = func(_ context.Context, _ string) error {
		return errors.New("db down")
	}
	svc := service.NewService(mock)

	if err := svc.Logout(context.Background(), "token"); err == nil {
		t.Fatal("expected error from Logout, got nil")
	}
}
