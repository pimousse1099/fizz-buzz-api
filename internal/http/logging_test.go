package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	slogctx "github.com/veqryn/slog-context"

	"github.com/pimousse1099/fizz-buzz-api/internal/domain"
	httpserver "github.com/pimousse1099/fizz-buzz-api/internal/http"
)

// failingStore makes RecordFizzBuzzRequestHit fail, to exercise the handler's
// best-effort warning path.
type failingStore struct{}

func (failingStore) RecordFizzBuzzRequestHit(context.Context, domain.GenerateFizzBuzzRequest) error {
	return errors.New("boom")
}

func (failingStore) GetFizzBuzzTopHits(context.Context) (domain.GetFizzBuzzTopHitsResponse, error) {
	return domain.GetFizzBuzzTopHitsResponse{}, nil
}

// TestRecordFailureLogIsRequestCorrelated proves the end-to-end wiring: the
// requestID middleware appends the id to the context, and slogctx stamps it onto
// the handler's WarnContext line.
func TestRecordFailureLogIsRequestCorrelated(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	logger := slog.New(slogctx.NewHandler(slog.NewJSONHandler(&buf, nil), nil))
	e := httpserver.New(logger, failingStore{}, testHTTPConfig(), testMaxLimit)

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/fizz-buzz?int1=2&int2=3&limit=10&str1=fizz&str2=buzz", http.NoBody)
	response := httptest.NewRecorder()
	e.ServeHTTP(response, request)

	var warn map[string]any

	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		entry := map[string]any{}

		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			continue
		}

		if entry["msg"] == "failed to record fizzbuzz request hit" {
			warn = entry
		}
	}

	check := assert.New(t)
	check.NotNil(warn, buf.String())
	check.NotEmpty(warn["request_id"])
}
