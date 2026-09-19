package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/store"
)

type fakeProfileStore struct {
	user         store.User
	found        bool
	lastUpdate   store.ProfileUpdate
	updateCalled bool
}

func (f *fakeProfileStore) GetByID(ctx context.Context, userID string) (store.User, bool, error) {
	return f.user, f.found, nil
}

func (f *fakeProfileStore) UpdateProfile(ctx context.Context, userID string, update store.ProfileUpdate) (store.User, error) {
	f.updateCalled = true
	f.lastUpdate = update
	if update.DisplayName != nil {
		f.user.DisplayName = *update.DisplayName
	}
	if update.Age != nil {
		f.user.Age = update.Age
	}
	if update.DefaultBaselineAmount != nil {
		f.user.DefaultBaselineAmount = update.DefaultBaselineAmount
	}
	f.user.OnboardingCompleted = true
	return f.user, nil
}

type fakeProfileBankStore struct {
	balance float64
}

func (f *fakeProfileBankStore) CurrentBalance(ctx context.Context, userID string) (float64, error) {
	return f.balance, nil
}

func TestGetProfile_ReturnsProfileWithBadge(t *testing.T) {
	users := &fakeProfileStore{user: store.User{ID: "user-1", Email: "a@example.com", DisplayName: "Rani"}, found: true}
	bank := &fakeProfileBankStore{balance: 250000}
	h := &ProfileHandler{Users: users, Bank: bank}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/profile", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp profileView
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Badge != "Thrifty" {
		t.Errorf("got badge %q, want Thrifty for balance 250000", resp.Badge)
	}
}

func TestGetProfile_NotFoundIs404(t *testing.T) {
	users := &fakeProfileStore{found: false}
	h := &ProfileHandler{Users: users, Bank: &fakeProfileBankStore{}}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/profile", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.GetProfile(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateProfile_PartialUpdateAndCompletesOnboarding(t *testing.T) {
	users := &fakeProfileStore{user: store.User{ID: "user-1"}, found: true}
	h := &ProfileHandler{Users: users, Bank: &fakeProfileBankStore{}}

	age := 21
	baseline := 90000.0
	body, _ := json.Marshal(updateProfileRequest{DisplayName: strPtr("Rani"), Age: &age, DefaultBaselineAmount: &baseline})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/profile", bytes.NewReader(body))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if !users.updateCalled {
		t.Fatalf("expected UpdateProfile to be called on the store")
	}
	var resp profileView
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if !resp.OnboardingCompleted {
		t.Errorf("expected onboarding_completed=true after the first profile update")
	}
}

func TestUpdateProfile_RejectsUnrealisticAge(t *testing.T) {
	users := &fakeProfileStore{user: store.User{ID: "user-1"}, found: true}
	h := &ProfileHandler{Users: users, Bank: &fakeProfileBankStore{}}

	age := 999
	body, _ := json.Marshal(updateProfileRequest{Age: &age})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/profile", bytes.NewReader(body))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if users.updateCalled {
		t.Errorf("store should not be called for an invalid age")
	}
}

func TestUpdateProfile_RejectsNegativeBaseline(t *testing.T) {
	users := &fakeProfileStore{user: store.User{ID: "user-1"}, found: true}
	h := &ProfileHandler{Users: users, Bank: &fakeProfileBankStore{}}

	baseline := -1000.0
	body, _ := json.Marshal(updateProfileRequest{DefaultBaselineAmount: &baseline})
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/profile", bytes.NewReader(body))
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUpdateProfile_Unauthenticated(t *testing.T) {
	h := &ProfileHandler{Users: &fakeProfileStore{}, Bank: &fakeProfileBankStore{}}

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/me/profile", bytes.NewReader([]byte(`{}`)))
	rec := httptest.NewRecorder()

	h.UpdateProfile(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func strPtr(s string) *string { return &s }
