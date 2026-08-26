package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwksTTL = 5 * time.Minute

// Claims carries the identity extracted from a verified Clerk JWT.
type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

type jwksKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

// JWKSProvider fetches Clerk's signing keys from the JWKS endpoint and caches
// them in memory with a TTL.
type JWKSProvider struct {
	jwksURL    string
	httpClient *http.Client

	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
}

func NewJWKSProvider(jwksURL string) *JWKSProvider {
	return &JWKSProvider{
		jwksURL:    jwksURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       make(map[string]*rsa.PublicKey),
	}
}

// key returns the RSA public key for kid, refreshing the cache when stale.
func (p *JWKSProvider) key(kid string) (*rsa.PublicKey, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Fetch happens under the mutex, so concurrent requests for a fresh JWKS
	// briefly serialize. Refresh is lazy: first request after the TTL expires
	// refetches. Upgrade to singleflight/background refresh if this ever shows
	// up in profiling.
	if time.Since(p.fetchedAt) > jwksTTL {
		if err := p.refresh(); err != nil {
			return nil, err
		}
	}

	key, ok := p.keys[kid]
	if !ok {
		return nil, fmt.Errorf("no signing key for kid %q", kid)
	}
	return key, nil
}

func (p *JWKSProvider) refresh() error {
	resp, err := p.httpClient.Get(p.jwksURL)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	keys := make(map[string]*rsa.PublicKey, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Use != "sig" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return fmt.Errorf("decode JWKS key %q modulus: %w", k.Kid, err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return fmt.Errorf("decode JWKS key %q exponent: %w", k.Kid, err)
		}
		keys[k.Kid] = &rsa.PublicKey{
			N: new(big.Int).SetBytes(nBytes),
			E: int(new(big.Int).SetBytes(eBytes).Int64()),
		}
	}

	p.keys = keys
	p.fetchedAt = time.Now()
	return nil
}

// AuthMiddleware returns middleware that verifies Clerk-issued JWTs and stores
// the claims in the request context. Requests pass through untouched when no
// JWKS URL is configured (development mode), so startup is never blocked.
func AuthMiddleware(provider *JWKSProvider) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if provider == nil || provider.jwksURL == "" {
				next.ServeHTTP(w, r)
				return
			}

			raw, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method %q", t.Method.Alg())
				}
				kid, _ := t.Header["kid"].(string)
				return provider.key(kid)
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithClaims(r.Context(), claims)))
		})
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	return header[len(prefix):], true
}
