package fizzbuzz

import "errors"

// ErrFailedToValidateGenerateRequest is wrapped by every validation failure of
// a GenerateRequest. Callers classify validation errors with errors.Is.
var ErrFailedToValidateGenerateRequest = errors.New("failed to validate generate request")

// ErrNoStatsRecorded is returned when no successful request has been recorded
// yet, so there is no "most frequent" request to report.
var ErrNoStatsRecorded = errors.New("failed to find recorded statistics")
