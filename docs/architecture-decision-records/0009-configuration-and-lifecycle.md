# 0009. Configuration and Lifecycle

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The service must be configurable without recompilation (12-factor) and must start and stop
cleanly. We need a configuration strategy and a lifecycle model for the process.

## Decision

### Configuration — `sethvargo/go-envconfig`

Configuration is loaded from environment variables via `sethvargo/go-envconfig` into a struct
grouped by concern, with env-var prefix per sub-struct:

| Sub-struct | Prefix | Required vars |
|---|---|---|
| `Env` | `ENV_` | `ENV_TYPE` |
| `HTTP` | `HTTP_` | `HTTP_ADDR` |
| `FizzBuzz` | `FIZZBUZZ_` | `FIZZBUZZ_MAX_SEQUENCE_LENGTH` |
| `Log` | `LOG_` | `LOG_LEVEL` |
| `Tracing` | `TRACING_` | — (optional, see [0018](0018-distributed-tracing-opentelemetry.md)) |

Rate-limit knobs live inside the `HTTP` sub-struct (`HTTP_RATE_LIMIT_REQUESTS`,
`HTTP_RATE_LIMIT_WINDOW`; defaults provided, see [0016](0016-rate-limiting-httprate.md)).

`AppName` / `AppVersion` are compile-time constants / ldflags-injected vars.

**Required vars have no defaults** — this forces explicit values at deployment time and makes
misconfiguration fail fast at startup, not silently at runtime.

`LOG_LEVEL` decodes straight into `slog.Level` via its `encoding.TextUnmarshaler` — parsing
lives in config, not in the logger.

Operational knobs (timeouts, rate-limit) keep sensible defaults.

`http.Server` is configured with `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout`,
`MaxHeaderBytes`, a `BaseContext` returning the container's base context (so in-flight connections
inherit it), and an `ErrorLog` routed through slog at error level (`slog.NewLogLogger`).

### Lifecycle (`cmd/main.go`)

```
startup banner (slog)
  → load config         (failure → log + os.Exit(1))
  → single base context context.WithCancel(Background)
  → build DI container  (base ctx passed in)
  → httpSrv.Start(errChan)
  → wait: SIGINT/SIGTERM or server error
  → graceful Stop(ctx)  (timeout derived from base context)
```

Lifecycle logs (starting on addr / signal received / shutting down) live in `main`; the `httpserver`
package (`server/server.go`) stays purely mechanical (Start/Stop, no log statements).

## Consequences

- Required env vars without defaults make misconfiguration loud and early — no silent half-initialised
  server.
- The single base context flows from `main` through the DI container into `http.Server.BaseContext`,
  ensuring clean cancellation propagation on shutdown.
- Grouping config by concern (Env/HTTP/FizzBuzz/Log/Tracing) keeps the struct navigable and maps
  naturally to deployment secret scopes.
- `go-envconfig` is the only configuration dependency; the stdlib has no equivalent for
  struct↔env mapping.
