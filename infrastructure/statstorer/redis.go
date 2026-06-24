package statstorer

import (
	"context"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// Redis is a placeholder for a durable, shared stat store. It demonstrates how
// a distributed backend would plug into the same StatRecorder/StatReader
// interfaces without touching the use-cases.
type Redis struct{}

// NewRedis builds the placeholder.
func NewRedis() *Redis {
	return &Redis{}
}

// RecordFizzBuzzStat is not implemented yet.
func (s *Redis) RecordFizzBuzzStat(_ context.Context, _ fizzbuzz.GenerateRequest) error {
	panic("implement me: durable stat recording via Redis")
}

// GetMostFrequentFizzbuzzRequest is not implemented yet.
func (s *Redis) GetMostFrequentFizzbuzzRequest(_ context.Context) (*fizzbuzz.GetStatsResponse, error) {
	panic("implement me: durable stat reading via Redis")
}
