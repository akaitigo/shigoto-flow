package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthWithSecret(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				writeUnauthorized(w)
				return
			}

			claims, err := ValidateToken(secret, token)
			if err != nil {
				writeUnauthorized(w)
				return
			}

			ctx := WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken reads the JWT from the session_token cookie first,
// then falls back to the Authorization: Bearer header.
func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie("session_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

// WithUserID returns a copy of ctx that carries the given user ID. It is the
// counterpart to UserIDFromContext, used by the auth middleware and by tests
// that need to simulate an authenticated request.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}

func isPublicPath(path string) bool {
	publicPaths := map[string]bool{
		"/api/v1/health": true,
	}
	if publicPaths[path] {
		return true
	}
	// OAuth auth paths are public (login flow)
	if strings.HasPrefix(path, "/api/v1/auth/") {
		return true
	}
	return false
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"authentication required","code":"UNAUTHORIZED"}`))
}
