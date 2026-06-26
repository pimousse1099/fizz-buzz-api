package main

import "sync"

// metricsCollector counts how many times each distinct fizz-buzz request has
// been served. It is safe for concurrent use: every access goes through the
// mutex. The request struct is used directly as the map key (all of its fields
// are comparable), so no stringly-typed key is needed.
type metricsCollector struct {
	mu     sync.Mutex
	counts map[fizzBuzzRequest]uint
}

func newMetricsCollector() *metricsCollector {
	return &metricsCollector{counts: make(map[fizzBuzzRequest]uint)}
}

// record increments the counter for the given request.
func (mc *metricsCollector) record(req fizzBuzzRequest) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.counts[req]++
}

// top returns the most frequently requested parameters and its hit count. ok is
// false when no request has been recorded yet.
func (mc *metricsCollector) top() (req fizzBuzzRequest, hits uint, ok bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for r, c := range mc.counts {
		if c > hits {
			req, hits, ok = r, c, true
		}
	}

	return req, hits, ok
}
