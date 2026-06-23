package di

import (
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/ratelimiter"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

// HTTPHandler builds the fully-wired HTTP handler (router + middleware stack).
// Exposed so tests can exercise it via httptest without binding a port.
func (c *Container) HTTPHandler() http.Handler {
	logger := c.GetLogger()

	router := server.NewRouter(
		handler.GenerateFizzBuzz(c.getGenerateUseCase(), logger),
		handler.GetFizzBuzzStats(c.getStatsUseCase(), logger),
	)

	limiter := ratelimiter.NewInMemory(c.config.RateLimitPerSec, c.config.RateLimitBurst)

	return server.Chain(
		router,
		server.Recovery(logger),
		server.RequestID,
		server.Logging(logger),
		server.RateLimit(limiter),
	)
}

// GetHTTPServer builds the memoized HTTP server with timeouts from config.
func (c *Container) GetHTTPServer() *server.Server {
	if c.httpServer == nil {
		srv := &http.Server{
			Addr:              c.config.HTTPAddr,
			Handler:           c.HTTPHandler(),
			ReadHeaderTimeout: c.config.ReadHeaderTimeout,
			WriteTimeout:      c.config.WriteTimeout,
			IdleTimeout:       c.config.IdleTimeout,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		}

		c.httpServer = server.New(srv, c.GetLogger())
	}

	return c.httpServer
}
