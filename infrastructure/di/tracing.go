package di

import (
	"context"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/tracing"
)

// GetTracingShutdown configures the global tracer provider from config (service
// identity + OTLP settings) and returns its shutdown (flush) func. It is a
// no-op when tracing is disabled.
func (c *Container) GetTracingShutdown(ctx context.Context) (func(context.Context) error, error) {
	return tracing.Init(ctx, c.config.Tracing, config.AppName, config.AppVersion, c.config.Env.Type)
}
