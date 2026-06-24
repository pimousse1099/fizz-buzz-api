package httpmiddleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/middleware"
)

func TestRequestID_SetsHeader(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := httpmiddleware.RequestID(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", http.NoBody))

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}
