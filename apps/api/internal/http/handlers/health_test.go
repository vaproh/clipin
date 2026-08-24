package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"clipin/apps/api/internal/http/handlers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

type mockDeps struct{}

func (m *mockDeps) CheckDB(ctx context.Context) bool    { return true }
func (m *mockDeps) CheckRedis(ctx context.Context) bool { return true }

func TestHealthHandler(t *testing.T) {
	router := chi.NewRouter()
	humaConfig := huma.DefaultConfig("ClipIN API", "1.0.0")
	api := humachi.New(router, humaConfig)

	handlers.RegisterHealthHandler(api, &mockDeps{}, "test")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", body["status"])
	}
	if body["service"] != "clipin-api" {
		t.Errorf("expected service 'clipin-api', got '%v'", body["service"])
	}
}
