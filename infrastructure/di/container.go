// Package di is the inversion-of-control container: it constructs and memoizes
// the application's dependencies. Getters are lazy (built on first use).
package di

import (
	"context"
	"log/slog"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// Container holds configuration and the lazily-built singletons.
type Container struct {
	ctx    context.Context //nolint:containedctx // base context for server lifecycle
	config *config.Config

	logger     *slog.Logger
	statStore  *statstorer.InMemory
	httpServer *server.Server

	generateUC *usecase.GenerateFizzBuzz
	statsUC    *usecase.GetFizzBuzzStats
}

// NewContainer builds a container from the base context and configuration.
func NewContainer(ctx context.Context, cfg *config.Config) *Container {
	return &Container{ctx: ctx, config: cfg}
}

func (c *Container) getStatStore() *statstorer.InMemory {
	if c.statStore == nil {
		c.statStore = statstorer.NewInMemory()
	}

	return c.statStore
}

func (c *Container) getGenerateUseCase() *usecase.GenerateFizzBuzz {
	if c.generateUC == nil {
		c.generateUC = usecase.NewGenerateFizzBuzz(c.config.MaxLimit, c.getStatStore())
	}

	return c.generateUC
}

func (c *Container) getStatsUseCase() *usecase.GetFizzBuzzStats {
	if c.statsUC == nil {
		c.statsUC = usecase.NewGetFizzBuzzStats(c.getStatStore())
	}

	return c.statsUC
}
