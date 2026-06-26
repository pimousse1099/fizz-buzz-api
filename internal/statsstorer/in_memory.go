// Package statsstorer provides storage for fizz-buzz request statistics.
package statsstorer

import (
	"context"
	"sync"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
)

// InMemory counts how many times each distinct request has been served and keeps
// track of the most frequent one as requests come in. It is safe for concurrent
// use: every access goes through the mutex. The request struct is used directly
// as the map key (all of its fields are comparable).
type InMemory struct {
	mu      sync.Mutex
	counts  map[domain.GenerateFizzBuzzRequest]uint
	topReq  domain.GenerateFizzBuzzRequest
	topHits uint
}

// NewInMemory returns a ready-to-use in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{counts: make(map[domain.GenerateFizzBuzzRequest]uint)}
}

// RecordFizzBuzzRequestHit increments the counter for req and updates the running
// most-frequent request. Both this and GetFizzBuzzTopHits are O(1) — no scan or
// sort. The in-memory store never fails, so it always returns a nil error; ctx is
// unused but kept for parity with fail-able implementations (e.g. Redis).
func (s *InMemory) RecordFizzBuzzRequestHit(_ context.Context, req domain.GenerateFizzBuzzRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[req]++

	if s.counts[req] > s.topHits {
		s.topReq = req
		s.topHits = s.counts[req]
	}

	return nil
}

// GetFizzBuzzTopHits returns the most frequently requested parameters and its hit
// count. When no request has been recorded yet it returns the zero response
// (Hits == 0).
func (s *InMemory) GetFizzBuzzTopHits(_ context.Context) (domain.GetFizzBuzzTopHitsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return domain.GetFizzBuzzTopHitsResponse{RequestParams: s.topReq, Hits: s.topHits}, nil
}
