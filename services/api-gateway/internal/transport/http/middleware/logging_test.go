package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

func TestLoggingMiddleware(t *testing.T) {

	called := false

	handler := mw.Logging()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("handler was not called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
