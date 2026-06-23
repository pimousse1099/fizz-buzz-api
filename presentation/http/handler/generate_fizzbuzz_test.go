package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

const (
	buzz = "buzz"
	fizz = "fizz"
)

func newGenerateHandler() http.HandlerFunc {
	uc := usecase.NewGenerateFizzBuzz(10000, statstorer.NewInMemory())

	return handler.GenerateFizzBuzz(uc, slog.New(slog.DiscardHandler))
}

func TestGenerateFizzBuzz_OK(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=5&str1=fizz&str2=buzz", http.NoBody)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}

	var got []string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not a JSON array of strings: %v", err)
	}

	want := []string{"1", "2", fizz, "4", buzz}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestGenerateFizzBuzz_ValidationError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fizzbuzz?int1=0&int2=5&limit=5&str1=fizz&str2=buzz", http.NoBody)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGenerateFizzBuzz_MalformedQuery(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fizzbuzz?int1=abc&int2=5&limit=5&str1=fizz&str2=buzz", http.NoBody)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
