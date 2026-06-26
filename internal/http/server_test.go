package http_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"

	httpserver "github.com/pimousse1099/fizz-buzz-api/internal/http"
	"github.com/pimousse1099/fizz-buzz-api/internal/statsstorer"
)

// testServer builds the server with a discard logger so test runs stay quiet.
func testServer() *echo.Echo {
	return httpserver.New(slog.New(slog.DiscardHandler), validator.New(), statsstorer.NewInMemory())
}

func get(t *testing.T, e *echo.Echo, target string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	return response
}

func TestFizzBuzzWithQueryParams(t *testing.T) {
	t.Parallel()

	response := get(t, testServer(), "/fizz-buzz?int1=2&int2=3&limit=10&str1=fizz&str2=buzz")

	check := assert.New(t)
	check.Equal(http.StatusOK, response.Code, response.Body.String())
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType), response.Body.String())
	check.JSONEq(`["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`, response.Body.String())
}

func TestFizzBuzzWithMissingParams(t *testing.T) {
	t.Parallel()

	// str2 is missing -> validation must fail
	response := get(t, testServer(), "/fizz-buzz?int1=2&int2=3&limit=20&str1=fizz")

	check := assert.New(t)
	check.Equal(http.StatusBadRequest, response.Code)
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType))
	check.JSONEq(`{"message":"Key: 'GenerateFizzBuzzRequest.Str2' Error:Field validation for 'Str2' failed on the 'required' tag"}`, response.Body.String())
}

func TestTopHitsStats(t *testing.T) {
	t.Parallel()

	e := testServer()

	// the "popular" call is made twice, another one once
	get(t, e, "/fizz-buzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	get(t, e, "/fizz-buzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz")
	get(t, e, "/fizz-buzz?int1=2&int2=4&limit=8&str1=x&str2=y")

	response := get(t, e, "/metrics/top-hits")

	check := assert.New(t)
	check.Equal(http.StatusOK, response.Code, response.Body.String())
	check.JSONEq(
		`{"request_params":{"str1":"fizz","str2":"buzz","int1":3,"int2":5,"limit":15},"nb_hits":2}`,
		response.Body.String(),
	)
}
