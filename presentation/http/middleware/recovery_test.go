package httpmiddleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRecovery_TurnsPanicInto500(t *testing.T) {
	t.Parallel()

	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	h := httpmiddleware.Recovery(discardLogger())(panicky)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}
