package main

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// testServer builds the HTTP server with its dependencies, using a logger that
// discards output so test runs stay quiet.
func testServer() *echo.Echo {
	return getHTTPServer(slog.New(slog.DiscardHandler), validator.New(), newMetricsCollector())
}

// =====================================================================================================================
// ============================================= INTEGRATION tests =====================================================
// =====================================================================================================================

func TestMainWithQueryParams(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/fizz-buzz?int1=2&int2=3&limit=10&str1=fizz&str2=buzz", http.NoBody)
	response := httptest.NewRecorder()

	testServer().ServeHTTP(response, request)

	check := assert.New(t)
	check.Equal(http.StatusOK, response.Code, response.Body.String())
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType), response.Body.String())
	check.JSONEq(`["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`, response.Body.String())
}

func TestMainWithMissingParams(t *testing.T) {
	t.Parallel()

	// str2 is missing -> validation must fail
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/fizz-buzz?int1=2&int2=3&limit=20&str1=fizz", http.NoBody)
	response := httptest.NewRecorder()

	testServer().ServeHTTP(response, request)

	check := assert.New(t)
	check.Equal(http.StatusBadRequest, response.Code)
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType))
	check.JSONEq(`{"message":"Key: 'fizzBuzzRequest.Str2' Error:Field validation for 'Str2' failed on the 'required' tag"}`, response.Body.String())
}

func TestMetricsCollectorConcurrent(t *testing.T) {
	t.Parallel()

	mc := newMetricsCollector()
	popular := fizzBuzzRequest{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
	other := fizzBuzzRequest{Int1: 2, Int2: 7, Limit: 10, Str1: "a", Str2: "b"}

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			mc.record(popular)
		}()

		go func() {
			defer wg.Done()

			mc.record(other)
			mc.record(popular) // popular gets twice the hits of other
		}()
	}

	wg.Wait()

	req, hits, ok := mc.top()
	check := assert.New(t)
	check.True(ok)
	check.Equal(popular, req)
	check.Equal(uint(200), hits)
}

// =====================================================================================================================
// ============================================= UNIT TESTS ============================================================
// =====================================================================================================================

func TestFizzBuzz(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		request          *fizzBuzzRequest
		expectedResponse *fizzBuzzResponse
	}{
		{
			name:             "test without data",
			request:          &fizzBuzzRequest{},
			expectedResponse: &fizzBuzzResponse{},
		},
		{
			name: "test with nominal data",
			request: &fizzBuzzRequest{
				Int1:  3,
				Int2:  5,
				Limit: 20,
				Str1:  "fizz",
				Str2:  "buzz",
			},
			expectedResponse: &fizzBuzzResponse{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz", "16", "17", "fizz", "19", "buzz",
			},
		},
		{
			name: "test with limit 0",
			request: &fizzBuzzRequest{
				Int1:  3,
				Int2:  5,
				Limit: 0,
				Str1:  "fizz",
				Str2:  "buzz",
			},
			expectedResponse: &fizzBuzzResponse{},
		},
		{
			name: "test with inverted int1 int2",
			request: &fizzBuzzRequest{
				Int1:  5,
				Int2:  3,
				Limit: 20,
				Str1:  "fizz",
				Str2:  "buzz",
			},
			expectedResponse: &fizzBuzzResponse{
				"1", "2", "buzz", "4", "fizz", "buzz", "7", "8", "buzz", "fizz", "11", "buzz", "13", "14", "fizzbuzz", "16", "17", "buzz", "19", "fizz",
			},
		},
		{
			name: "test with spaces",
			request: &fizzBuzzRequest{
				Int1:  3,
				Int2:  5,
				Limit: 20,
				Str1:  "fizz ",
				Str2:  " buzz",
			},
			expectedResponse: &fizzBuzzResponse{
				"1", "2", "fizz ", "4", " buzz", "fizz ", "7", "8", "fizz ", " buzz", "11", "fizz ", "13", "14", "fizz  buzz", "16", "17", "fizz ", "19", " buzz",
			},
		},
		{
			name: "test with special chars",
			request: &fizzBuzzRequest{
				Int1:  3,
				Int2:  5,
				Limit: 20,
				Str1:  "fizz&é\"'è§",
				Str2:  " #(~?-_`buzz",
			},
			expectedResponse: &fizzBuzzResponse{
				"1", "2", "fizz&é\"'è§", "4", " #(~?-_`buzz", "fizz&é\"'è§", "7", "8", "fizz&é\"'è§", " #(~?-_`buzz", "11", "fizz&é\"'è§", "13", "14", "fizz&é\"'è§ #(~?-_`buzz", "16", "17", "fizz&é\"'è§", "19", " #(~?-_`buzz",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actualResponse := fizzBuzzController(tt.request)
			// compare fizzbuzz responses
			assert.Equal(t, tt.expectedResponse, actualResponse)
			// check fizzbuzz response size
			assert.Len(t, *actualResponse, int(tt.request.Limit))
		})
	}
}
