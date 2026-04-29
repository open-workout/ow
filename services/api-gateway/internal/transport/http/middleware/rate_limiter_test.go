package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/open-workout/ow/services/api-gateway/internal/config"
	mw "github.com/open-workout/ow/services/api-gateway/internal/transport/http/middleware"
)

func TestRateLimiter_Blocking(t *testing.T) {

	cfg := &config.Config{
		RateLimitEnabled: true,
		RateLimitRPS:     1,
	}

	handler := mw.RateLimiter(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// first request should pass
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)

	// second request should fail
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr2.Code)
	}
}
