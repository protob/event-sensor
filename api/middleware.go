package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/protob/event-sensor/internal/auth"
)

type contextKey string

const (
	contextKeyUserID   contextKey = "user_id"
	contextKeyUsername contextKey = "username"
)

// BodyLimit caps request bodies at n bytes. The largest legitimate payload is a
// manual festival with a long lineup - a few tens of KB - so a megabyte leaves
// generous headroom while keeping a runaway client from buffering unbounded JSON.
func BodyLimit(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, n)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// AuthMiddleware returns a Chi middleware that validates JWT tokens.
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := auth.ValidateToken(parts[1], secret)
			if err != nil {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyUserID, claims.Subject)
			ctx = context.WithValue(ctx, contextKeyUsername, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext extracts the user ID from the request context.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyUserID).(string); ok {
		return v
	}
	return ""
}

// UsernameFromContext extracts the username from the request context.
func UsernameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeyUsername).(string); ok {
		return v
	}
	return ""
}
