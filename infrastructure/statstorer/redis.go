package statstorer

import "github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"

// Redis is a placeholder for a durable, shared stat store. It demonstrates how
// a distributed backend would plug into the same StatRecorder/StatReader
// interfaces without touching the use-cases.
type Redis struct{}

// NewRedis builds the placeholder.
func NewRedis() *Redis {
	return &Redis{}
}

// Record is not implemented yet.
func (s *Redis) Record(_ fizzbuzz.GenerateRequest) {
	panic("implement me: durable stat recording via Redis")
}

// MostFrequent is not implemented yet.
func (s *Redis) MostFrequent() (fizzbuzz.GenerateRequest, int, bool) {
	panic("implement me: durable stat reading via Redis")
}
