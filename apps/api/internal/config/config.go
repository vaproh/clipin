package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	Env                   string
	DatabaseURL           string
	RedisURL              string
	ClerkJWKSURL          string
	VerifierAPIKey        string
	AllowedOrigins        []string
	RazorpayWebhookSecret string
}

func Load() (*Config, error) {
	// Attempt to load .env file if present, ignore error if missing
	_ = godotenv.Load()

	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		Env:                   getEnv("ENV", "development"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://clipin:clipin_dev_pass@localhost:5433/clipin_dev?sslmode=disable"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6380/0"),
		ClerkJWKSURL:          getEnv("CLERK_JWKS_URL", ""),
		VerifierAPIKey:        getEnv("VERIFIER_API_KEY", ""),
		AllowedOrigins:        parseCSV(getEnv("ALLOWED_ORIGINS", "http://localhost:3000")),
		RazorpayWebhookSecret: getEnv("RAZORPAY_WEBHOOK_SECRET", ""),
	}

	return cfg, nil
}

// parseCSV splits a comma-separated env value into a trimmed, non-empty list.
func parseCSV(raw string) []string {
	var items []string
	for _, item := range strings.Split(raw, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
