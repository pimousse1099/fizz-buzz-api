// Package statsstorer provides storage for fizz-buzz request statistics.
package statsstorer

import (
	"sync"

	"github.com/pimousse1099/fizz_buzz_api/internal/domain"
)

// InMemory counts how many times each distinct request has been served and keeps
// track of the most frequent one as requests come in. It is safe for concurrent
// use: every access goes through the mutex. The request struct is used directly
// as the map key (all of its fields are comparable).
type InMemory struct {
	mu      sync.Mutex
	counts  map[domain.Request]uint
	topReq  domain.Request
	topHits uint
}

// NewInMemory returns a ready-to-use in-memory store.
func NewInMemory() *InMemory {
	return &InMemory{counts: make(map[domain.Request]uint)}
}

// Record increments the counter for req and updates the running most-frequent
// request. Both this and TopHits are O(1) — no scan or sort.
func (s *InMemory) Record(req domain.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[req]++

	if s.counts[req] > s.topHits {
		s.topReq = req
		s.topHits = s.counts[req]
	}
}

// TopHits returns the most frequently requested parameters and its hit count. ok
// is false when no request has been recorded yet.
func (s *InMemory) TopHits() (req domain.Request, hits uint, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.topReq, s.topHits, s.topHits > 0
}
