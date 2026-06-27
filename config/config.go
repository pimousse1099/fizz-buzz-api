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
	Env      Env      `env:",prefix=ENV_"`
	HTTP     HTTP     `env:",prefix=HTTP_"`
	FizzBuzz FizzBuzz `env:",prefix=FIZZBUZZ_"`
	Log      Log      `env:",prefix=LOG_"`
	Tracing  Tracing  `env:",prefix=TRACING_"`
}

// Env identifies the deployment environment, used to tag logs/observability.
type Env struct {
	Type string `env:"TYPE,required"` // ENV_TYPE  e.g. production, staging, dev
	Name string `env:"NAME"`          // ENV_NAME  optional free-form instance label
}

// HTTP holds the HTTP server and edge (rate-limit) configuration.
//
// The four timeouts cover different layers and are not interchangeable:
//
//   - ReadHeaderTimeout — connection-level: max time to read the request headers.
//     The main guard against slowloris-style header stalls.
//   - WriteTimeout — connection-level: deadline from the end of header read until
//     the response write finishes. Protects the socket, but does NOT cancel the
//     handler (the goroutine keeps running); keep it >= RequestTimeout.
//   - IdleTimeout — connection-level: how long a kept-alive connection may sit
//     idle between requests before being closed. Bounds idle conns, not handlers.
//   - RequestTimeout — application-level (chi middleware.Timeout): per-request
//     deadline that cancels the request context and returns 504 if a handler
//     outlives it. Unlike WriteTimeout it propagates cancellation to ctx-aware
//     work (e.g. a slow store call); keep it < WriteTimeout so the 504 is written
//     before the socket write deadline trips.
type HTTP struct {
	Addr string `env:"ADDR,required"` // HTTP_ADDR — listen address, e.g. ":8080"

	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT,default=2s"` // HTTP_READ_HEADER_TIMEOUT
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT,default=10s"`      // HTTP_WRITE_TIMEOUT
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT,default=120s"`      // HTTP_IDLE_TIMEOUT
	RequestTimeout    time.Duration `env:"REQUEST_TIMEOUT,default=5s"`     // HTTP_REQUEST_TIMEOUT

	RateLimitRequests int           `env:"RATE_LIMIT_REQUESTS,default=100"` // HTTP_RATE_LIMIT_REQUESTS — N requests per window, per client IP
	RateLimitWindow   time.Duration `env:"RATE_LIMIT_WINDOW,default=1m"`    // HTTP_RATE_LIMIT_WINDOW
}

// FizzBuzz holds the business configuration.
type FizzBuzz struct {
	// MaxSequenceLength is the inclusive upper bound accepted for the `limit`
	// query parameter; it caps the size of a generated sequence.
	MaxSequenceLength int `env:"MAX_SEQUENCE_LENGTH,required"` // FIZZBUZZ_MAX_SEQUENCE_LENGTH
}

// Log holds logging/observability configuration.
type Log struct {
	// Level is parsed directly into a slog.Level by go-envconfig via the
	// type's encoding.TextUnmarshaler (accepts debug/info/warn/error).
	Level slog.Level `env:"LEVEL,required"` // LOG_LEVEL
}

// Tracing holds OpenTelemetry tracing configuration. Disabled by default so the
// app runs with no collector; HTTP-perf metrics are delegated to infra (§2.10),
// only tracing is instrumented in-app.
type Tracing struct {
	Enabled     bool    `env:"ENABLED,default=false"`  // TRACING_ENABLED
	SampleRatio float64 `env:"SAMPLE_RATIO,default=1"` // TRACING_SAMPLE_RATIO (0..1)
	// OTLPEndpoint is the OTLP/HTTP collector host:port; empty falls back to the
	// standard OTEL_EXPORTER_OTLP_ENDPOINT env var / SDK default.
	OTLPEndpoint string `env:"OTLP_ENDPOINT"` // TRACING_OTLP_ENDPOINT
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
