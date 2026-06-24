package di

import (
	"log/slog"
	"os"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// GetLogger returns the memoized structured JSON logger, tagged with the base
// context fields (application, version, environment, host).
func (c *Container) GetLogger() *slog.Logger {
	if c.logger == nil {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: c.config.Observability.LogLevel,
		})

		logger := slog.New(handler).With(
			"application_name", config.AppName,
			"application_version", config.AppVersion,
			"environment_type", c.config.Env.Type,
		)

		if c.config.Env.Name != "" {
			logger = logger.With("environment_name", c.config.Env.Name)
		}

		if hostname, err := os.Hostname(); err == nil {
			logger = logger.With("host_name", hostname)
		} else {
			logger.Warn("failed to resolve hostname", "error", err)
		}

		c.logger = logger
	}

	return c.logger
}
