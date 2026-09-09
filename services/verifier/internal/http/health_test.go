package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	verifierhttp "clipin/services/verifier/internal/http"
	"clipin/services/verifier/internal/monitor"

	"github.com/go-chi/chi/v5"
)

func TestVerifierHealthHandler(t *testing.T) {
	r := chi.NewRouter()
	verifierhttp.RegisterRoutes(r, "test")

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}

	var resp verifierhttp.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response json: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
	if resp.Service != "clipin-verifier" {
		t.Errorf("expected service 'clipin-verifier', got '%s'", resp.Service)
	}
}

func TestVerifierMetricsHandler(t *testing.T) {
	r := chi.NewRouter()
	m := &monitor.Monitor{}
	m.PollStarted()
	verifierhttp.RegisterRoutes(r, "test", m)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "clipin_verifier_poll_cycles_total 1") {
		t.Fatal("metrics did not contain poll cycle counter")
	}
}
