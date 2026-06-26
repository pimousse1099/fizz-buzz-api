// Command fizz-buzz-api starts the fizz-buzz REST API server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-playground/validator/v10"

	httpserver "github.com/pimousse1099/fizz-buzz-api/internal/http"
	"github.com/pimousse1099/fizz-buzz-api/internal/statsstorer"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	validate := validator.New()
	store := statsstorer.NewInMemory()

	e := httpserver.New(logger, validate, store)

	// Cancel the start context on SIGINT/SIGTERM so the server shuts down gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting server", "address", httpserver.ListenAddress)

	err := httpserver.Run(ctx, e)
	if err != nil {
		// stop() is deferred; it only releases signal handlers, which is moot here.
		logger.Error("server terminated unexpectedly", "error", err)
		os.Exit(1) //nolint:gocritic // exit is intentional on fatal startup error
	}

	logger.Info("server stopped gracefully")
}
