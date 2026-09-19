package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireAuth_MissingHeader(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")
	called := false
	handler := RequireAuth(issuer, func(w http.ResponseWriter, r *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	if called {
		t.Error("inner handler should not have been called")
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")
	handler := RequireAuth(issuer, func(w http.ResponseWriter, r *http.Request) {})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer garbage")
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuth_ValidToken(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")
	token, _ := issuer.IssueSessionToken("user-42")

	var gotUserID string
	handler := RequireAuth(issuer, func(w http.ResponseWriter, r *http.Request) {
		gotUserID, _ = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	if gotUserID != "user-42" {
		t.Errorf("got user id %q, want %q", gotUserID, "user-42")
	}
}
