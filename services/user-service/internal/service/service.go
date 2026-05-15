package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/open-workout/ow/services/user-service/internal/domain"
)

type Service struct {
	repo domain.UserRepository
}

func NewService(repo domain.UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return s.repo.GetUser(ctx, id)
}

func (s *Service) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return s.repo.UpdateUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, id int64) error {
	return s.repo.DeleteUser(ctx, id)
}

func (s *Service) UpdateSplit(ctx context.Context, userID int64, split domain.Split) (*domain.User, error) {
	return s.repo.UpdateSplit(ctx, userID, split)
}

func (s *Service) Login(ctx context.Context, username, password string) (*domain.LoginResult, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(b)
	h := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(h[:])
	expiresAt := time.Now().Add(90 * 24 * time.Hour)
	if err := s.repo.CreateRefreshToken(ctx, user.UserId, tokenHash, expiresAt); err != nil {
		return nil, err
	}
	return &domain.LoginResult{UserID: user.UserId, RefreshToken: token}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (int64, error) {
	h := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(h[:])
	userID, err := s.repo.GetUserIDByRefreshToken(ctx, tokenHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, domain.ErrNotFound
		}
		return 0, err
	}
	return userID, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	h := sha256.Sum256([]byte(refreshToken))
	tokenHash := hex.EncodeToString(h[:])
	return s.repo.DeleteRefreshToken(ctx, tokenHash)
}
