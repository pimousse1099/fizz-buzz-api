package httphandler

import (
	"errors"
	"log/slog"
	"net/http"

	ctxlog "github.com/go-chi/httplog/v2"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// GetFizzBuzzStatsRoute is the route path for the statistics endpoint.
const GetFizzBuzzStatsRoute = "/fizzbuzz/stats"

// GetFizzBuzzStats returns the most frequent request as JSON, or 404 if none.
// The domain GetStatsResponse is serialized directly (it carries the JSON tags).
func GetFizzBuzzStats(uc *usecase.GetFizzBuzzStats) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enrich the request-scoped log entry so downstream logs sharing this
		// context also carry the http_handler field.
		ctxlog.LogEntrySetField(r.Context(), "http_handler", slog.StringValue("get_fizzbuzz_stats"))
		l := ctxlog.LogEntry(r.Context())

		resp, err := uc.Execute(r.Context())
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
				l.Info("failed to find recorded statistics")
				writeError(w, http.StatusNotFound, "no statistics recorded yet")

				return
			}

			l.Error("failed to execute use-case", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		l.Debug("returning most frequent request", "total_hits", resp.TotalHits)
		writeJSON(w, http.StatusOK, resp)
	}
}
