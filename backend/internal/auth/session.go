// Package auth handles Google ID token verification and ORBIT's own
// stateless session tokens (self-signed JWTs, no server-side session
// storage - keeps the approved database schema untouched).
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionTTL = 30 * 24 * time.Hour

var ErrInvalidSession = errors.New("invalid or expired session token")

type SessionIssuer struct {
	secret []byte
}

func NewSessionIssuer(secret string) *SessionIssuer {
	return &SessionIssuer{secret: []byte(secret)}
}

// IssueSessionToken creates a signed session token for the given user id,
// valid for sessionTTL from now.
func (s *SessionIssuer) IssueSessionToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(sessionTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("signing session token: %w", err)
	}
	return signed, nil
}

// ParseSessionToken validates a session token and returns the user id it
// was issued for. Returns ErrInvalidSession for anything wrong with it:
// bad signature, expired, malformed.
func (s *SessionIssuer) ParseSessionToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidSession
	}
	if claims.Subject == "" {
		return "", ErrInvalidSession
	}
	return claims.Subject, nil
}
