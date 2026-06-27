package httphandler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	ctxlog "github.com/go-chi/httplog/v2"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// GenerateFizzBuzzRoute is the route path for the generate endpoint.
const GenerateFizzBuzzRoute = "/fizzbuzz"

// Query parameter names for the generate endpoint. All are required.
const (
	queryParamInt1  = "int1"
	queryParamInt2  = "int2"
	queryParamLimit = "limit"
	queryParamStr1  = "str1"
	queryParamStr2  = "str2"
)

// errInvalidQueryParam is wrapped by every query-parsing failure so the handler
// can classify it as a client (400) error via errors.Is.
var errInvalidQueryParam = errors.New("failed to validate HTTP query parameter")

// GenerateFizzBuzz handles GET /fizzbuzz. All query params are required:
//   - int1, int2, limit : integers (int1/int2 must be positive, limit is bounded by config)
//   - str1, str2        : non-empty strings
//
// Parsing and validation failures map to 400; unexpected failures to 500.
func GenerateFizzBuzz(uc *usecase.GenerateFizzBuzz) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enrich the request-scoped log entry so downstream logs (use-case, etc.)
		// sharing this context also carry the http_handler field.
		ctxlog.LogEntrySetField(r.Context(), "http_handler", slog.StringValue("generate_fizzbuzz"))
		l := ctxlog.LogEntry(r.Context())

		req, err := parseGenerateRequest(r)
		if err != nil {
			l.Warn("failed to parse query parameters", "error", err)
			writeError(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := uc.Execute(r.Context(), *req)
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
				l.Warn("failed to validate request", "error", err)
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

	int1, err := parseIntParam(q, queryParamInt1)
	if err != nil {
		return nil, err
	}

	int2, err := parseIntParam(q, queryParamInt2)
	if err != nil {
		return nil, err
	}

	limit, err := parseIntParam(q, queryParamLimit)
	if err != nil {
		return nil, err
	}

	str1, err := parseStringParam(q, queryParamStr1)
	if err != nil {
		return nil, err
	}

	str2, err := parseStringParam(q, queryParamStr2)
	if err != nil {
		return nil, err
	}

	return &fizzbuzz.GenerateRequest{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}, nil
}

func parseIntParam(q url.Values, name string) (int, error) {
	raw := q.Get(name)
	if raw == "" {
		return 0, fmt.Errorf("%w: %s is required", errInvalidQueryParam, name)
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s must be an integer", errInvalidQueryParam, name)
	}

	return v, nil
}

func parseStringParam(q url.Values, name string) (string, error) {
	raw := q.Get(name)
	if raw == "" {
		return "", fmt.Errorf("%w: %s is required", errInvalidQueryParam, name)
	}

	return raw, nil
}
