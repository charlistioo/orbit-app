package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

const jwksCacheTTL = 1 * time.Hour

// publicKey returns the RSA public key for the given key id, fetching
// (and caching) Google's current key set as needed.
func (v *RealGoogleVerifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	needsRefresh := v.keys == nil || time.Now().After(v.keysUntil)
	v.mu.Unlock()

	if needsRefresh {
		if err := v.refreshKeys(ctx); err != nil {
			return nil, fmt.Errorf("fetching Google signing keys: %w", err)
		}
	}

	v.mu.Lock()
	defer v.mu.Unlock()
	parts, ok := v.keys[kid]
	if !ok {
		return nil, fmt.Errorf("no matching Google signing key for kid %q", kid)
	}
	return &rsa.PublicKey{N: parts.n, E: parts.e}, nil
}

func (v *RealGoogleVerifier) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleCertsURL, nil)
	if err != nil {
		return err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var set jwkSet
	if err := json.Unmarshal(body, &set); err != nil {
		return fmt.Errorf("parsing JWKS response: %w", err)
	}

	keys := make(map[string]*rsaPublicKeyParts, len(set.Keys))
	for _, k := range set.Keys {
		if k.Kty != "RSA" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := new(big.Int).SetBytes(eBytes)
		keys[k.Kid] = &rsaPublicKeyParts{n: n, e: int(e.Int64())}
	}

	v.mu.Lock()
	v.keys = keys
	v.keysUntil = time.Now().Add(jwksCacheTTL)
	v.mu.Unlock()

	return nil
}
