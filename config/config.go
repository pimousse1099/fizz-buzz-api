// Package config loads the service configuration from environment variables.
package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// AppName is the service name used in logs.
const AppName = "fizz-buzz-api"

// AppVersion is set at build time via -ldflags; defaults to DEV.
//
//nolint:gochecknoglobals // build-time version, the single allowed global var
var AppVersion = "DEV"

// Config holds all runtime configuration, populated from the environment.
type Config struct {
	HTTPAddr          string        `env:"HTTP_ADDR, default=:8080"`
	MaxLimit          int           `env:"MAX_LIMIT, default=10000"`
	RateLimitPerSec   float64       `env:"RATE_LIMIT_PER_SEC, default=50"`
	RateLimitBurst    int           `env:"RATE_LIMIT_BURST, default=100"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT, default=2s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT, default=10s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT, default=120s"`
	LogLevel          string        `env:"LOG_LEVEL, default=info"`
}

// New loads configuration from the environment, applying defaults.
func New(ctx context.Context) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return cfg, nil
}
