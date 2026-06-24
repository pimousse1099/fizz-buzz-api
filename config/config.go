// Package config loads the service configuration from environment variables,
// grouped by concern: deployment environment, HTTP/edge, business and
// observability.
package config

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// AppName is the service name used in logs.
const AppName = "fizz-buzz-api"

// AppVersion is set at build time via -ldflags; defaults to DEV.
//
//nolint:gochecknoglobals // build-time version, the single allowed global var
var AppVersion = "DEV"

// Config is the root configuration, grouped by concern.
type Config struct {
	Env           Env           `env:",prefix=ENV_"`
	HTTP          HTTP          `env:",prefix=HTTP_"`
	FizzBuzz      FizzBuzz      `env:",prefix=FIZZBUZZ_"`
	Observability Observability `env:",prefix=LOG_"`
}

// Env identifies the deployment environment, used to tag logs/observability.
type Env struct {
	Type string `env:"TYPE,required"` // ENV_TYPE  e.g. production, staging, dev
	Name string `env:"NAME"`          // ENV_NAME  optional free-form instance label
}

// HTTP holds the HTTP server and edge (rate-limit) configuration.
type HTTP struct {
	Addr              string        `env:"ADDR,required"`                  // HTTP_ADDR
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT,default=2s"` // HTTP_READ_HEADER_TIMEOUT
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT,default=10s"`      // HTTP_WRITE_TIMEOUT
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT,default=120s"`      // HTTP_IDLE_TIMEOUT
	RateLimitPerSec   float64       `env:"RATE_LIMIT_PER_SEC,default=50"`  // HTTP_RATE_LIMIT_PER_SEC
	RateLimitBurst    int           `env:"RATE_LIMIT_BURST,default=100"`   // HTTP_RATE_LIMIT_BURST
}

// FizzBuzz holds the business configuration.
type FizzBuzz struct {
	// MaxSequenceLength is the inclusive upper bound accepted for the `limit`
	// query parameter; it caps the size of a generated sequence.
	MaxSequenceLength int `env:"MAX_SEQUENCE_LENGTH,required"` // FIZZBUZZ_MAX_SEQUENCE_LENGTH
}

// Observability holds logging/observability configuration.
type Observability struct {
	// LogLevel is parsed directly into a slog.Level by go-envconfig via the
	// type's encoding.TextUnmarshaler (accepts debug/info/warn/error).
	LogLevel slog.Level `env:"LEVEL,required"` // LOG_LEVEL
}

// New loads configuration from the environment. Required variables without a
// value cause an error.
func New(ctx context.Context) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return cfg, nil
}
