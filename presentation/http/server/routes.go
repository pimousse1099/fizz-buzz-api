package httpserver

import "net/http"

// Route binds a ServeMux pattern (Go 1.22+ method+path, e.g. "GET /fizzbuzz") to
// its handler. Business routes are declared as constants next to their handler.
type Route struct {
	Pattern string
	Handler http.HandlerFunc
}

// NewRouter builds the ServeMux from the given business routes plus an
// operational health probe at GET /healthz.
func NewRouter(routes ...Route) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, route := range routes {
		mux.HandleFunc(route.Pattern, route.Handler)
	}

	return mux
}
