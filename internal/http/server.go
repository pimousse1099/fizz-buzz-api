// Package http wires the echo HTTP server: middlewares, routes and handlers for
// the fizz-buzz API.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/pimousse1099/fizz_buzz_api/internal/domain"
)

const (
	// ListenAddress is the fixed address the server binds to.
	ListenAddress = ":8080"

	// maxRequestBody caps the accepted request body size.
	maxRequestBody = 1 << 20 // 1 MiB
	// rateLimit is the per-client request rate (requests/second).
	rateLimit = 20
	// requestTimeout bounds the time spent in the handler chain.
	requestTimeout = 10 * time.Second
	// shutdownTimeout bounds how long graceful shutdown waits for in-flight requests.
	shutdownTimeout = 10 * time.Second

	routeFizzBuzz     = "/fizz-buzz"
	routeTopHitsStats = "/metrics/top-hits"
)

// StatsStorer records fizz-buzz requests and reports the most frequent one.
type StatsStorer interface {
	Record(req domain.Request)
	TopHits() (req domain.Request, hits uint, ok bool)
}

// New builds the echo server with its middlewares and routes wired.
func New(logger *slog.Logger, validate *validator.Validate, store StatsStorer) *echo.Echo {
	e := echo.New()
	e.Logger = logger // echo v5 logs through slog natively

	useMiddlewares(e, logger)

	e.GET(routeFizzBuzz, fizzBuzzHandler(validate, store)) // parameters passed as query string
	e.GET(routeTopHitsStats, topHitsHandler(store))

	return e
}

// Run starts the server and blocks until ctx is canceled (SIGINT/SIGTERM),
// performing a graceful shutdown. It returns nil on a clean shutdown.
func Run(ctx context.Context, e *echo.Echo) error {
	sc := echo.StartConfig{Address: ListenAddress, GracefulTimeout: shutdownTimeout}

	err := sc.Start(ctx, e)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
