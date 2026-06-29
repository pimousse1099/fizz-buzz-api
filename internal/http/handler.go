package http

import (
	"context"
	"errors"
	"net/http"

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

func fizzBuzzHandler(store StatsStorer, maxLimit uint) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// bind query parameters into the request
		req := new(domain.GenerateFizzBuzzRequest)

		err := c.Bind(req)
		if err != nil {
			// echo wraps the real cause inside its *HTTPError; unwrap it so the
			// client gets "failed to bind request: <cause>" rather than echo's
			// re-stringified "code=400, message=..." envelope.
			cause := errors.Unwrap(err)
			if cause == nil {
				cause = err
			}

			return echo.NewHTTPError(http.StatusBadRequest, "failed to bind request: "+cause.Error())
		}

		// the request validates itself; maxLimit bounds the sequence size so a
		// huge limit can't allocate a huge slice (DoS). The bound is inclusive.
		err = req.Validate(maxLimit)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		resp := domain.GenerateFizzBuzz(*req)

		ctx := c.Request().Context()

		// best-effort: a stats failure must not fail the user's request.
		// WarnContext passes ctx so slogctx stamps the request_id onto the line.
		err = store.RecordFizzBuzzRequestHit(ctx, *req)
		if err != nil {
			c.Logger().WarnContext(ctx, "failed to record fizzbuzz request hit", "error", err)
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
