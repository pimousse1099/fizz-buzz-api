package main

import "sync"

// metricsCollector counts how many times each distinct fizz-buzz request has
// been served and keeps track of the most frequent one as requests come in. It
// is safe for concurrent use: every access goes through the mutex. The request
// struct is used directly as the map key (all of its fields are comparable), so
// no stringly-typed key is needed.
type metricsCollector struct {
	mu      sync.Mutex
	counts  map[fizzBuzzRequest]uint
	topReq  fizzBuzzRequest
	topHits uint
}

func newMetricsCollector() *metricsCollector {
	return &metricsCollector{counts: make(map[fizzBuzzRequest]uint)}
}

// record increments the counter for the given request and updates the running
// most-frequent request. Both this and top() are O(1) — no scan or sort.
func (mc *metricsCollector) record(req fizzBuzzRequest) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counts[req]++

	if mc.counts[req] > mc.topHits {
		mc.topReq = req
		mc.topHits = mc.counts[req]
	}
}

// top returns the most frequently requested parameters and its hit count. ok is
// false when no request has been recorded yet.
func (mc *metricsCollector) top() (req fizzBuzzRequest, hits uint, ok bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	return mc.topReq, mc.topHits, mc.topHits > 0
}
