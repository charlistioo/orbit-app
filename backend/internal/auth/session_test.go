package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestSessionToken_RoundTrip(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")

	token, err := issuer.IssueSessionToken("user-123")
	if err != nil {
		t.Fatalf("IssueSessionToken failed: %v", err)
	}

	userID, err := issuer.ParseSessionToken(token)
	if err != nil {
		t.Fatalf("ParseSessionToken failed: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("got user id %q, want %q", userID, "user-123")
	}
}

func TestSessionToken_RejectsWrongSecret(t *testing.T) {
	issuer := NewSessionIssuer("secret-a")
	token, _ := issuer.IssueSessionToken("user-123")

	otherIssuer := NewSessionIssuer("secret-b")
	_, err := otherIssuer.ParseSessionToken(token)
	if err != ErrInvalidSession {
		t.Errorf("got err %v, want ErrInvalidSession", err)
	}
}

func TestSessionToken_RejectsExpired(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")

	claims := jwt.RegisteredClaims{
		Subject:   "user-123",
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := expiredToken.SignedString(issuer.secret)
	if err != nil {
		t.Fatalf("signing expired token: %v", err)
	}

	_, err = issuer.ParseSessionToken(signed)
	if err != ErrInvalidSession {
		t.Errorf("got err %v, want ErrInvalidSession", err)
	}
}

func TestSessionToken_RejectsGarbage(t *testing.T) {
	issuer := NewSessionIssuer("test-secret")

	_, err := issuer.ParseSessionToken("not-a-real-token")
	if err != ErrInvalidSession {
		t.Errorf("got err %v, want ErrInvalidSession", err)
	}
}
