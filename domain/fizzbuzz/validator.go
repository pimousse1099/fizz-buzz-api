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
		return fmt.Errorf("int1 must be a positive integer, got %d: %w", r.Int1, ErrFailedToValidateGenerateRequest)
	case r.Int2 <= 0:
		return fmt.Errorf("int2 must be a positive integer, got %d: %w", r.Int2, ErrFailedToValidateGenerateRequest)
	case r.Limit < 1 || r.Limit > maxLimit:
		return fmt.Errorf("limit must be between 1 and %d, got %d: %w", maxLimit, r.Limit, ErrFailedToValidateGenerateRequest)
	case r.Str1 == "":
		return fmt.Errorf("str1 must not be empty: %w", ErrFailedToValidateGenerateRequest)
	case r.Str2 == "":
		return fmt.Errorf("str2 must not be empty: %w", ErrFailedToValidateGenerateRequest)
	case len(r.Str1) > maxStrLen:
		return fmt.Errorf("str1 must be at most %d characters: %w", maxStrLen, ErrFailedToValidateGenerateRequest)
	case len(r.Str2) > maxStrLen:
		return fmt.Errorf("str2 must be at most %d characters: %w", maxStrLen, ErrFailedToValidateGenerateRequest)
	default:
		return nil
	}
}
