package di

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	ctxlog "github.com/go-chi/httplog/v2"
	"github.com/go-chi/httprate"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

// GetHTTPServer builds the memoized HTTP server: timeouts from config, a base
// context tied to the container lifecycle, and an error log routed through slog.
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

// getHTTPHandler builds the chi router and its middleware stack. Order (outer to
// inner): request-id, structured request logging, panic recovery, rate-limit.
// The rate limiter is per-IP (local guard); a distributed counter is available
// via go-chi/httprate-redis (httprate.WithLimitCounter) — see the ADR.
func (c *Container) getHTTPHandler() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(ctxlog.RequestLogger(c.getHTTPLogger()))
	router.Use(middleware.Recoverer)
	router.Use(httprate.LimitByIP(c.config.HTTP.RateLimitRequests, c.config.HTTP.RateLimitWindow))

	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	router.Get(httphandler.GenerateFizzBuzzRoute, httphandler.GenerateFizzBuzz(c.getGenerateFizzBuzzUseCase()))
	router.Get(httphandler.GetFizzBuzzStatsRoute, httphandler.GetFizzBuzzStats(c.getGetFizzBuzzStatsUseCase()))

	return router
}
