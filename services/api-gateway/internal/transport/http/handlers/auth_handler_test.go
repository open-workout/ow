package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/open-workout/ow/services/api-gateway/internal/clients/authclient"
	"github.com/open-workout/ow/services/api-gateway/internal/config"
	"github.com/open-workout/ow/services/api-gateway/internal/transport/http/handlers"
)

func newTestAuthHandler(t *testing.T, userSvcHandler http.Handler) *handlers.AuthHandler {
	t.Helper()
	srv := httptest.NewServer(userSvcHandler)
	t.Cleanup(srv.Close)
	cfg := &config.Config{
		JWTSecret:      "test-secret",
		JWTIssuer:      "test",
		AccessTokenTTL: time.Hour,
	}
	return handlers.NewAuthHandler(cfg, authclient.New(srv.URL))
}

func parseToken(t *testing.T, tokenStr string) jwt.MapClaims {
	t.Helper()
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !tok.Valid {
		t.Fatalf("invalid JWT: %v", err)
	}
	claims, _ := tok.Claims.(jwt.MapClaims)
	return claims
}

// --- Login ---

func TestAuthHandler_Login_Success(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"user_id": 42, "refresh_token": "opaque-tok"})
	}))

	body, _ := json.Marshal(map[string]string{"username": "john", "password": "pw"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if resp.RefreshToken != "opaque-tok" {
		t.Errorf("expected refresh_token opaque-tok, got %s", resp.RefreshToken)
	}

	claims := parseToken(t, resp.AccessToken)
	if sub, _ := claims["sub"].(string); sub != "42" {
		t.Errorf("expected sub 42, got %s", sub)
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))

	body, _ := json.Marshal(map[string]string{"username": "john", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_Login_BadBody(t *testing.T) {
	h := newTestAuthHandler(t, http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Login_UserServiceDown(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))

	body, _ := json.Marshal(map[string]string{"username": "john", "password": "pw"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

// --- Refresh ---

func TestAuthHandler_Refresh_Success(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"user_id": 7})
	}))

	body, _ := json.Marshal(map[string]string{"refresh_token": "some-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp struct {
		AccessToken string `json:"access_token"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	claims := parseToken(t, resp.AccessToken)
	if sub, _ := claims["sub"].(string); sub != "7" {
		t.Errorf("expected sub 7, got %s", sub)
	}
}

func TestAuthHandler_Refresh_InvalidToken(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))

	body, _ := json.Marshal(map[string]string{"refresh_token": "bad"})
	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_Refresh_BadBody(t *testing.T) {
	h := newTestAuthHandler(t, http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	h.Refresh(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- Logout ---

func TestAuthHandler_Logout_Success(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	body, _ := json.Marshal(map[string]string{"refresh_token": "some-token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestAuthHandler_Logout_BadBody(t *testing.T) {
	h := newTestAuthHandler(t, http.NotFoundHandler())

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Logout_UserServiceDown(t *testing.T) {
	h := newTestAuthHandler(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))

	body, _ := json.Marshal(map[string]string{"refresh_token": "token"})
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Logout(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}
