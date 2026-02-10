package handlers

import (
	"context"
	"email-reminder-system/models"
	"email-reminder-system/services"
	"fmt"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserContextKey is the key for storing user in request context
	UserContextKey contextKey = "user"
)

// AuthMiddleware creates middleware that checks for valid session cookie
type AuthMiddleware struct {
	authService services.AuthService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth wraps an HTTP handler and requires authentication
// Redirects to login page if session is invalid or missing
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			// No session cookie, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session
		user, err := m.authService.ValidateSession(cookie.Value)
		if err != nil {
			// Invalid or expired session, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Store user information in request context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext retrieves the authenticated user from request context
func GetUserFromContext(r *http.Request) (*models.User, bool) {
	user, ok := r.Context().Value(UserContextKey).(*models.User)
	return user, ok
}

// RecoverMiddleware recovers from panics in HTTP handlers
func RecoverMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger := services.GetLogger()
				logger.Error("HTTP handler panic recovered", map[string]interface{}{
					"panic": fmt.Sprintf("%v", err),
					"path":  r.URL.Path,
					"method": r.Method,
				})
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}
