package usecase

import (
	"context"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// StatReader reads the most frequently recorded request as a stats response,
// or fizzbuzz.ErrNoStatsRecorded when nothing has been recorded yet.
type StatReader interface {
	GetMostFrequentFizzbuzzRequest(ctx context.Context) (*fizzbuzz.GetStatsResponse, error)
}

// GetFizzBuzzStats is the application entry point for the statistics query.
type GetFizzBuzzStats struct {
	reader StatReader
}

// NewGetFizzBuzzStats builds the use-case with its reader.
func NewGetFizzBuzzStats(reader StatReader) *GetFizzBuzzStats {
	return &GetFizzBuzzStats{reader: reader}
}

// Execute returns the most frequent request and its hit count, propagating
// fizzbuzz.ErrNoStatsRecorded when there is nothing to report.
func (uc *GetFizzBuzzStats) Execute(ctx context.Context) (*fizzbuzz.GetStatsResponse, error) {
	return uc.reader.GetMostFrequentFizzbuzzRequest(ctx)
}
