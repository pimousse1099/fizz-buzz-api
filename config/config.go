// Package config loads the service configuration from environment variables,
// grouped by concern (HTTP/edge, business and logging).
package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// Config is the root configuration, grouped by concern. Required variables have
// no default, so a missing value fails fast at startup rather than silently at
// runtime; operational knobs keep sensible defaults.
type Config struct {
	HTTP     HTTP     `env:",prefix=HTTP_"`
	FizzBuzz FizzBuzz `env:",prefix=FIZZBUZZ_"`
	Log      Log      `env:",prefix=LOG_"`
}

// HTTP holds the HTTP server and edge (rate-limit, body-limit) configuration.
type HTTP struct {
	Addr            string        `env:"ADDR,required"`                // HTTP_ADDR              e.g. :8080
	RateLimit       int           `env:"RATE_LIMIT,default=20"`        // HTTP_RATE_LIMIT        requests/second/IP
	BodyLimit       int64         `env:"BODY_LIMIT,default=1048576"`   // HTTP_BODY_LIMIT        max request body, bytes (1 MiB)
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT,default=10s"`  // HTTP_REQUEST_TIMEOUT
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT,default=10s"` // HTTP_SHUTDOWN_TIMEOUT
}

// FizzBuzz holds the business configuration.
type FizzBuzz struct {
	// MaxLimit is the inclusive upper bound accepted for the `limit` query
	// parameter; it caps the size of the generated sequence (DoS guard).
	MaxLimit uint `env:"MAX_LIMIT,required"` // FIZZBUZZ_MAX_LIMIT
}

// Log holds logging configuration.
type Log struct {
	// Level is parsed directly into a slog.Level by go-envconfig via the type's
	// encoding.TextUnmarshaler (accepts debug/info/warn/error).
	Level slog.Level `env:"LEVEL,required"` // LOG_LEVEL
}

// New loads configuration from the environment. Required variables without a
// value cause an error.
func New(ctx context.Context) (*Config, error) {
	cfg := &Config{}

	err := envconfig.Process(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return cfg, nil
}
