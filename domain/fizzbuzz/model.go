// Package fizzbuzz holds the fizz-buzz business model: its value objects,
// validation rules and sentinel errors. It depends on nothing above it (no
// use-case, transport or infrastructure imports). Generation logic lives in
// the use-case, not here.
//
// The JSON tags make these value objects double as the API wire contract,
// which is sufficient for this service; a separate DTO would be introduced only
// if the wire format had to diverge from the model.
package fizzbuzz

// GenerateRequest is the set of parameters of a fizz-buzz generation.
type GenerateRequest struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

// GenerateResponse is the result of a fizz-buzz generation.
type GenerateResponse struct {
	Result []string `json:"result"`
}

// GetStatsResponse is the most frequently requested generation and how many
// times it was requested.
type GetStatsResponse struct {
	Request   GenerateRequest `json:"request"`
	TotalHits int             `json:"total_hits"`
}
