package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DevAnuragT/context_keeper/internal/models"
	"github.com/DevAnuragT/context_keeper/internal/services"
)

// UserContextKey is the key for storing user in request context
type UserContextKey struct{}

// AuthRequired is a middleware that requires JWT authentication
func AuthRequired(authSvc services.AuthService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Authorization header required")
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid authorization header format")
			return
		}

		token := parts[1]
		if token == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Token required")
			return
		}

		// Validate token (will be implemented in task 3.2)
		if authSvc == nil {
			writeError(w, http.StatusNotImplemented, "not_implemented", "Authentication not implemented yet")
			return
		}

		user, err := authSvc.ValidateJWT(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", "Invalid token")
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext extracts the user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserContextKey{}).(*models.User)
	return user, ok
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := models.ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    status,
	}
	json.NewEncoder(w).Encode(response)
}
