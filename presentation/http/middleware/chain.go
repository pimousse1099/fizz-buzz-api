// Package httpmiddleware provides the HTTP middleware stack (recovery,
// request-id, logging, rate-limit) and a helper to compose it. It lives in the
// directory presentation/http/middleware.
package httpmiddleware

import "net/http"

// Chain wraps h with the given middlewares so the first listed is outermost.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}
