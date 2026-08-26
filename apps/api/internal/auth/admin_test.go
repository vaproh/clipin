package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"clipin/apps/api/internal/auth"
	sqlc "clipin/apps/api/internal/db/sqlc"
)

func adminUser() *sqlc.User {
	return &sqlc.User{
		ID:    "admin_1",
		Email: "admin@test.com",
		Role:  "admin",
	}
}

func TestAdminOnly_Admin_Passes(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := auth.AdminOnly(inner)
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), adminUser()))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminOnly_NonAdmin_Returns403(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})

	handler := auth.AdminOnly(inner)
	user := &sqlc.User{ID: "user_1", Role: "clipper"}
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminOnly_Owner_Returns403(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})

	handler := auth.AdminOnly(inner)
	user := &sqlc.User{ID: "user_1", Role: "owner"}
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	req = req.WithContext(auth.ContextWithUser(req.Context(), user))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestAdminOnly_NoUser_Returns401(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("should not reach handler")
	})

	handler := auth.AdminOnly(inner)
	req := httptest.NewRequest(http.MethodGet, "/admin/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
