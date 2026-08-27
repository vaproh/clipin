package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clipin/apps/api/internal/config"
	"clipin/apps/api/internal/db"
	apphttp "clipin/apps/api/internal/http"
	"clipin/apps/api/internal/payout"
	"clipin/apps/api/internal/redis"
	"clipin/apps/api/internal/service"
	"clipin/apps/api/internal/worker"
)

func main() {
	migrateFlag := flag.Bool("migrate", false, "Run database migrations before starting")
	flag.Parse()

	log.Println("Starting ClipIN API service...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *migrateFlag {
		log.Println("Running migrations...")
		if err := db.MigrateUp(cfg.DatabaseURL); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations applied")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Scaffolding database connection
	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("Warning: Database connection failed (is PostgreSQL running?): %v", err)
	} else {
		defer database.Close()
		log.Println("PostgreSQL connection pool initialized")

		// Seed default campaign templates (idempotent).
		templateSvc := service.NewTemplateService(database.Queries)
		if err := templateSvc.SeedDefaultTemplates(context.Background()); err != nil {
			log.Printf("Warning: Failed to seed default templates: %v", err)
		}

		// Start auto-approve worker with graceful shutdown.
		workerCtx, workerCancel := context.WithCancel(context.Background())
		defer workerCancel()
		submissionSvc := service.NewSubmissionService(database.Queries)
		worker.StartAutoApproveWorker(workerCtx, submissionSvc, 5*time.Minute)

		// Start payout processor worker with graceful shutdown.
		rp := payout.NewRazorpayProviderOrStub(
			cfg.RazorpayKeyID,
			cfg.RazorpayKeySecret,
			cfg.RazorpayAccountNumber,
		)
		payoutSvc := service.NewPayoutService(database.Queries, rp)
		worker.StartPayoutProcessorWorker(workerCtx, payoutSvc, 10*time.Minute)
	}

	// Scaffolding Redis connection
	redisClient, err := redis.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Printf("Warning: Redis connection failed (is Redis running?): %v", err)
	} else {
		defer redisClient.Close()
		log.Println("Redis client initialized")
	}

	deps := &apphttp.AppDependencies{
		Config: cfg,
		DB:     database,
		Redis:  redisClient,
	}

	router := apphttp.NewRouter(deps)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Shutdown handling
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("ClipIN API server running on http://localhost:%s", cfg.Port)
		log.Printf("OpenAPI documentation available at http://localhost:%s/docs", cfg.Port)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Fatalf("Forced server close failed: %v", err)
			}
		}
	}

	log.Println("ClipIN API service stopped clean")
}
