package auth

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidGoogleToken = errors.New("invalid or expired Google ID token")

const googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
const googleIssuer = "https://accounts.google.com"

// GoogleClaims is the subset of a verified Google ID token's payload that
// ORBIT needs to create or look up a user.
type GoogleClaims struct {
	GoogleID    string // the token's "sub" claim
	Email       string
	DisplayName string
}

// GoogleVerifier verifies a raw Google ID token and returns the identity
// claims it carries. Implemented as an interface so the HTTP handler can
// be tested without making real calls to Google.
type GoogleVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (GoogleClaims, error)
}

// jwk is one entry from Google's public JWK Set.
type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// RealGoogleVerifier verifies tokens against Google's public signing
// keys, fetched from Google's well-known JWKS endpoint and cached for a
// while - checks signature, expiry, issuer, and audience (must match our
// OAuth Client ID). No client secret is needed for this: verifying an ID
// token's signature only ever requires the *public* key.
type RealGoogleVerifier struct {
	clientID   string
	httpClient *http.Client

	mu        sync.Mutex
	keys      map[string]*rsaPublicKeyParts
	keysUntil time.Time
}

type rsaPublicKeyParts struct {
	n *big.Int
	e int
}

func NewRealGoogleVerifier(clientID string) *RealGoogleVerifier {
	return &RealGoogleVerifier{
		clientID:   clientID,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (v *RealGoogleVerifier) Verify(ctx context.Context, rawIDToken string) (GoogleClaims, error) {
	claims := jwt.MapClaims{}

	_, err := jwt.ParseWithClaims(rawIDToken, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		kid, _ := t.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("token missing kid header")
		}
		key, err := v.publicKey(ctx, kid)
		if err != nil {
			return nil, err
		}
		return key, nil
	}, jwt.WithIssuer(googleIssuer), jwt.WithAudience(v.clientID))

	if err != nil {
		return GoogleClaims{}, fmt.Errorf("%w: %v", ErrInvalidGoogleToken, err)
	}

	sub, _ := claims["sub"].(string)
	if sub == "" {
		return GoogleClaims{}, ErrInvalidGoogleToken
	}
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)

	return GoogleClaims{GoogleID: sub, Email: email, DisplayName: name}, nil
}
