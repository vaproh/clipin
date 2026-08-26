package auth_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"clipin/apps/api/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

const testKid = "test-key"

func signToken(t *testing.T, priv *rsa.PrivateKey, sub string, exp time.Time) string {
	t.Helper()
	claims := auth.Claims{
		Sub: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKid
	signed, err := token.SignedString(priv)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func jwksHandler(pub *rsa.PublicKey) http.Handler {
	payload := map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"kid": testKid,
			"alg": "RS256",
			"use": "sig",
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}},
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(payload)
	})
}

func newTestProvider(t *testing.T) (*auth.JWKSProvider, *rsa.PrivateKey, func()) {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	server := httptest.NewServer(jwksHandler(&priv.PublicKey))
	return auth.NewJWKSProvider(server.URL), priv, server.Close
}

func withMiddleware(provider *auth.JWKSProvider, identity *string) http.Handler {
	return auth.AuthMiddleware(provider)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := auth.UserIDFromContext(r.Context()); ok {
			*identity = uid
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func doRequest(t *testing.T, handler http.Handler, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestMiddlewareAcceptsValidToken(t *testing.T) {
	provider, priv, close := newTestProvider(t)
	defer close()

	var identity string
	rec := doRequest(t, withMiddleware(provider, &identity),
		signToken(t, priv, "user_2abc123", time.Now().Add(time.Hour)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if identity != "user_2abc123" {
		t.Errorf("expected user_2abc123 in context, got %q", identity)
	}
}

func TestMiddlewareRejectsExpiredToken(t *testing.T) {
	provider, priv, close := newTestProvider(t)
	defer close()

	var identity string
	rec := doRequest(t, withMiddleware(provider, &identity),
		signToken(t, priv, "user_2abc123", time.Now().Add(-time.Hour)))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if identity != "" {
		t.Errorf("expected no identity, got %q", identity)
	}
}

func TestMiddlewareRejectsMissingToken(t *testing.T) {
	provider, _, close := newTestProvider(t)
	defer close()

	var identity string
	rec := doRequest(t, withMiddleware(provider, &identity), "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsInvalidToken(t *testing.T) {
	provider, _, close := newTestProvider(t)
	defer close()

	var identity string
	rec := doRequest(t, withMiddleware(provider, &identity), "not-a-jwt")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareRejectsTokenSignedByUnknownKey(t *testing.T) {
	provider, _, close := newTestProvider(t)
	defer close()

	// Sign with a key the JWKS endpoint has never published.
	otherPriv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	var identity string
	rec := doRequest(t, withMiddleware(provider, &identity),
		signToken(t, otherPriv, "user_evil", time.Now().Add(time.Hour)))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewareSkipsAuthWhenNoJWKSConfigured(t *testing.T) {
	var identity string
	rec := doRequest(t, withMiddleware(auth.NewJWKSProvider(""), &identity), "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 when auth is disabled, got %d", rec.Code)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	if _, ok := auth.UserIDFromContext(ctx); ok {
		t.Fatal("expected no user in empty context")
	}

	ctx = auth.ContextWithClaims(ctx, &auth.Claims{Sub: "user_abc"})
	uid, ok := auth.UserIDFromContext(ctx)
	if !ok || uid != "user_abc" {
		t.Fatalf("expected user_abc, got %q (ok=%v)", uid, ok)
	}
}
