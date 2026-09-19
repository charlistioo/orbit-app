// Package api contains ORBIT's HTTP handlers.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// UserStore is the subset of store.UserStore this handler needs -
// defined as an interface so tests can supply a fake with no database.
type UserStore interface {
	FindByGoogleID(ctx context.Context, googleID string) (store.User, bool, error)
	Create(ctx context.Context, googleID, email, displayName string) (store.User, error)
	GetByID(ctx context.Context, userID string) (store.User, bool, error)
}

type AuthHandler struct {
	Verifier auth.GoogleVerifier
	Users    UserStore
	Sessions *auth.SessionIssuer
}

type googleSignInRequest struct {
	IDToken string `json:"id_token"`
}

type googleSignInResponse struct {
	SessionToken string   `json:"session_token"`
	User         userView `json:"user"`
}

type userView struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h *AuthHandler) GoogleSignIn(w http.ResponseWriter, r *http.Request) {
	var req googleSignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		writeJSONError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	claims, err := h.Verifier.Verify(r.Context(), req.IDToken)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidGoogleToken) {
			writeJSONError(w, http.StatusUnauthorized, "invalid or expired Google ID token")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to verify Google ID token")
		return
	}

	user, found, err := h.Users.FindByGoogleID(r.Context(), claims.GoogleID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to look up user")
		return
	}
	if !found {
		user, err = h.Users.Create(r.Context(), claims.GoogleID, claims.Email, claims.DisplayName)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	}

	sessionToken, err := h.Sessions.IssueSessionToken(user.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to issue session token")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(googleSignInResponse{
		SessionToken: sessionToken,
		User: userView{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	})
}

// Me handles GET /api/v1/me - confirms the session token is valid and
// reports whether onboarding (Nama/Umur/Budget) still needs completing,
// so the web app knows whether to route a freshly-signed-in user into
// onboarding or straight to the dashboard.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, found, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to look up user")
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "user not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user_id":              userID,
		"onboarding_completed": user.OnboardingCompleted,
	})
}
