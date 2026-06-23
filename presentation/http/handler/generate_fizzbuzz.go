package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// GenerateFizzBuzz parses the query, runs the use-case and writes the result.
// Parsing and validation failures map to 400; unexpected failures to 500.
func GenerateFizzBuzz(uc *usecase.GenerateFizzBuzz, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseGenerateRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := uc.Execute(r.Context(), req)
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
				writeError(w, http.StatusBadRequest, err.Error())

				return
			}

			logger.Error("generate fizzbuzz failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		writeJSON(w, http.StatusOK, resp.Result)
	}
}
