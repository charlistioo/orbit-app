package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "orbit_user_id"

// RequireAuth wraps a handler so it only runs when the request carries a
// valid ORBIT session token in "Authorization: Bearer <token>". The
// authenticated user id is injected into the request context - handlers
// read it via UserIDFromContext instead of trusting any client-supplied
// user id (this is the security requirement from the technical design:
// never scope by a value the client sends).
func RequireAuth(issuer *SessionIssuer, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			http.Error(w, `{"error":"missing or malformed Authorization header"}`, http.StatusUnauthorized)
			return
		}

		userID, err := issuer.ParseSessionToken(token)
		if err != nil {
			http.Error(w, `{"error":"invalid or expired session token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, userID)
		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext returns the authenticated user id set by RequireAuth.
// Only call this on a request that has passed through RequireAuth.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey).(string)
	return id, ok
}

// ContextWithUserID injects a user id the same way RequireAuth does.
// Exists for handler tests that need an authenticated context without
// going through a real session token and the HTTP middleware.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}
