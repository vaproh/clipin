package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Env           string
	APIBase       string
	APIKey        string
	PollInterval  time.Duration
	BatchSize     int
	YouTubeAPIKey string
	HTTPTimeout   time.Duration
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	_ = godotenv.Load("../../.env")

	pollInterval, err := parseDuration(getEnv("VERIFIER_POLL_INTERVAL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("VERIFIER_POLL_INTERVAL: %w", err)
	}

	batchSize, err := parseInt(getEnv("VERIFIER_BATCH_SIZE", "20"))
	if err != nil {
		return nil, fmt.Errorf("VERIFIER_BATCH_SIZE: %w", err)
	}
	if batchSize < 1 || batchSize > 100 {
		return nil, fmt.Errorf("VERIFIER_BATCH_SIZE must be between 1 and 100")
	}

	httpTimeout, err := parseDuration(getEnv("VERIFIER_HTTP_TIMEOUT", "30s"))
	if err != nil {
		return nil, fmt.Errorf("VERIFIER_HTTP_TIMEOUT: %w", err)
	}

	cfg := &Config{
		Port:          getEnv("VERIFIER_PORT", "8081"),
		Env:           getEnv("VERIFIER_ENV", "development"),
		APIBase:       getEnv("VERIFIER_API_BASE", "http://localhost:8080"),
		APIKey:        getEnv("VERIFIER_API_KEY", ""),
		PollInterval:  pollInterval,
		BatchSize:     batchSize,
		YouTubeAPIKey: getEnv("YOUTUBE_API_KEY", ""),
		HTTPTimeout:   httpTimeout,
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func parseDuration(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration %q must be positive", s)
	}
	return d, nil
}

func parseInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return n, nil
}
