package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	mw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

func makeToken(t *testing.T, secret, subject, role string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  subject,
		"role": role,
		"exp":  time.Now().Add(time.Hour).Unix(),
		"iss":  "open-workout",
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestAuth_ValidToken(t *testing.T) {
	tok := makeToken(t, "secret", "user-123", "user")

	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mw.GetUserID(r)
		if userID != "user-123" {
			t.Errorf("got %s, want user-123", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuth_AdminToken(t *testing.T) {
	tok := makeToken(t, "secret", "1", "admin")

	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if mw.GetUserID(r) != "1" {
			t.Errorf("got userID %s, want 1", mw.GetUserID(r))
		}
		if mw.GetUserRole(r) != "admin" {
			t.Errorf("got role %s, want admin", mw.GetUserRole(r))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAuth_MissingToken(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuth_WrongSecret(t *testing.T) {
	tok := makeToken(t, "other-secret", "user-1", "user")

	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
