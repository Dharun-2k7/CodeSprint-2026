package middleware

import (
	"codesprint/utils"
	"fmt"
	"net/http"
	"strings"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID, email, err := utils.ValidateJWT(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Store user info in request context
		r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
		r.Header.Set("X-User-Email", email)

		next(w, r)
	}
}

// AdminMiddleware checks if user is admin (for MVP, we'll use created_by check)
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		// For MVP, admin check can be done in handlers
		// This middleware just ensures user is authenticated
		next(w, r)
	})
}

