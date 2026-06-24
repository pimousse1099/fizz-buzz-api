package httphandler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/reqctx"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// GenerateFizzBuzzRoute is the ServeMux pattern for the generate endpoint.
const GenerateFizzBuzzRoute = "GET /fizzbuzz"

var errInvalidQueryParam = errors.New("invalid or missing query parameter")

// GenerateFizzBuzz parses the query, runs the use-case and writes the result.
// Parsing and validation failures map to 400; unexpected failures to 500.
func GenerateFizzBuzz(uc *usecase.GenerateFizzBuzz, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := reqctx.Logger(r.Context(), logger).With("http_handler", "generate_fizzbuzz")

		req, err := parseGenerateRequest(r)
		if err != nil {
			l.Warn("failed to parse HTTP request", "error", err)
			writeError(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := uc.Execute(r.Context(), *req)
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
				l.Warn("rejected invalid request", "error", err)
				writeError(w, http.StatusBadRequest, err.Error())

				return
			}

			l.Error("failed to execute use-case", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		l.Debug("generated fizz-buzz sequence", "size", len(resp.Result))
		writeJSON(w, http.StatusOK, resp.Result)
	}
}

// parseGenerateRequest builds a GenerateRequest from the query parameters. It
// returns nil and a wrapped errInvalidQueryParam on the first malformed param.
func parseGenerateRequest(r *http.Request) (*fizzbuzz.GenerateRequest, error) {
	q := r.URL.Query()

	int1, err := parseIntParam(q, "int1")
	if err != nil {
		return nil, err
	}

	int2, err := parseIntParam(q, "int2")
	if err != nil {
		return nil, err
	}

	limit, err := parseIntParam(q, "limit")
	if err != nil {
		return nil, err
	}

	str1, err := parseStringParam(q, "str1")
	if err != nil {
		return nil, err
	}

	str2, err := parseStringParam(q, "str2")
	if err != nil {
		return nil, err
	}

	return &fizzbuzz.GenerateRequest{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}, nil
}

func parseIntParam(q url.Values, name string) (int, error) {
	raw := q.Get(name)
	if raw == "" {
		return 0, fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, errInvalidQueryParam)
	}

	return v, nil
}

func parseStringParam(q url.Values, name string) (string, error) {
	raw := q.Get(name)
	if raw == "" {
		return "", fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	return raw, nil
}
