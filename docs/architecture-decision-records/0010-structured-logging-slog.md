# 0010. Structured Logging — stdlib `log/slog`

- **Status:** Superseded by [0015](0015-request-logging-httplog.md)
- **Date:** 2026-06-23

## Context

The service needs structured logging for observability. We must pick a logger that is consistent
with the `pure-go` intent, supports JSON output, and can carry per-request context fields.

## Decision

Use the standard library `log/slog` (Go 1.21+) with a JSON handler rather than `logrus` or any
other third-party logger. The DI container builds one memoised logger tagged with base context
fields (reezoback naming convention):

- `application_name`
- `application_version`
- `environment_type`
- `environment_name` (when set)
- `host_name` (from `os.Hostname`)

Log level comes from config as a `slog.Level` (decoded by go-envconfig via
`encoding.TextUnmarshaler` — see [0009](0009-configuration-and-lifecycle.md)).

Request-scoped logging (per-request fields, duration) was handled by a hand-rolled logging
middleware as part of [0007](0007-http-stack-stdlib.md).

**Rationale at the time:** structured, zero-dependency, modern (Go 1.21+), consistent with the
`pure-go` intent. Base fields make every log line attributable to an app/version/environment/host.

## Consequences

- **Superseded:** when the chi router replaced the stdlib `ServeMux` (see [0014](0014-adopt-chi-router-and-middleware.md)),
  the hand-rolled logging middleware was removed and replaced by `go-chi/httplog/v2`
  (see [0015](0015-request-logging-httplog.md)).
- The underlying `*slog.Logger` from the DI container is still used for startup/shutdown logs and
  `http.Server.ErrorLog` — only the per-request middleware changed.
