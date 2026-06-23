// Package fizzbuzz holds the fizz-buzz business model: its value objects,
// validation rules and sentinel errors. It depends on nothing above it (no
// use-case, transport or infrastructure imports). Generation logic lives in
// the use-case, not here.
package fizzbuzz

// GenerateRequest is the set of parameters of a fizz-buzz generation.
type GenerateRequest struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// GenerateResponse is the result of a fizz-buzz generation.
type GenerateResponse struct {
	Result []string
}

// GetStatsRequest carries no parameter: the statistics query takes none.
type GetStatsRequest struct{}

// GetStatsResponse is the most frequently requested generation and its hit count.
type GetStatsResponse struct {
	Request GenerateRequest
	Hits    int
}
