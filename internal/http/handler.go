package http

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"

	"github.com/pimousse1099/fizz_buzz_api/internal/domain"
)

// statsResponse is the payload of the statistics endpoint: the parameters of the
// most frequently requested fizz-buzz call and how many times it was made.
type statsResponse struct {
	RequestParams domain.Request `json:"request_params"`
	Hits          uint           `json:"nb_hits"`
}

func fizzBuzzHandler(validate *validator.Validate, store StatsStorer) echo.HandlerFunc {
	return func(c *echo.Context) error {
		// bind query parameters into the request
		req := new(domain.Request)

		err := c.Bind(req)
		if err != nil {
			// return the error so it flows through the error handler and is logged
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = validate.Struct(req)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		store.Record(*req)

		return c.JSON(http.StatusOK, domain.Generate(*req))
	}
}

func topHitsHandler(store StatsStorer) echo.HandlerFunc {
	return func(c *echo.Context) error {
		req, hits, ok := store.TopHits()
		if !ok {
			return c.JSON(http.StatusOK, "no data collected yet")
		}

		return c.JSON(http.StatusOK, statsResponse{RequestParams: req, Hits: hits})
	}
}
