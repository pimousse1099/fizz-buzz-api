// Package domain holds the fizz-buzz request/response types and the core
// generation logic, free of any transport or storage concern.
package domain

import "strconv"

type (
	// Request is the set of parameters of a fizz-buzz call. The struct tags drive
	// query-string binding (query), JSON encoding (json) and validation.
	Request struct {
		Str1  string `json:"str1"  query:"str1"  validate:"required"`
		Str2  string `json:"str2"  query:"str2"  validate:"required"`
		Int1  uint   `json:"int1"  query:"int1"  validate:"required"`
		Int2  uint   `json:"int2"  query:"int2"  validate:"required"`
		Limit uint   `json:"limit" query:"limit" validate:"required"`
	}

	// Response is the produced list of values, from 1 to Limit.
	Response []string
)

// Generate builds the fizz-buzz response for req: multiples of Int1 become Str1,
// multiples of Int2 become Str2, multiples of both become Str1+Str2, and every
// other number is itself. Int1 and Int2 are assumed non-zero (enforced by
// validation before this is called).
func Generate(req Request) Response {
	resp := make(Response, req.Limit)

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
