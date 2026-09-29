package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPPort        string
	ShutdownTimeout time.Duration
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
}

func Load() (Config, error) {
	shutdown, err := time.ParseDuration(getEnv("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT: %w", err)
	}

	cfg := Config{
		AppEnv:          getEnv("APP_ENV", "development"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		ShutdownTimeout: shutdown,
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   os.Getenv("REDIS_PASSWORD"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
