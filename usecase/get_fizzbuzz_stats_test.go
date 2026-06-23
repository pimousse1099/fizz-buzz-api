package usecase_test

import (
	"errors"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type stubReader struct {
	req  fizzbuzz.GenerateRequest
	hits int
	ok   bool
}

func (s stubReader) MostFrequent() (fizzbuzz.GenerateRequest, int, bool) {
	return s.req, s.hits, s.ok
}

func TestGetFizzBuzzStats_Execute_WithData(t *testing.T) {
	t.Parallel()

	want := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}
	uc := usecase.NewGetFizzBuzzStats(stubReader{req: want, hits: 7, ok: true})

	resp, err := uc.Execute(fizzbuzz.GetStatsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Request != want || resp.Hits != 7 {
		t.Fatalf("got %+v, want request %+v hits 7", resp, want)
	}
}

func TestGetFizzBuzzStats_Execute_Empty(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetFizzBuzzStats(stubReader{ok: false})

	_, err := uc.Execute(fizzbuzz.GetStatsRequest{})
	if !errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
		t.Fatalf("expected ErrNoStatsRecorded, got %v", err)
	}
}
