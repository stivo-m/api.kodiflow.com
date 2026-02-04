package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/stivo-m/api.kodiflow.com/pkg/hashing"
	"github.com/stivo-m/api.kodiflow.com/pkg/helpers"
)

// Checks for the bearer auth token on headers
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Authorization header format must be Bearer {token}", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		userID, err := hashing.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Set user in context
		ctx := context.WithValue(r.Context(), helpers.UserContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
