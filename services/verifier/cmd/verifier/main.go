package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clipin/services/verifier/internal/api"
	"clipin/services/verifier/internal/config"
	verifierhttp "clipin/services/verifier/internal/http"
	"clipin/services/verifier/internal/poller"
	"clipin/services/verifier/internal/provider"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	client := api.NewClient(cfg.APIBase, cfg.APIKey, cfg.HTTPTimeout)

	p := &poller.Poller{
		Client:      client,
		YouTubeKey:  cfg.YouTubeAPIKey,
		BatchSize:   cfg.BatchSize,
		Logger:      logger,
		HTTPTimeout: cfg.HTTPTimeout,
	}
	p.ProviderFor = func(postURL string) provider.Provider {
		return provider.RouteToProvider(postURL, cfg.YouTubeAPIKey, &http.Client{Timeout: cfg.HTTPTimeout})
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	verifierhttp.RegisterRoutes(r, cfg.Env)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start poller in background
	go p.Run(ctx, cfg.PollInterval)

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("verifier starting", "port", cfg.Port, "env", cfg.Env)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	case sig := <-shutdown:
		slog.Info("shutdown signal received", "signal", sig)
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				slog.Error("forced server close failed", "error", err)
				os.Exit(1)
			}
		}
	}

	slog.Info("verifier stopped")
}
