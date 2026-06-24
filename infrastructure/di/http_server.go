package di

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

// GetHTTPServer builds the memoized HTTP server: timeouts and rate-limit from
// config, a base context tied to the container lifecycle, and an error log
// routed through slog.
func (c *Container) GetHTTPServer() *server.Server {
	if c.httpServer == nil {
		srv := &http.Server{
			Addr:              c.config.HTTP.Addr,
			Handler:           c.getHTTPHandler(),
			ReadHeaderTimeout: c.config.HTTP.ReadHeaderTimeout,
			WriteTimeout:      c.config.HTTP.WriteTimeout,
			IdleTimeout:       c.config.HTTP.IdleTimeout,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
			BaseContext:       func(net.Listener) context.Context { return c.ctx },
			ErrorLog:          slog.NewLogLogger(c.GetLogger().Handler(), slog.LevelError),
		}

		c.httpServer = server.New(srv)
	}

	return c.httpServer
}

// getHTTPHandler assembles the router wrapped in the middleware stack. Order is
// outermost-first: recovery, request-id, logging, then the rate-limit guard.
func (c *Container) getHTTPHandler() http.Handler {
	return server.Chain(
		c.getRouter(),
		server.Recovery(c.GetLogger()),
		server.RequestID,
		server.Logging(c.GetLogger()),
		server.RateLimit(c.getRateLimiter()),
	)
}

// getRouter builds the route table wired to the handlers.
func (c *Container) getRouter() http.Handler {
	return server.NewRouter(
		handler.GenerateFizzBuzz(c.getGenerateFizzBuzzUseCase(), c.GetLogger()),
		handler.GetFizzBuzzStats(c.getFizzBuzzStatsUseCase(), c.GetLogger()),
	)
}
