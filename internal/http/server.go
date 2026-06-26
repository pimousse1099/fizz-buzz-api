// Package http wires the echo HTTP server: middlewares, routes and handlers for
// the fizz-buzz API.
package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/pimousse1099/fizz-buzz-api/config"
)

// New builds the echo server with its middlewares and routes wired. It takes only
// the configuration it needs: the HTTP/edge settings and the fizz-buzz `limit`
// ceiling — not the whole Config.
func New(logger *slog.Logger, validate *validator.Validate, store StatsStorer, httpCfg config.HTTP, maxLimit uint) *echo.Echo {
	e := echo.New()
	e.Logger = logger // echo v5 logs through slog natively

	useMiddlewares(e, httpCfg)

	e.GET(fizzbuzzRoute, fizzBuzzHandler(validate, store, maxLimit)) // parameters passed as query string
	e.GET(topHitsRoute, topHitsHandler(store))

	return e
}

// Run starts the server and blocks until ctx is canceled (SIGINT/SIGTERM),
// performing a graceful shutdown. It returns nil on a clean shutdown.
func Run(ctx context.Context, e *echo.Echo, httpCfg config.HTTP) error {
	sc := echo.StartConfig{Address: httpCfg.Addr, GracefulTimeout: httpCfg.ShutdownTimeout}

	err := sc.Start(ctx, e)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
