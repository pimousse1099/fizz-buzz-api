// Package usecase orchestrates the application logic on top of the fizzbuzz
// domain. It defines the (segregated) interfaces it needs from infrastructure
// and applies the business logic (validate, generate, record).
package usecase

import (
	"context"
	"log/slog"
	"strconv"

	ctxlog "github.com/go-chi/httplog/v2"
	"go.opentelemetry.io/otel"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// tracerName is the OpenTelemetry instrumentation scope for the use-case layer.
const tracerName = "github.com/Pimousse1099/fizz-buzz-api/usecase"

// StatRecorder records a successful generation request for statistics.
type StatRecorder interface {
	RecordFizzBuzzStat(ctx context.Context, req fizzbuzz.GenerateRequest) error
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
func (uc *GenerateFizzBuzz) Execute(ctx context.Context, req fizzbuzz.GenerateRequest) (*fizzbuzz.GenerateResponse, error) {
	ctx, span := otel.Tracer(tracerName).Start(ctx, "usecase.generate_fizzbuzz")
	defer span.End()

	ctxlog.LogEntrySetField(ctx, "use_case", slog.StringValue("generate_fizzbuzz"))

	err := req.Validate(uc.maxLimit)
	if err != nil {
		span.RecordError(err)

		return nil, err // returns an ErrFailedToValidateGenerateRequest error.
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

	err = uc.recorder.RecordFizzBuzzStat(ctx, req)
	if err != nil {
		// Stats recording is best-effort: a failure must not fail an otherwise
		// successful generation, so we log a warning (with the request-scoped
		// fields) and still return the result.
		ctxlog.LogEntry(ctx).Warn("failed to record fizzbuzz stat", "error", err)
	}

	return &fizzbuzz.GenerateResponse{Result: result}, nil
}
