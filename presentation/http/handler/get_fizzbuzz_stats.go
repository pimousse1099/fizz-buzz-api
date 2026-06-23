package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type statsDTO struct {
	Request statsRequestDTO `json:"request"`
	Hits    int             `json:"hits"`
}

type statsRequestDTO struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

func newStatsDTO(resp fizzbuzz.GetStatsResponse) statsDTO {
	return statsDTO{
		Request: statsRequestDTO{
			Int1:  resp.Request.Int1,
			Int2:  resp.Request.Int2,
			Limit: resp.Request.Limit,
			Str1:  resp.Request.Str1,
			Str2:  resp.Request.Str2,
		},
		Hits: resp.Hits,
	}
}

// GetFizzBuzzStats returns the most frequent request as JSON, or 404 if none.
func GetFizzBuzzStats(uc *usecase.GetFizzBuzzStats, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp, err := uc.Execute(fizzbuzz.GetStatsRequest{})
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
				writeError(w, http.StatusNotFound, "no statistics recorded yet")

				return
			}

			logger.Error("get fizzbuzz stats failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		writeJSON(w, http.StatusOK, newStatsDTO(resp))
	}
}
