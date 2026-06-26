package http

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
)

const (
	fizzbuzzRoute = "/fizz-buzz"
	topHitsRoute  = "/metrics/top-hits"
)

// StatsStorer records fizz-buzz requests and reports the most frequent one. It is
// defined here, on the consumer side, since the handlers are its only users.
type StatsStorer interface {
	RecordFizzBuzzRequestHit(ctx context.Context, req domain.GenerateFizzBuzzRequest) error
	GetFizzBuzzTopHits(ctx context.Context) (domain.GetFizzBuzzTopHitsResponse, error)
}

func fizzBuzzHandler(validate *validator.Validate, store StatsStorer) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// bind query parameters into the request
		req := new(domain.GenerateFizzBuzzRequest)

		err := c.Bind(req)
		if err != nil {
			// return the error so it flows through the error handler and is logged
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = validate.Struct(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		resp := domain.GenerateFizzBuzz(*req)

		ctx := c.Request().Context()

		// best-effort: a stats failure must not fail the user's request.
		// NOTE: this uses Warn, not WarnContext, so the line is NOT correlated to
		// the request (no request_id) — the base slog handler reads nothing from
		// the context, so passing it would change nothing here. Proper per-log
		// correlation would need a context-aware slog handler (see clean-archi-2026).
		err = store.RecordFizzBuzzRequestHit(ctx, *req)
		if err != nil {
			c.Logger().Warn("failed to record fizzbuzz request hit", "error", err)
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func topHitsHandler(store StatsStorer) echo.HandlerFunc {
	return func(c *echo.Context) error {
		resp, err := store.GetFizzBuzzTopHits(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}

		return c.JSON(http.StatusOK, resp)
	}
}
