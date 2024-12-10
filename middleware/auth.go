package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"loan-engine/config"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	// In a real implementation, this would be connected to your auth service
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// In a real implementation, validate the token and get user info
		// For now, we'll just pass the token as user ID
		ctx := context.WithValue(r.Context(), "user_id", parts[1])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
		if err != nil {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}
		cfg := config.LoadConfig()

		expectedUsername := cfg.AuthUsername
		expectedPassword := cfg.AuthPassword

		if subtle.ConstantTimeCompare([]byte(pair[0]), []byte(expectedUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pair[1]), []byte(expectedPassword)) != 1 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
