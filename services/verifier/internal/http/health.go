package http

import (
	"encoding/json"
	"net/http"

	"clipin/services/verifier/internal/monitor"

	"github.com/go-chi/chi/v5"
)

type HealthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

func RegisterRoutes(r chi.Router, env string, monitors ...*monitor.Monitor) {
	var m *monitor.Monitor
	if len(monitors) > 0 {
		m = monitors[0]
	}
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		resp := HealthResponse{
			Status:      "ok",
			Service:     "clipin-verifier",
			Environment: env,
			Description: "Social metrics polling worker",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	})
	r.Get("/metrics", func(w http.ResponseWriter, r *http.Request) {
		if m == nil {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		if err := m.WritePrometheus(w); err != nil {
			http.Error(w, "failed to write metrics", http.StatusInternalServerError)
		}
	})
}
