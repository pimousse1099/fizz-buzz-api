package ratelimiter

// Redis is a placeholder for a distributed rate limiter backed by a shared
// store, which is what an authoritative global limit requires when scaling out.
type Redis struct{}

// NewRedis builds the placeholder.
func NewRedis() *Redis {
	return &Redis{}
}

// Allow is not implemented yet.
func (l *Redis) Allow() bool {
	panic("implement me: distributed rate limiting via a shared store")
}
