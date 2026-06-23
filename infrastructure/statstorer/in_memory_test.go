package statstorer_test

import (
	"context"
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

	_, _, ok := statstorer.NewInMemory().MostFrequent(context.Background())
	if ok {
		t.Fatal("expected ok=false on empty store")
	}
}

func TestInMemory_MostFrequent(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	s.Record(context.Background(), a)
	s.Record(context.Background(), a)
	s.Record(context.Background(), a)
	s.Record(context.Background(), b)

	got, hits, ok := s.MostFrequent(context.Background())
	if !ok || got != a || hits != 3 {
		t.Fatalf("got %+v hits=%d ok=%v, want %+v hits=3", got, hits, ok, a)
	}
}

func TestInMemory_TieBreakFirstToReachMax(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	s.Record(context.Background(), a) // a=1
	s.Record(context.Background(), a) // a=2  -> top a
	s.Record(context.Background(), b) // b=1
	s.Record(context.Background(), b) // b=2  -> not strictly greater, top stays a

	got, hits, _ := s.MostFrequent(context.Background())
	if got != a || hits != 2 {
		t.Fatalf("tie must keep first to reach max: got %+v hits=%d, want %+v hits=2", got, hits, a)
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
				s.Record(context.Background(), a)
			}
		}()
	}

	wg.Wait()

	_, hits, ok := s.MostFrequent(context.Background())
	if !ok || hits != goroutines*perGoroutine {
		t.Fatalf("got hits=%d ok=%v, want %d", hits, ok, goroutines*perGoroutine)
	}
}
