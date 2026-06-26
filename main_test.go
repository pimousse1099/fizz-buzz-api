package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

// =====================================================================================================================
// ============================================= INTEGRATION tests =====================================================
// =====================================================================================================================

func TestMainWithQueryParams(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/fizz-buzz?int1=2&int2=3&limit=10&str1=fizz&str2=buzz", nil)
	response := httptest.NewRecorder()

	getHTTPServer().ServeHTTP(response, request)

	check := assert.New(t)
	check.Equal(http.StatusOK, response.Code, response.Body.String())
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType), response.Body.String())
	check.JSONEq(`["1","fizz","buzz","fizz","5","fizzbuzz","7","fizz","buzz","fizz"]`, response.Body.String())
}

func TestMainWithMissingParams(t *testing.T) {
	// str2 is missing -> validation must fail
	request := httptest.NewRequest(http.MethodGet, "/fizz-buzz?int1=2&int2=3&limit=20&str1=fizz", nil)
	response := httptest.NewRecorder()

	getHTTPServer().ServeHTTP(response, request)

	check := assert.New(t)
	check.Equal(http.StatusBadRequest, response.Code)
	check.Equal(echo.MIMEApplicationJSON, response.Header().Get(echo.HeaderContentType))
	check.JSONEq(`"Key: 'fizzBuzzRequest.Str2' Error:Field validation for 'Str2' failed on the 'required' tag"`, response.Body.String())
}

// =====================================================================================================================
// ============================================= UNIT TESTS ============================================================
// =====================================================================================================================

func TestFizzBuzz(t *testing.T) {
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
			actualResponse := fizzBuzzController(tt.request)
			// compare fizzbuzz responses
			assert.Equal(t, tt.expectedResponse, actualResponse)
			// check fizzbuzz response size
			assert.Len(t, *actualResponse, int(tt.request.Limit))
		})
	}
}
