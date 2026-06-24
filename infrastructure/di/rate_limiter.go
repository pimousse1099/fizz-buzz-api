package di

import (
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/ratelimiter"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
)

// getRateLimiter returns the memoized rate limiter (local per-instance guard).
func (c *Container) getRateLimiter() httpmiddleware.RateLimiter {
	if c.rateLimiter == nil {
		c.rateLimiter = ratelimiter.NewInMemory(
			c.config.HTTP.RateLimitPerSec,
			c.config.HTTP.RateLimitBurst,
		)
	}

	return c.rateLimiter
}
