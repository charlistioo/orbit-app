package api

import (
	"context"
	"encoding/json"
	"net/http"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// ProfileStore is the subset of store.UserStore this handler needs.
type ProfileStore interface {
	GetByID(ctx context.Context, userID string) (store.User, bool, error)
	UpdateProfile(ctx context.Context, userID string, update store.ProfileUpdate) (store.User, error)
}

// ProfileBankStore is the subset of store.ConsumptionBankStore needed to
// report the current badge tier on the profile.
type ProfileBankStore interface {
	CurrentBalance(ctx context.Context, userID string) (float64, error)
}

type ProfileHandler struct {
	Users ProfileStore
	Bank  ProfileBankStore
}

type profileView struct {
	ID                    string   `json:"id"`
	Email                 string   `json:"email"`
	DisplayName           string   `json:"display_name"`
	Age                   *int     `json:"age"`
	AvatarData            *string  `json:"avatar_data"`
	DefaultBaselineAmount *float64 `json:"default_baseline_amount"`
	Badge                 string   `json:"badge"`
	OnboardingCompleted   bool     `json:"onboarding_completed"`
}

func toProfileView(u store.User, badge string) profileView {
	return profileView{
		ID:                    u.ID,
		Email:                 u.Email,
		DisplayName:           u.DisplayName,
		Age:                   u.Age,
		AvatarData:            u.AvatarData,
		DefaultBaselineAmount: u.DefaultBaselineAmount,
		Badge:                 badge,
		OnboardingCompleted:   u.OnboardingCompleted,
	}
}

// GetProfile handles GET /api/v1/me/profile - the full profile shown on
// the Settings/Profil screen (name, age, avatar, default daily budget,
// current Consumption Bank badge).
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, found, err := h.Users.GetByID(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load profile")
		return
	}
	if !found {
		writeJSONError(w, http.StatusNotFound, "profile not found")
		return
	}

	balance, err := h.Bank.CurrentBalance(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load Consumption Bank balance")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProfileView(user, badgeTierFor(balance)))
}

type updateProfileRequest struct {
	DisplayName           *string  `json:"display_name"`
	Age                   *int     `json:"age"`
	AvatarData            *string  `json:"avatar_data"`
	DefaultBaselineAmount *float64 `json:"default_baseline_amount"`
}

// UpdateProfile handles PATCH /api/v1/me/profile - a partial update used
// both by the one-time onboarding step (Nama/Umur/Budget) and later by
// Settings. Any field left null in the request is left unchanged.
// onboarding_completed becomes true the first time this succeeds (see
// store.UserStore.UpdateProfile), never reset afterward.
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Age != nil && (*req.Age < 1 || *req.Age > 130) {
		writeJSONError(w, http.StatusBadRequest, "age must be a realistic value")
		return
	}
	if req.DefaultBaselineAmount != nil && *req.DefaultBaselineAmount < 0 {
		writeJSONError(w, http.StatusBadRequest, "default_baseline_amount must not be negative")
		return
	}

	user, err := h.Users.UpdateProfile(r.Context(), userID, store.ProfileUpdate{
		DisplayName:           req.DisplayName,
		Age:                   req.Age,
		AvatarData:            req.AvatarData,
		DefaultBaselineAmount: req.DefaultBaselineAmount,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProfileView(user, ""))
}
