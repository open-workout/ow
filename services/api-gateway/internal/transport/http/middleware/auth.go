package middleware

import (
	"context"
	"net/http"
	"net/url"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(issuerURL, audience string) func(http.Handler) http.Handler {
	parsed, err := url.Parse(issuerURL)
	if err != nil {
		panic("auth middleware: invalid issuer URL: " + err.Error())
	}

	provider := jwks.NewCachingProvider(parsed, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL,
		[]string{audience},
	)
	if err != nil {
		panic("auth middleware: failed to create JWT validator: " + err.Error())
	}

	m := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
		}),
	)

	return func(next http.Handler) http.Handler {
		return m.CheckJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
			if !ok || claims.RegisteredClaims.Subject == "" {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, claims.RegisteredClaims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		}))
	}
}

func GetUserID(r *http.Request) string {
	val := r.Context().Value(UserIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}
