// Command fizz-buzz-api starts the HTTP server.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/di"
)

const shutdownTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New(ctx)
	if err != nil {
		log.Println("failed to load config:", err)

		return
	}

	container := di.NewContainer(ctx, cfg)
	logger := container.GetLogger()

	httpSrv := container.GetHTTPServer()
	errChan := make(chan error, 1)
	httpSrv.Start(errChan)

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case startErr := <-errChan:
		logger.Error("server failed", "error", startErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if stopErr := httpSrv.Stop(shutdownCtx); stopErr != nil {
		logger.Error("graceful shutdown failed", "error", stopErr)
	}
}
