package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

func TestAuth_ValidToken(t *testing.T) {
	handler := mw.Auth("secret")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := mw.GetUserID(r)
		if userID != "user-123" {
			t.Errorf("got %s, want user-123", userID)
		}

		w.WriteHeader(http.StatusOK)

	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer dev-token")
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
