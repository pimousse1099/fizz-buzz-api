package di

import (
	"log/slog"
	"os"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// GetLogger returns the memoized structured JSON logger.
func (c *Container) GetLogger() *slog.Logger {
	if c.logger == nil {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(c.config.LogLevel)})
		c.logger = slog.New(handler).With(
			"app", config.AppName,
			"version", config.AppVersion,
		)
	}

	return c.logger
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
