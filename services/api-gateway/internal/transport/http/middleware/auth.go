package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	UserRoleKey contextKey = "user_role"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
			}

			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			//TODO: replace this with real jwt parsing/validation

			// For now we fake user extraction from token
			userID, role := parseFakeIdentity(token)
			if userID == "" {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))

		}

		return http.HandlerFunc(fn)
	}

}

func parseFakeIdentity(token string) (userID, role string) {
	switch token {
	case "dev-token":
		return "user-123", "user"
	case "admin-token":
		return "1", "admin"
	default:
		return "", ""
	}
}

func GetUserID(r *http.Request) string {
	val := r.Context().Value(UserIDKey)
	if val == nil {
		return ""
	}
	return val.(string)
}

func GetUserRole(r *http.Request) string {
	val := r.Context().Value(UserRoleKey)
	if val == nil {
		return ""
	}
	return val.(string)
}
