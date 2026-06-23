package server

import "net/http"

// NewRouter builds the ServeMux with the application routes plus a health
// probe. Method+path patterns (Go 1.22+) yield 405 on method mismatch and 404
// on unknown paths automatically.
func NewRouter(generate, stats http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /fizzbuzz", generate)
	mux.HandleFunc("GET /fizzbuzz/stats", stats)

	return mux
}
