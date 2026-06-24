package httpmiddleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
)

type stubLimiter struct{ allow bool }

func (s stubLimiter) Allow(context.Context) bool { return s.allow }

func TestRateLimit_Rejects(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := httpmiddleware.RateLimit(stubLimiter{allow: false})(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
}

func TestRateLimit_Allows(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := httpmiddleware.RateLimit(stubLimiter{allow: true})(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody))

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418 (passed through)", rec.Code)
	}
}
