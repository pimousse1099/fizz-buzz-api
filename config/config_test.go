package config_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// Each test sets env vars directly with t.Setenv (which forbids t.Parallel and
// is cleaned up per test), so these tests are intentionally not parallel.

func TestNew_Defaults(t *testing.T) {
	t.Setenv("ENV_TYPE", "test")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("FIZZBUZZ_MAX_SEQUENCE_LENGTH", "10000")
	t.Setenv("LOG_LEVEL", "info")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTP.ReadHeaderTimeout != 2*time.Second {
		t.Errorf("ReadHeaderTimeout = %v, want 2s", cfg.HTTP.ReadHeaderTimeout)
	}

	if cfg.HTTP.RateLimitRequests != 100 {
		t.Errorf("RateLimitRequests = %d, want 100", cfg.HTTP.RateLimitRequests)
	}

	if cfg.HTTP.RateLimitWindow != time.Minute {
		t.Errorf("RateLimitWindow = %v, want 1m", cfg.HTTP.RateLimitWindow)
	}

	if cfg.Observability.LogLevel != slog.LevelInfo {
		t.Errorf("LogLevel = %v, want INFO", cfg.Observability.LogLevel)
	}
}

func TestNew_RequiredAndOverrides(t *testing.T) {
	t.Setenv("ENV_TYPE", "test")
	t.Setenv("ENV_NAME", "blue")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("HTTP_WRITE_TIMEOUT", "30s")
	t.Setenv("FIZZBUZZ_MAX_SEQUENCE_LENGTH", "42")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch {
	case cfg.HTTP.Addr != ":9090":
		t.Errorf("Addr = %q, want :9090", cfg.HTTP.Addr)
	case cfg.HTTP.WriteTimeout != 30*time.Second:
		t.Errorf("WriteTimeout = %v, want 30s", cfg.HTTP.WriteTimeout)
	case cfg.FizzBuzz.MaxSequenceLength != 42:
		t.Errorf("MaxSequenceLength = %d, want 42", cfg.FizzBuzz.MaxSequenceLength)
	case cfg.Observability.LogLevel != slog.LevelDebug:
		t.Errorf("LogLevel = %v, want DEBUG", cfg.Observability.LogLevel)
	case cfg.Env.Name != "blue":
		t.Errorf("Env.Name = %q, want blue", cfg.Env.Name)
	}
}

func TestNew_MissingRequiredFails(t *testing.T) {
	// All required vars set EXCEPT FIZZBUZZ_MAX_SEQUENCE_LENGTH (left absent),
	// so go-envconfig's `required` must fail.
	t.Setenv("ENV_TYPE", "test")
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "info")

	if _, err := config.New(context.Background()); err == nil {
		t.Fatal("expected an error when a required variable is missing")
	}
}
