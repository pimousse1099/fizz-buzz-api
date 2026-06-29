package statsstorer_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
	"github.com/pimousse1099/fizz-buzz-api/internal/statsstorer"
)

func TestInMemoryEmpty(t *testing.T) {
	t.Parallel()

	resp, err := statsstorer.NewInMemory().GetFizzBuzzTopHits(context.Background())

	check := assert.New(t)
	check.NoError(err)
	check.Zero(resp.Hits)
}

func TestInMemoryConcurrent(t *testing.T) {
	t.Parallel()

	store := statsstorer.NewInMemory()
	popular := domain.GenerateFizzBuzzRequest{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"}
	other := domain.GenerateFizzBuzzRequest{Int1: 2, Int2: 7, Limit: 10, Str1: "a", Str2: "b"}

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			_ = store.RecordFizzBuzzRequestHit(context.Background(), popular)
		}()

		go func() {
			defer wg.Done()

			_ = store.RecordFizzBuzzRequestHit(context.Background(), other)
			_ = store.RecordFizzBuzzRequestHit(context.Background(), popular) // popular gets twice the hits of other
		}()
	}

	wg.Wait()

	resp, err := store.GetFizzBuzzTopHits(context.Background())
	check := assert.New(t)
	check.NoError(err)
	check.Equal(popular, resp.RequestParams)
	check.Equal(uint(200), resp.Hits)
}
