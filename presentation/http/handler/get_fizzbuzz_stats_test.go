package handler_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

func TestGetFizzBuzzStats_Empty404(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetFizzBuzzStats(statstorer.NewInMemory())
	h := handler.GetFizzBuzzStats(uc, slog.New(slog.DiscardHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fizzbuzz/stats", http.NoBody))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestGetFizzBuzzStats_OK(t *testing.T) {
	t.Parallel()

	store := statstorer.NewInMemory()
	store.Record(context.Background(), fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: fizz, Str2: buzz})

	uc := usecase.NewGetFizzBuzzStats(store)
	h := handler.GetFizzBuzzStats(uc, slog.New(slog.DiscardHandler))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/fizzbuzz/stats", http.NoBody))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Request struct {
			Int1  int    `json:"int1"`
			Int2  int    `json:"int2"`
			Limit int    `json:"limit"`
			Str1  string `json:"str1"`
			Str2  string `json:"str2"`
		} `json:"request"`
		Hits int `json:"hits"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if body.Hits != 1 || body.Request.Int1 != 3 || body.Request.Str2 != buzz {
		t.Fatalf("unexpected body: %+v", body)
	}
}
