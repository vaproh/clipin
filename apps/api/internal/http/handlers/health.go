package handlers

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type HealthCheckDependencies interface {
	CheckDB(ctx context.Context) bool
	CheckRedis(ctx context.Context) bool
}

type HealthOutput struct {
	Body struct {
		Status      string            `json:"status" example:"ok"`
		Service     string            `json:"service" example:"clipin-api"`
		Environment string            `json:"environment" example:"development"`
		Checks      map[string]string `json:"checks,omitempty"`
	}
}

func RegisterHealthHandler(api huma.API, deps HealthCheckDependencies, env string) {
	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      "GET",
		Path:        "/health",
		Summary:     "API Health Check",
		Description: "Returns status of the API server and database/redis connectivity.",
		Tags:        []string{"Health"},
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		resp := &HealthOutput{}
		resp.Body.Status = "ok"
		resp.Body.Service = "clipin-api"
		resp.Body.Environment = env
		resp.Body.Checks = make(map[string]string)

		if deps != nil {
			if deps.CheckDB(ctx) {
				resp.Body.Checks["postgres"] = "up"
			} else {
				resp.Body.Checks["postgres"] = "down"
				resp.Body.Status = "degraded"
			}

			if deps.CheckRedis(ctx) {
				resp.Body.Checks["redis"] = "up"
			} else {
				resp.Body.Checks["redis"] = "down"
			}
		}

		return resp, nil
	})
}
