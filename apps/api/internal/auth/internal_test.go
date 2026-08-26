package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"clipin/apps/api/internal/auth"
)

func TestInternalAuthMiddleware_EmptyKey_PassesThrough(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := auth.InternalAuthMiddleware("")(inner)

	req := httptest.NewRequest(http.MethodGet, "/internal/snapshots", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestInternalAuthMiddleware_MissingHeader_Returns401(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})

	handler := auth.InternalAuthMiddleware("secret-key")(inner)

	req := httptest.NewRequest(http.MethodGet, "/internal/snapshots", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestInternalAuthMiddleware_WrongKey_Returns401(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})

	handler := auth.InternalAuthMiddleware("secret-key")(inner)

	req := httptest.NewRequest(http.MethodGet, "/internal/snapshots", nil)
	req.Header.Set("X-Verifier-Key", "wrong-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestInternalAuthMiddleware_CorrectKey_PassesThrough(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := auth.InternalAuthMiddleware("secret-key")(inner)

	req := httptest.NewRequest(http.MethodGet, "/internal/snapshots", nil)
	req.Header.Set("X-Verifier-Key", "secret-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
