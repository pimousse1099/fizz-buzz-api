package fizzbuzz

import "fmt"

const maxStrLen = 100

// Validate checks the business invariants of a GenerateRequest. Every failure
// wraps ErrFailedToValidateGenerateRequest. maxLimit is supplied by the caller
// (configuration) rather than hard-coded.
//
// A value receiver is used deliberately: GenerateRequest is a small, immutable
// value object (also used as a map key), so copying is cheap and avoids a heap
// escape; see the ADR for the full rationale.
func (r GenerateRequest) Validate(maxLimit int) error {
	switch {
	case r.Int1 <= 0:
		return fmt.Errorf("%w: int1 must be a positive integer, got %d", ErrFailedToValidateGenerateRequest, r.Int1)
	case r.Int2 <= 0:
		return fmt.Errorf("%w: int2 must be a positive integer, got %d", ErrFailedToValidateGenerateRequest, r.Int2)
	case r.Limit < 1 || r.Limit > maxLimit:
		return fmt.Errorf("%w: limit must be between 1 and %d, got %d", ErrFailedToValidateGenerateRequest, maxLimit, r.Limit)
	case r.Str1 == "":
		return fmt.Errorf("%w: str1 must not be empty", ErrFailedToValidateGenerateRequest)
	case r.Str2 == "":
		return fmt.Errorf("%w: str2 must not be empty", ErrFailedToValidateGenerateRequest)
	case len(r.Str1) > maxStrLen:
		return fmt.Errorf("%w: str1 must be at most %d characters", ErrFailedToValidateGenerateRequest, maxStrLen)
	case len(r.Str2) > maxStrLen:
		return fmt.Errorf("%w: str2 must be at most %d characters", ErrFailedToValidateGenerateRequest, maxStrLen)
	default:
		return nil
	}
}
