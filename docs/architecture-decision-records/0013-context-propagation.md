# 0013. Context Propagation

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

Production services must honour request cancellation, deadlines, and distributed trace propagation.
`context.Context` must flow from the HTTP handler down to every I/O boundary. We need a clear
policy on which functions receive a context and which legitimately do not.

## Decision

`context.Context` is threaded from the HTTP handler down through the use-case to every I/O
boundary. The handler passes `r.Context()`; the affected signatures take `ctx` as their first
parameter:

```go
// Use-cases
func (uc *GenerateFizzBuzz) Execute(ctx context.Context, req fizzbuzz.GenerateRequest) (fizzbuzz.GenerateResponse, error)
func (uc *GetFizzBuzzStats) Execute(ctx context.Context, req fizzbuzz.GetStatsRequest) (fizzbuzz.GetStatsResponse, error)

// Ports
type StatRecorder interface {
    Record(ctx context.Context, req fizzbuzz.GenerateRequest) error
}
type StatReader interface {
    MostFrequent(ctx context.Context) (fizzbuzz.GenerateRequest, int, bool)
}
```

**Pure domain compute stays ctx-free:** `GenerateRequest.Validate(maxLimit int)` and the
fizz-buzz generation loop perform no I/O and gain nothing from a context. Adding context to pure
functions is non-idiomatic Go — they keep no `ctx` parameter.

**In-memory implementations accept `ctx`** to satisfy the interface contract even though they
ignore it. This means network-bound implementations (Redis stat store, distributed rate limiter)
drop in without any signature changes.

## Consequences

- Cancellation, deadlines, and trace propagation flow with every request through every layer
  (handler → use-case → store).
- The port interfaces already carry `ctx`, so the Redis implementations of `StatRecorder` and
  `StatReader` can be wired in without touching the use-cases or handlers — the seams are ready.
- Pure domain functions remain clean and ctx-free, which is idiomatic and avoids misleading callers
  into thinking I/O is happening inside validation or generation.
