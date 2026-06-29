// Package domain holds the fizz-buzz request/response types and the core
// generation and validation logic, free of any transport or storage concern.
package domain

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrInvalidRequest is wrapped by every GenerateFizzBuzzRequest validation
// failure. Callers classify validation errors with errors.Is.
var ErrInvalidRequest = errors.New("failed to validate fizz-buzz request")

type (
	// GenerateFizzBuzzRequest is the set of parameters of a fizz-buzz call. The struct tags drive
	// query-string binding (query) and JSON encoding (json); validation lives in the Validate method.
	GenerateFizzBuzzRequest struct {
		Str1  string `json:"str1"  query:"str1"`
		Str2  string `json:"str2"  query:"str2"`
		Int1  uint   `json:"int1"  query:"int1"`
		Int2  uint   `json:"int2"  query:"int2"`
		Limit uint   `json:"limit" query:"limit"`
	}

	// GenerateFizzBuzzResponse is the produced list of values, from 1 to Limit.
	GenerateFizzBuzzResponse []string

	// GetFizzBuzzTopHitsResponse is the payload of the statistics endpoint: the
	// parameters of the most frequently requested fizz-buzz call and how many
	// times it was made.
	GetFizzBuzzTopHitsResponse struct {
		RequestParams GenerateFizzBuzzRequest `json:"request_params"`
		Hits          uint                    `json:"nb_hits"`
	}
)

// Validate checks the invariants of a fizz-buzz request: positive int1/int2,
// a limit in [1, maxLimit], and non-empty str1/str2. Every failure wraps
// ErrInvalidRequest (sentinel first, then the specific detail). maxLimit is
// supplied by the caller (configuration) rather than hard-coded.
//
// A value receiver is used deliberately: the request is a small value object
// (also used as a map key), so copying is cheap.
func (r GenerateFizzBuzzRequest) Validate(maxLimit uint) error {
	switch {
	case r.Int1 == 0:
		return fmt.Errorf("%w: int1 must be a positive integer", ErrInvalidRequest)
	case r.Int2 == 0:
		return fmt.Errorf("%w: int2 must be a positive integer", ErrInvalidRequest)
	case r.Limit == 0 || r.Limit > maxLimit:
		return fmt.Errorf("%w: limit must be between 1 and %d, got %d", ErrInvalidRequest, maxLimit, r.Limit)
	case r.Str1 == "":
		return fmt.Errorf("%w: str1 must not be empty", ErrInvalidRequest)
	case r.Str2 == "":
		return fmt.Errorf("%w: str2 must not be empty", ErrInvalidRequest)
	default:
		return nil
	}
}

// GenerateFizzBuzz builds the fizz-buzz response for req: multiples of Int1 become
// Str1, multiples of Int2 become Str2, multiples of both become Str1+Str2, and
// every other number is itself. Int1 and Int2 are assumed non-zero (enforced by
// validation before this is called).
func GenerateFizzBuzz(req GenerateFizzBuzzRequest) GenerateFizzBuzzResponse {
	resp := make(GenerateFizzBuzzResponse, req.Limit)

	for i := uint(1); i <= req.Limit; i++ {
		res := ""

		if i%req.Int1 == 0 {
			res += req.Str1
		}

		if i%req.Int2 == 0 {
			res += req.Str2
		}

		if res != "" {
			resp[i-1] = res

			continue
		}

		resp[i-1] = strconv.FormatUint(uint64(i), 10)
	}

	return resp
}
