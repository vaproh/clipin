package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Env         string
	DatabaseURL string
	RedisURL    string
}

func Load() (*Config, error) {
	// Attempt to load .env file if present, ignore error if missing
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		Env:         getEnv("ENV", "development"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://clipin:clipin_dev_pass@localhost:5433/clipin_dev?sslmode=disable"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6380/0"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
