// Package usecase orchestrates the application logic on top of the fizzbuzz
// domain. It defines the (segregated) interfaces it needs from infrastructure
// and applies the business logic (validate, generate, record).
package usecase

// tracerName is the OpenTelemetry instrumentation scope for the use-case layer.
const tracerName = "github.com/Pimousse1099/fizz-buzz-api/usecase"
