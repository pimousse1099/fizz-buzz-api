package usecase

import (
	"context"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// StatReader reads the most frequently recorded generation request.
type StatReader interface {
	MostFrequent(ctx context.Context) (req fizzbuzz.GenerateRequest, hits int, ok bool)
}

// GetFizzBuzzStats returns the most frequent request and its hit count.
type GetFizzBuzzStats struct {
	reader StatReader
}

// NewGetFizzBuzzStats builds the use-case with its reader.
func NewGetFizzBuzzStats(reader StatReader) *GetFizzBuzzStats {
	return &GetFizzBuzzStats{reader: reader}
}

// Execute returns the most frequent request, or ErrNoStatsRecorded if none.
func (uc *GetFizzBuzzStats) Execute(ctx context.Context, _ fizzbuzz.GetStatsRequest) (fizzbuzz.GetStatsResponse, error) {
	req, hits, ok := uc.reader.MostFrequent(ctx)
	if !ok {
		return fizzbuzz.GetStatsResponse{}, fizzbuzz.ErrNoStatsRecorded
	}

	return fizzbuzz.GetStatsResponse{Request: req, Hits: hits}, nil
}
