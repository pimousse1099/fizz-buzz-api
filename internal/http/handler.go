package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
)

// StatsStorer records fizz-buzz requests and reports the most frequent one. It is
// defined here, on the consumer side, since the handlers are its only users.
type StatsStorer interface {
	RecordFizzBuzzRequestHit(req domain.GenerateFizzBuzzRequest)
	GetFizzBuzzTopHits() (req domain.GenerateFizzBuzzRequest, hits uint, ok bool)
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

		store.RecordFizzBuzzRequestHit(*req)

		return c.JSON(http.StatusOK, domain.Generate(*req))
	}
}

func topHitsHandler(store StatsStorer) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req, hits, ok := store.GetFizzBuzzTopHits()
		if !ok {
			return c.JSON(http.StatusOK, "no data collected yet")
		}

		return c.JSON(http.StatusOK, domain.GetFizzBuzzTopHitsResponse{RequestParams: req, Hits: hits})
	}
}
