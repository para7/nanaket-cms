package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/para7/nanaket-cms/internal/db"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserContextKey is the key for storing user in context
	UserContextKey ContextKey = "user"
	// CookieName is the name of the auth token cookie
	CookieName = "auth_token"
)

// AuthMiddleware creates a middleware that validates access tokens
// It checks Authorization header first, then falls back to cookie
func AuthMiddleware(queries db.Querier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			// Validate token using GetUserByToken
			user, err := queries.GetUserByToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
					return
				}
				log.Printf("Error validating token: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Store user in context
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken extracts the token from Authorization header or cookie
// Priority: 1. Authorization header (Bearer token) 2. Cookie (auth_token)
func extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Expected format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return strings.TrimSpace(parts[1])
		}
	}

	// Fall back to cookie
	cookie, err := r.Cookie(CookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(ctx context.Context) (db.User, bool) {
	user, ok := ctx.Value(UserContextKey).(db.User)
	return user, ok
}
