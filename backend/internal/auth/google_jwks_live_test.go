package auth

import (
	"context"
	"testing"
)

// TestRealGoogleVerifier_CanFetchRealJWKS hits Google's actual JWKS
// endpoint (network required) to confirm the fetch/parse code works
// against Google's real response shape - it deliberately does not try
// to verify a signed token, since producing a genuine Google-signed ID
// token requires an interactive login this test cannot perform.
func TestRealGoogleVerifier_CanFetchRealJWKS(t *testing.T) {
	v := NewRealGoogleVerifier("dummy-client-id")

	err := v.refreshKeys(context.Background())
	if err != nil {
		t.Fatalf("refreshKeys against real Google JWKS endpoint failed: %v", err)
	}

	v.mu.Lock()
	keyCount := len(v.keys)
	v.mu.Unlock()

	if keyCount == 0 {
		t.Fatal("expected at least one signing key from Google's JWKS endpoint")
	}
	t.Logf("fetched %d real Google signing keys", keyCount)
}
