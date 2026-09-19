package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"orbit/backend/internal/auth"
	"orbit/backend/internal/store"
)

// fakeVerifier lets tests control what a "Google ID token" resolves to,
// without ever calling Google.
type fakeVerifier struct {
	claims auth.GoogleClaims
	err    error
}

func (f *fakeVerifier) Verify(ctx context.Context, rawIDToken string) (auth.GoogleClaims, error) {
	return f.claims, f.err
}

// fakeUserStore is an in-memory stand-in for store.UserStore.
type fakeUserStore struct {
	byGoogleID  map[string]store.User
	nextID      int
	createCalls int
}

func newFakeUserStore() *fakeUserStore {
	return &fakeUserStore{byGoogleID: map[string]store.User{}}
}

func (s *fakeUserStore) FindByGoogleID(ctx context.Context, googleID string) (store.User, bool, error) {
	u, ok := s.byGoogleID[googleID]
	return u, ok, nil
}

func (s *fakeUserStore) GetByID(ctx context.Context, userID string) (store.User, bool, error) {
	for _, u := range s.byGoogleID {
		if u.ID == userID {
			return u, true, nil
		}
	}
	return store.User{}, false, nil
}

func (s *fakeUserStore) Create(ctx context.Context, googleID, email, displayName string) (store.User, error) {
	s.createCalls++
	s.nextID++
	u := store.User{
		ID:          "generated-id",
		GoogleID:    googleID,
		Email:       email,
		DisplayName: displayName,
	}
	s.byGoogleID[googleID] = u
	return u, nil
}

func doSignIn(t *testing.T, h *AuthHandler, idToken string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"id_token": idToken})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/google", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.GoogleSignIn(rec, req)
	return rec
}

func TestGoogleSignIn_NewUserCreatesExactlyOneProfile(t *testing.T) {
	users := newFakeUserStore()
	h := &AuthHandler{
		Verifier: &fakeVerifier{claims: auth.GoogleClaims{GoogleID: "g-1", Email: "a@example.com", DisplayName: "A"}},
		Users:    users,
		Sessions: auth.NewSessionIssuer("test-secret"),
	}

	rec := doSignIn(t, h, "whatever-real-token-value")

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if users.createCalls != 1 {
		t.Errorf("got %d Create calls, want exactly 1", users.createCalls)
	}

	var resp googleSignInResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding response: %v", err)
	}
	if resp.SessionToken == "" {
		t.Error("expected a non-empty session token")
	}
}

func TestGoogleSignIn_ExistingUserDoesNotDuplicate(t *testing.T) {
	users := newFakeUserStore()
	users.byGoogleID["g-1"] = store.User{ID: "existing-id", GoogleID: "g-1", Email: "a@example.com"}

	h := &AuthHandler{
		Verifier: &fakeVerifier{claims: auth.GoogleClaims{GoogleID: "g-1", Email: "a@example.com"}},
		Users:    users,
		Sessions: auth.NewSessionIssuer("test-secret"),
	}

	rec := doSignIn(t, h, "whatever-real-token-value")

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	if users.createCalls != 0 {
		t.Errorf("got %d Create calls, want 0 (user already existed)", users.createCalls)
	}
}

func TestGoogleSignIn_InvalidTokenRejectedWith401(t *testing.T) {
	users := newFakeUserStore()
	h := &AuthHandler{
		Verifier: &fakeVerifier{err: auth.ErrInvalidGoogleToken},
		Users:    users,
		Sessions: auth.NewSessionIssuer("test-secret"),
	}

	rec := doSignIn(t, h, "bad-token")

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if users.createCalls != 0 {
		t.Errorf("got %d Create calls, want 0 for a rejected token", users.createCalls)
	}
}

func TestGoogleSignIn_MissingIDTokenIsBadRequest(t *testing.T) {
	users := newFakeUserStore()
	h := &AuthHandler{
		Verifier: &fakeVerifier{},
		Users:    users,
		Sessions: auth.NewSessionIssuer("test-secret"),
	}

	rec := doSignIn(t, h, "")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestMe_ReportsOnboardingCompleted(t *testing.T) {
	users := newFakeUserStore()
	users.byGoogleID["g-1"] = store.User{ID: "user-1", GoogleID: "g-1", Email: "a@example.com", OnboardingCompleted: true}
	h := &AuthHandler{Users: users, Sessions: auth.NewSessionIssuer("test-secret")}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "user-1"))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, body %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["onboarding_completed"] != true {
		t.Errorf("got onboarding_completed=%v, want true", resp["onboarding_completed"])
	}
}

func TestMe_UnknownUserIs404(t *testing.T) {
	users := newFakeUserStore()
	h := &AuthHandler{Users: users, Sessions: auth.NewSessionIssuer("test-secret")}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(contextWithTestUser(req.Context(), "ghost"))
	rec := httptest.NewRecorder()

	h.Me(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGoogleSignIn_UnexpectedVerifierErrorIs500(t *testing.T) {
	users := newFakeUserStore()
	h := &AuthHandler{
		Verifier: &fakeVerifier{err: errors.New("network blip")},
		Users:    users,
		Sessions: auth.NewSessionIssuer("test-secret"),
	}

	rec := doSignIn(t, h, "some-token")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
