// Command fizz-buzz-api starts the fizz-buzz REST API server.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	slogctx "github.com/veqryn/slog-context"

	"github.com/pimousse1099/fizz-buzz-api/config"
	httpserver "github.com/pimousse1099/fizz-buzz-api/internal/http"
	"github.com/pimousse1099/fizz-buzz-api/internal/statsstorer"
)

func main() {
	// Load configuration first: a missing required variable must fail fast,
	// before any server is built. No logger exists yet, so report to stderr.
	cfg, err := config.New(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config:", err)
		os.Exit(1)
	}

	// slogctx.NewHandler stamps each record with attributes carried in the context
	// (request_id, added by the requestID middleware), so *Context log calls are
	// automatically correlated. The level comes from config.
	handler := slogctx.NewHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Log.Level}), nil)
	logger := slog.New(handler)
	store := statsstorer.NewInMemory()

	e := httpserver.New(logger, store, cfg.HTTP, cfg.FizzBuzz.MaxLimit)

	// Cancel the start context on SIGINT/SIGTERM so the server shuts down gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting server", "address", cfg.HTTP.Addr)

	err = httpserver.Run(ctx, e, cfg.HTTP)
	if err != nil {
		// stop() is deferred; it only releases signal handlers, which is moot here.
		logger.Error("server terminated unexpectedly", "error", err)
		os.Exit(1) //nolint:gocritic // exit is intentional on fatal startup error
	}

	logger.Info("server stopped gracefully")
}
