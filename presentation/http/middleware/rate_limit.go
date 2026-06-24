package httpmiddleware

import (
	"context"
	"net/http"
)

// RateLimiter decides whether a request may proceed now.
type RateLimiter interface {
	Allow(ctx context.Context) bool
}

// RateLimit rejects requests with 429 when the limiter denies them.
func RateLimit(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow(r.Context()) {
				http.Error(w, "too many requests", http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
