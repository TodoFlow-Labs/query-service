package config

import (
	"fmt"
	"os"
)

type Config struct {
	LogLevel    string
	MetricsAddr string
	HTTPAddr    string
	DatabaseURL string
}

func Load() (*Config, error) {
	cfg := &Config{
		LogLevel:    os.Getenv("LOG_LEVEL"),
		MetricsAddr: os.Getenv("METRICS_ADDR"),
		HTTPAddr:    os.Getenv("HTTP_ADDR"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	// Validation
	var missing []string

	if cfg.LogLevel == "" {
		missing = append(missing, "LOG_LEVEL")
	}
	if cfg.MetricsAddr == "" {
		missing = append(missing, "METRICS_ADDR")
	}
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.HTTPAddr == "" {
		missing = append(missing, "HTTP_ADDR")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required env vars: %v", missing)
	}

	return cfg, nil
}
