package usecase_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type spyRecorder struct {
	recorded []fizzbuzz.GenerateRequest
}

func (s *spyRecorder) Record(req fizzbuzz.GenerateRequest) {
	s.recorded = append(s.recorded, req)
}

func TestGenerateFizzBuzz_Execute_Output(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  fizzbuzz.GenerateRequest
		want []string
	}{
		{
			name: "classic fizzbuzz up to 15",
			req:  fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
			want: []string{"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz"},
		},
		{
			name: "concatenation order is str1 then str2",
			req:  fizzbuzz.GenerateRequest{Int1: 2, Int2: 3, Limit: 6, Str1: "a", Str2: "b"},
			want: []string{"1", "a", "b", "a", "5", "ab"},
		},
		{
			name: "int1 == int2 always concatenates on multiples",
			req:  fizzbuzz.GenerateRequest{Int1: 2, Int2: 2, Limit: 4, Str1: "x", Str2: "y"},
			want: []string{"1", "xy", "3", "xy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := usecase.NewGenerateFizzBuzz(1000, &spyRecorder{})

			resp, err := uc.Execute(tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(resp.Result, tt.want) {
				t.Fatalf("Result = %v, want %v", resp.Result, tt.want)
			}
		})
	}
}

func TestGenerateFizzBuzz_Execute_RecordsOnlyOnSuccess(t *testing.T) {
	t.Parallel()

	rec := &spyRecorder{}
	uc := usecase.NewGenerateFizzBuzz(1000, rec)
	req := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 5, Str1: "fizz", Str2: "buzz"}

	if _, err := uc.Execute(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rec.recorded) != 1 || rec.recorded[0] != req {
		t.Fatalf("expected request recorded once, got %v", rec.recorded)
	}
}

func TestGenerateFizzBuzz_Execute_Invalid(t *testing.T) {
	t.Parallel()

	rec := &spyRecorder{}
	uc := usecase.NewGenerateFizzBuzz(1000, rec)
	req := fizzbuzz.GenerateRequest{Int1: 0, Int2: 5, Limit: 5, Str1: "fizz", Str2: "buzz"}

	_, err := uc.Execute(req)
	if !errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if len(rec.recorded) != 0 {
		t.Fatalf("invalid request must not be recorded, got %v", rec.recorded)
	}
}
