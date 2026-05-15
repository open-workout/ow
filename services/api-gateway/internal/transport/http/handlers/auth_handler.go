package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/authclient"
	"github.com/open-workout/ow/services/api-gateway/internal/config"
)

type AuthHandler struct {
	cfg    *config.Config
	client *authclient.Client
}

func NewAuthHandler(cfg *config.Config, client *authclient.Client) *AuthHandler {
	return &AuthHandler{cfg: cfg, client: client}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken string `json:"access_token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, refreshToken, err := h.client.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, authclient.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "login failed", http.StatusBadGateway)
		return
	}

	accessToken, err := h.signAccessToken(userID, "user")
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{AccessToken: accessToken, RefreshToken: refreshToken})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := h.client.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, authclient.ErrInvalidToken) {
			http.Error(w, "invalid or expired refresh token", http.StatusUnauthorized)
			return
		}
		http.Error(w, "refresh failed", http.StatusBadGateway)
		return
	}

	accessToken, err := h.signAccessToken(userID, "user")
	if err != nil {
		http.Error(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(refreshResponse{AccessToken: accessToken})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.client.Logout(r.Context(), req.RefreshToken); err != nil {
		http.Error(w, "logout failed", http.StatusBadGateway)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type jwtClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (h *AuthHandler) signAccessToken(userID int64, role string) (string, error) {
	now := time.Now()
	c := jwtClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			Issuer:    h.cfg.JWTIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(h.cfg.AccessTokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}
