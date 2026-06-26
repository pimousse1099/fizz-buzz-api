package statsstorer_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
	"github.com/pimousse1099/fizz-buzz-api/internal/statsstorer"
)

func TestInMemoryEmpty(t *testing.T) {
	t.Parallel()

	_, _, ok := statsstorer.NewInMemory().TopHits()
	assert.False(t, ok)
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

			store.Record(popular)
		}()

		go func() {
			defer wg.Done()

			store.Record(other)
			store.Record(popular) // popular gets twice the hits of other
		}()
	}

	wg.Wait()

	req, hits, ok := store.TopHits()
	check := assert.New(t)
	check.True(ok)
	check.Equal(popular, req)
	check.Equal(uint(200), hits)
}
