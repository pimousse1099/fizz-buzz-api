package di

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

// GetHTTPServer builds the memoized HTTP server: timeouts and rate-limit from
// config, a base context tied to the container lifecycle, and an error log
// routed through slog.
func (c *Container) GetHTTPServer() *httpserver.Server {
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

		c.httpServer = httpserver.New(srv)
	}

	return c.httpServer
}

// getHTTPHandler assembles the router wrapped in the middleware stack. Order is
// outermost-first: recovery, request-id, logging, then the rate-limit guard.
func (c *Container) getHTTPHandler() http.Handler {
	return httpmiddleware.Chain(
		c.getRouter(),
		httpmiddleware.Recovery(c.GetLogger()),
		httpmiddleware.RequestID,
		httpmiddleware.Logging(c.GetLogger()),
		httpmiddleware.RateLimit(c.getRateLimiter()),
	)
}

// getRouter wires the business routes (whose patterns are declared next to their
// handlers) to the router.
func (c *Container) getRouter() http.Handler {
	return httpserver.NewRouter(
		httpserver.Route{
			Pattern: httphandler.GenerateFizzBuzzRoute,
			Handler: httphandler.GenerateFizzBuzz(c.getGenerateFizzBuzzUseCase(), c.GetLogger()),
		},
		httpserver.Route{
			Pattern: httphandler.GetFizzBuzzStatsRoute,
			Handler: httphandler.GetFizzBuzzStats(c.getGetFizzBuzzStatsUseCase(), c.GetLogger()),
		},
	)
}
