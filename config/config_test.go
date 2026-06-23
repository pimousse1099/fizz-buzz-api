package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

//nolint:paralleltest // TestNew_FromEnv uses t.Setenv, forbidding parallelism in this package
func TestNew_Defaults(t *testing.T) {
	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}

	if cfg.MaxLimit != 10000 {
		t.Errorf("MaxLimit = %d, want 10000", cfg.MaxLimit)
	}
}

func TestNew_FromEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("MAX_LIMIT", "42")
	t.Setenv("WRITE_TIMEOUT", "30s")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPAddr != ":9090" || cfg.MaxLimit != 42 || cfg.WriteTimeout != 30*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
