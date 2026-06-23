package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

func TestNewRouter_RoutesAndHealthz(t *testing.T) {
	t.Parallel()

	generate := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("gen"))
	}
	stats := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("stats"))
	}

	router := server.NewRouter(generate, stats)

	cases := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{"/healthz", http.StatusOK, ""},
		{"/fizzbuzz", http.StatusOK, "gen"},
		{"/fizzbuzz/stats", http.StatusOK, "stats"},
		{"/unknown", http.StatusNotFound, ""},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, http.NoBody)) //nolint:noctx // test helper

			if rec.Code != tc.wantCode {
				t.Fatalf("path %s: status = %d, want %d", tc.path, rec.Code, tc.wantCode)
			}

			if tc.wantBody != "" && rec.Body.String() != tc.wantBody {
				t.Fatalf("path %s: body = %q, want %q", tc.path, rec.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestNewRouter_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	router := server.NewRouter(
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/fizzbuzz", http.NoBody)) //nolint:noctx // test helper

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
