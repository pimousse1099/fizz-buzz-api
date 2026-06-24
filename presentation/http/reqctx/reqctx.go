// Package reqctx carries request-scoped values (the request id) through the
// context and builds request-scoped loggers from them. It is shared by the HTTP
// middleware (which sets the values) and the handlers (which read them), so
// neither layer depends on the other for request context.
package reqctx

import (
	"context"
	"log/slog"
)

type contextKey int

const requestIDKey contextKey = iota

// WithRequestID returns a copy of ctx carrying the request id.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID returns the request id stored in ctx, or "" if none.
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)

	return id
}

// Logger returns base enriched with the request id from ctx (when present), so
// handlers can log with request-scoped context.
func Logger(ctx context.Context, base *slog.Logger) *slog.Logger {
	if id := RequestID(ctx); id != "" {
		return base.With("request_id", id)
	}

	return base
}
