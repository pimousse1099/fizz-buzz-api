package config_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/pimousse1099/fizz-buzz-api/config"
)

// Each test sets env vars directly with t.Setenv (which forbids t.Parallel and
// is cleaned up per test), so these tests are intentionally not parallel.

func TestNew_Defaults(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("FIZZBUZZ_MAX_LIMIT", "100000")
	t.Setenv("LOG_LEVEL", "info")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch {
	case cfg.HTTP.RateLimit != 20:
		t.Errorf("RateLimit = %d, want 20", cfg.HTTP.RateLimit)
	case cfg.HTTP.BodyLimit != 1<<20:
		t.Errorf("BodyLimit = %d, want %d", cfg.HTTP.BodyLimit, 1<<20)
	case cfg.HTTP.RequestTimeout != 10*time.Second:
		t.Errorf("RequestTimeout = %v, want 10s", cfg.HTTP.RequestTimeout)
	case cfg.HTTP.ShutdownTimeout != 10*time.Second:
		t.Errorf("ShutdownTimeout = %v, want 10s", cfg.HTTP.ShutdownTimeout)
	case cfg.Log.Level != slog.LevelInfo:
		t.Errorf("Log.Level = %v, want INFO", cfg.Log.Level)
	}
}

func TestNew_RequiredAndOverrides(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("HTTP_RATE_LIMIT", "50")
	t.Setenv("HTTP_BODY_LIMIT", "2048")
	t.Setenv("HTTP_REQUEST_TIMEOUT", "30s")
	t.Setenv("HTTP_SHUTDOWN_TIMEOUT", "5s")
	t.Setenv("FIZZBUZZ_MAX_LIMIT", "42")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch {
	case cfg.HTTP.Addr != ":9090":
		t.Errorf("Addr = %q, want :9090", cfg.HTTP.Addr)
	case cfg.HTTP.RateLimit != 50:
		t.Errorf("RateLimit = %d, want 50", cfg.HTTP.RateLimit)
	case cfg.HTTP.BodyLimit != 2048:
		t.Errorf("BodyLimit = %d, want 2048", cfg.HTTP.BodyLimit)
	case cfg.HTTP.RequestTimeout != 30*time.Second:
		t.Errorf("RequestTimeout = %v, want 30s", cfg.HTTP.RequestTimeout)
	case cfg.HTTP.ShutdownTimeout != 5*time.Second:
		t.Errorf("ShutdownTimeout = %v, want 5s", cfg.HTTP.ShutdownTimeout)
	case cfg.FizzBuzz.MaxLimit != 42:
		t.Errorf("MaxLimit = %d, want 42", cfg.FizzBuzz.MaxLimit)
	case cfg.Log.Level != slog.LevelDebug:
		t.Errorf("Log.Level = %v, want DEBUG", cfg.Log.Level)
	}
}

func TestNew_MissingRequiredFails(t *testing.T) {
	// All required vars set EXCEPT FIZZBUZZ_MAX_LIMIT (left absent), so
	// go-envconfig's `required` must fail.
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("LOG_LEVEL", "info")

	if _, err := config.New(context.Background()); err == nil {
		t.Fatal("expected an error when a required variable is missing")
	}
}
