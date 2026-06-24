// Command fizz-buzz-api starts the HTTP server.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/di"
)

const shutdownTimeout = 10 * time.Second

func main() {
	slog.Info("starting application",
		"application_name", config.AppName,
		"application_version", config.AppVersion,
		"go_version", runtime.Version(),
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
	)

	// Loaded before the base context so an early exit skips no deferred cleanup.
	cfg, err := config.New(context.Background())
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Base context for the whole app lifecycle: it backs the HTTP server's
	// BaseContext and the shutdown deadline, and is canceled on exit. It is
	// independent of the OS signal so graceful shutdown can still use it.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container := di.NewContainer(ctx, cfg)
	logger := container.GetLogger()

	httpSrv := container.GetHTTPServer()
	errChan := make(chan error, 1)

	logger.Info("starting http server", "addr", cfg.HTTP.Addr)
	httpSrv.Start(errChan)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		logger.Info("received shutdown signal", "signal", sig.String())
	case startErr := <-errChan:
		logger.Error("http server failed", "error", startErr)
	}

	logger.Info("shutting down http server", "timeout", shutdownTimeout.String())

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
	defer shutdownCancel()

	if stopErr := httpSrv.Stop(shutdownCtx); stopErr != nil {
		logger.Error("failed to shut down http server gracefully", "error", stopErr)
	}
}
