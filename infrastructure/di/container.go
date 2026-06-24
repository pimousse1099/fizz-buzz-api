// Package di is the inversion-of-control container: it constructs and memoizes
// the application's dependencies. Getters are lazy (built on first use) and live
// in concern-specific files (logger.go, stat_store.go, rate_limiter.go,
// use_case.go, http_server.go).
package di

import (
	"context"

	"github.com/go-chi/httplog/v2"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// Container holds configuration and the lazily-built singletons.
type Container struct {
	ctx    context.Context //nolint:containedctx // base context for the server lifecycle
	config *config.Config

	httpLogger  *httplog.Logger
	statStore   *statstorer.InMemory
	rateLimiter httpmiddleware.RateLimiter
	httpServer  *httpserver.Server

	generateFizzBuzzUseCase *usecase.GenerateFizzBuzz
	getFizzBuzzStatsUseCase *usecase.GetFizzBuzzStats
}

// NewContainer builds a container from the base context and configuration.
func NewContainer(ctx context.Context, cfg *config.Config) *Container {
	return &Container{ctx: ctx, config: cfg}
}
