package di

import (
	"log/slog"
	"os"

	ctxlog "github.com/go-chi/httplog/v2"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// GetLogger returns the underlying structured logger for non-request logging
// (startup/shutdown, the HTTP server error log).
func (c *Container) GetLogger() *slog.Logger {
	return c.getHTTPLogger().Logger
}

// getHTTPLogger returns the memoized httplog logger (slog-backed), tagged with
// the base context fields. It also drives the request logging middleware.
func (c *Container) getHTTPLogger() *ctxlog.Logger {
	if c.httpLogger == nil {
		tags := map[string]string{
			"application_name":    config.AppName,
			"application_version": config.AppVersion,
			"environment_type":    c.config.Env.Type,
		}

		if c.config.Env.Name != "" {
			tags["environment_name"] = c.config.Env.Name
		}

		hostname, err := os.Hostname()
		if err == nil {
			tags["host_name"] = hostname
		}

		c.httpLogger = ctxlog.NewLogger(config.AppName, ctxlog.Options{
			JSON:     true,
			LogLevel: c.config.Observability.LogLevel,
			Tags:     tags,
		})
	}

	return c.httpLogger
}
