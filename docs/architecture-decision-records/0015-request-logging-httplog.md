# 0015. Request Logging — `go-chi/httplog/v2`

- **Status:** Accepted
- **Date:** 2026-06-24
- **Supersedes:** [0010](0010-structured-logging-slog.md)

## Context

When the hand-rolled logging middleware was removed as part of adopting chi ([0014](0014-adopt-chi-router-and-middleware.md)),
a replacement for structured per-request logging was needed. The replacement must integrate with
the slog-backed logger already in the DI container and allow handlers and use-cases to enrich the
log entry with contextual fields.

## Decision

Use **`github.com/go-chi/httplog/v2`** for structured per-request logging.

**Configuration:** `httplog.RequestLogger` middleware is configured from `config` (JSON format,
log level, base fields as `Tags`).

**Request-scoped enrichment:** handlers and use-cases obtain the request-scoped log entry via
`httplog.LogEntry(ctx)` and enrich it with contextual fields:

```go
// aliased import
import ctxlog "github.com/go-chi/httplog/v2"

// in a handler or use-case
ctxlog.LogEntry(ctx).Set("http_handler", "generate_fizzbuzz")
ctxlog.LogEntry(ctx).Set("use_case", "generate_fizzbuzz")
```

The import is aliased as `ctxlog` to express intent (context-scoped log enrichment) and to
distinguish it from the base `*slog.Logger` held by the DI container.

**Continued use of `*slog.Logger`:** the DI container still exposes the underlying `*slog.Logger`
(`GetLogger`) for startup/shutdown lifecycle logs and `http.Server.ErrorLog`. Only the
per-request middleware path changed.

**Dependencies added:** `github.com/go-chi/httplog/v2`.

## Consequences

- Per-request logs include duration, status code, method, path, and any fields added via
  `ctxlog.LogEntry(ctx).Set(...)` — without any hand-rolled accumulation logic.
- Handlers and use-cases can enrich the same log entry without passing a logger explicitly
  through every function signature.
- The `sloglint` linter ([0012](0012-linting-golangci-lint.md)) continues to enforce consistent
  slog usage; httplog/v2 is slog-backed so there is no conflict.
- The hand-rolled logging middleware and `presentation/http/reqctx` package are gone.
