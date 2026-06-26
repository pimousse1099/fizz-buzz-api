package main

import "sync"

type metricsCollector struct {
	RequestCounters RequestCounters
	mu              sync.Mutex
}

func (mc *metricsCollector) IncRequestCounter(reqBody string) {
	// check if request have already been made
	for _, rc := range mc.RequestCounters {
		if rc.ReqBody == reqBody {
			mc.mu.Lock()
			// increment existing request counter
			rc.Counter++
			mc.mu.Unlock()

			return
		}
	}
	// create new request counter entry
	mc.mu.Lock()
	mc.RequestCounters = append(mc.RequestCounters, &RequestCounter{ReqBody: reqBody, Counter: 1})
	mc.mu.Unlock()
}

type (
	RequestCounter struct {
		ReqBody string `json:"request_params"`
		Counter uint   `json:"nb_hits"`
	}
	RequestCounters []*RequestCounter
)

// Len is the number of elements in the collection.
func (rcs RequestCounters) Len() int {
	return len(rcs)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (rcs RequestCounters) Less(i, j int) bool {
	return rcs[i].Counter > rcs[j].Counter
}

// Swap swaps the elements with indexes i and j.
func (rcs RequestCounters) Swap(i, j int) {
	rcs[i], rcs[j] = rcs[j], rcs[i]
}
