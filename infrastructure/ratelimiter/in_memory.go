// Package ratelimiter holds rate-limiter implementations used by the HTTP
// middleware: a process-local token-bucket limiter and a Redis placeholder.
//
// A process-local limiter is a per-instance guard only. In a horizontally
// scaled deployment the authoritative limit belongs at the edge (gateway /
// ingress / load balancer) or a shared store; see the project ADR.
package ratelimiter

import (
	"context"

	"golang.org/x/time/rate"
)

// InMemory is a token-bucket limiter backed by golang.org/x/time/rate.
type InMemory struct {
	limiter *rate.Limiter
}

// NewInMemory builds a limiter allowing perSecond sustained requests with the
// given burst capacity.
func NewInMemory(perSecond float64, burst int) *InMemory {
	return &InMemory{limiter: rate.NewLimiter(rate.Limit(perSecond), burst)}
}

// Allow reports whether a request may proceed now (non-blocking).
func (l *InMemory) Allow(_ context.Context) bool {
	return l.limiter.Allow()
}
