package fizzbuzz_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

func TestGenerateRequest_Validate(t *testing.T) {
	t.Parallel()

	const maxLimit = 1000

	valid := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}

	tests := []struct {
		name    string
		mutate  func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest
		wantErr bool
	}{
		{"valid", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { return r }, false},
		{"int1 zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int1 = 0; return r }, true},
		{"int1 negative", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int1 = -1; return r }, true},
		{"int2 zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int2 = 0; return r }, true},
		{"limit zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = 0; return r }, true},
		{"limit above max", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = maxLimit + 1; return r }, true},
		{"limit at max", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = maxLimit; return r }, false},
		{"str1 empty", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str1 = ""; return r }, true},
		{"str2 empty", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str2 = ""; return r }, true},
		{"str1 too long", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str1 = strings.Repeat("a", 101); return r }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mutate(valid).Validate(maxLimit)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
					t.Fatalf("error %v does not wrap ErrFailedToValidateGenerateRequest", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
