package auth

import (
	"context"
	"net/http"
)

// WebSocketMiddleware validates JWT token from query parameter for WebSocket connections
// WebSockets don't support Authorization headers, so we use query parameters
func WebSocketMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from query parameter
			token := r.URL.Query().Get("token")
			if token == "" {
				http.Error(w, "No authentication token provided", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				if err == ErrTokenExpired {
					http.Error(w, "Token expired", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UsernameKey, claims.Username)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
