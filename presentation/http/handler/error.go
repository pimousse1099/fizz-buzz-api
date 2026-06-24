package httphandler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// writeServiceError maps a handler/use-case error to an HTTP response: client
// errors (bad query params, invalid request) become 400, an empty stats store
// becomes 404, and anything else is logged and returned as a generic 500.
func writeServiceError(w http.ResponseWriter, l *slog.Logger, err error) {
	switch {
	case errors.Is(err, errInvalidQueryParam),
		errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest):
		l.Warn("rejected client request", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, fizzbuzz.ErrNoStatsRecorded):
		l.Info("no statistics recorded yet")
		writeError(w, http.StatusNotFound, "no statistics recorded yet")
	default:
		l.Error("failed to execute use-case", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
