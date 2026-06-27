package statstorer_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
)

func req(int1 int) fizzbuzz.GenerateRequest {
	return fizzbuzz.GenerateRequest{Int1: int1, Int2: 5, Limit: 10, Str1: "fizz", Str2: "buzz"}
}

func TestInMemory_Empty(t *testing.T) {
	t.Parallel()

	_, err := statstorer.NewInMemory().GetMostFrequentFizzbuzzRequest(context.Background())
	if !errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
		t.Fatalf("expected ErrNoStatsRecorded on empty store, got %v", err)
	}
}

func TestInMemory_MostFrequent(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	_ = s.RecordFizzBuzzStat(context.Background(), a)
	_ = s.RecordFizzBuzzStat(context.Background(), a)
	_ = s.RecordFizzBuzzStat(context.Background(), a)
	_ = s.RecordFizzBuzzStat(context.Background(), b)

	resp, err := s.GetMostFrequentFizzbuzzRequest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Request != a || resp.TotalHits != 3 {
		t.Fatalf("got %+v, want request %+v total hits 3", resp, a)
	}
}

func TestInMemory_TieBreakFirstToReachMax(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	_ = s.RecordFizzBuzzStat(context.Background(), a) // a=1
	_ = s.RecordFizzBuzzStat(context.Background(), a) // a=2  -> top a
	_ = s.RecordFizzBuzzStat(context.Background(), b) // b=1
	_ = s.RecordFizzBuzzStat(context.Background(), b) // b=2  -> not strictly greater, top stays a

	resp, err := s.GetMostFrequentFizzbuzzRequest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Request != a || resp.TotalHits != 2 {
		t.Fatalf("tie must keep first to reach max: got %+v, want request %+v total hits 2", resp, a)
	}
}

func TestInMemory_ConcurrentRecord(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a := req(3)

	const goroutines, perGoroutine = 50, 100

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			for range perGoroutine {
				_ = s.RecordFizzBuzzStat(context.Background(), a)
			}
		}()
	}

	wg.Wait()

	resp, err := s.GetMostFrequentFizzbuzzRequest(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TotalHits != goroutines*perGoroutine {
		t.Fatalf("got total hits=%d, want %d", resp.TotalHits, goroutines*perGoroutine)
	}
}
