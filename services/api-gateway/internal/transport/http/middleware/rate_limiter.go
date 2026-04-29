package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/open-workout/ow/services/api-gateway/internal/config"
)

type client struct {
	tokens     int
	lastAccess time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rps     int
	enabled bool
}

func RateLimiter(cfg *config.Config) func(http.Handler) http.Handler {

	rl := &rateLimiter{
		clients: make(map[string]*client),
		rps:     cfg.RateLimitRPS,
		enabled: cfg.RateLimitEnabled,
	}

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			if !rl.enabled {
				next.ServeHTTP(w, r)
				return
			}

			ip := r.RemoteAddr

			if !rl.allow(ip) {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	c, exists := rl.clients[ip]

	now := time.Now()

	if !exists {
		rl.clients[ip] = &client{
			tokens:     rl.rps - 1,
			lastAccess: now,
		}
		return true
	}

	// simple refill logic (very basic token bucket)
	elapsed := now.Sub(c.lastAccess).Seconds()
	c.tokens += int(elapsed * float64(rl.rps))
	if c.tokens > rl.rps {
		c.tokens = rl.rps
	}

	c.lastAccess = now

	if c.tokens <= 0 {
		return false
	}

	c.tokens--
	return true
}
