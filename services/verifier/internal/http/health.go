package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type HealthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Description string `json:"description"`
}

func RegisterRoutes(r chi.Router, env string) {
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
}
