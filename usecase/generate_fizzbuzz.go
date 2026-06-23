// Package usecase orchestrates the application logic on top of the fizzbuzz
// domain. It defines the (segregated) interfaces it needs from infrastructure
// and applies the business logic (validate, generate, record).
package usecase

import (
	"strconv"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// StatRecorder records a successful generation request for statistics.
type StatRecorder interface {
	Record(req fizzbuzz.GenerateRequest)
}

// GenerateFizzBuzz validates a request, generates the sequence, and records the
// request only on success.
type GenerateFizzBuzz struct {
	maxLimit int
	recorder StatRecorder
}

// NewGenerateFizzBuzz builds the use-case with its max-limit bound and recorder.
func NewGenerateFizzBuzz(maxLimit int, recorder StatRecorder) *GenerateFizzBuzz {
	return &GenerateFizzBuzz{maxLimit: maxLimit, recorder: recorder}
}

// Execute validates the request, generates the fizz-buzz sequence, and records
// the request. The generation is the application's business logic and lives
// here rather than on the domain type.
func (uc *GenerateFizzBuzz) Execute(req fizzbuzz.GenerateRequest) (fizzbuzz.GenerateResponse, error) {
	if err := req.Validate(uc.maxLimit); err != nil {
		return fizzbuzz.GenerateResponse{}, err
	}

	result := make([]string, 0, req.Limit)

	for n := 1; n <= req.Limit; n++ {
		switch {
		case n%req.Int1 == 0 && n%req.Int2 == 0:
			result = append(result, req.Str1+req.Str2)
		case n%req.Int1 == 0:
			result = append(result, req.Str1)
		case n%req.Int2 == 0:
			result = append(result, req.Str2)
		default:
			result = append(result, strconv.Itoa(n))
		}
	}

	uc.recorder.Record(req)

	return fizzbuzz.GenerateResponse{Result: result}, nil
}
