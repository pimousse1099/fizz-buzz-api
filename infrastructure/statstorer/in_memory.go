// Package statstorer holds implementations of the use-case stat interfaces
// (StatRecorder, StatReader): an in-memory store and a Redis placeholder.
package statstorer

import (
	"context"
	"sync"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// InMemory is a concurrency-safe, process-local stat counter. The current most
// frequent request is memoized: it is updated only when a count becomes
// strictly greater than the current maximum, so on ties the first request to
// reach the maximum is kept. State is lost on restart.
type InMemory struct {
	mu      sync.Mutex
	counts  map[fizzbuzz.GenerateRequest]int
	topReq  fizzbuzz.GenerateRequest
	topHits int
}

// NewInMemory builds an empty in-memory stat store.
func NewInMemory() *InMemory {
	return &InMemory{counts: make(map[fizzbuzz.GenerateRequest]int)}
}

// RecordFizzBuzzStat increments the counter for req.
func (s *InMemory) RecordFizzBuzzStat(_ context.Context, req fizzbuzz.GenerateRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[req]++

	if s.counts[req] > s.topHits {
		s.topHits = s.counts[req]
		s.topReq = req
	}
}

// GetMostFrequentFizzbuzzRequest returns the most frequent request and its hit
// count, or fizzbuzz.ErrNoStatsRecorded if nothing has been recorded yet.
func (s *InMemory) GetMostFrequentFizzbuzzRequest(_ context.Context) (*fizzbuzz.GetStatsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.topHits == 0 {
		return nil, fizzbuzz.ErrNoStatsRecorded
	}

	return &fizzbuzz.GetStatsResponse{Request: s.topReq, TotalHits: s.topHits}, nil
}
