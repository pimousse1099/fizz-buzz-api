package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type stubReader struct {
	resp *fizzbuzz.GetStatsResponse
	err  error
}

func (s stubReader) GetMostFrequentFizzbuzzRequest(_ context.Context) (*fizzbuzz.GetStatsResponse, error) {
	return s.resp, s.err
}

func TestGetFizzBuzzStats_Execute_WithData(t *testing.T) {
	t.Parallel()

	want := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}
	uc := usecase.NewGetFizzBuzzStats(stubReader{
		resp: &fizzbuzz.GetStatsResponse{Request: want, TotalHits: 7},
	})

	resp, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Request != want || resp.TotalHits != 7 {
		t.Fatalf("got %+v, want request %+v total hits 7", resp, want)
	}
}

func TestGetFizzBuzzStats_Execute_Empty(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetFizzBuzzStats(stubReader{err: fizzbuzz.ErrNoStatsRecorded})

	_, err := uc.Execute(context.Background())
	if !errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
		t.Fatalf("expected ErrNoStatsRecorded, got %v", err)
	}
}
