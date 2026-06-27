# 0005. Error-to-HTTP-Status Mapping

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The domain defines typed errors; the HTTP handler must translate them into appropriate status
codes. We need a clear, maintainable mapping that keeps validity rules in the domain and
transport semantics in the presentation layer.

## Decision

The domain defines **sentinel errors** (prefixed `Err`, per the `errname` linter) in
`domain/fizzbuzz/error.go`. Callers wrap them with `%w` (per the `err113` linter). The handler
classifies them via `errors.Is`:

| Error / condition | HTTP status |
|---|---|
| `fizzbuzz.ErrFailedToValidateGenerateRequest` (wrapped with field detail) | `400 Bad Request` |
| Query-param parse failures (non-integer, missing parameter) | `400 Bad Request` |
| `fizzbuzz.ErrNoStatsRecorded` (no request counted yet) | `404 Not Found` |
| Any other / unexpected error | `500 Internal Server Error` |

**Example wrapping in the domain:**
```go
fmt.Errorf("int1 must be a positive integer: %w", ErrFailedToValidateGenerateRequest)
```

The field-specific detail travels in the wrap message; the sentinel carries the classification
for `errors.Is`. No `ValidationError` struct is used — a wrapped sentinel satisfies `err113`
(which discourages bare `errors.New`/`fmt.Errorf` without a wrapped sentinel).

**Error message convention:**
- Operational/wrapped errors: `failed to …` (e.g. `failed to shut down http server`).
- User-facing validation messages: field-specific (`int1 must be a positive integer`).

## Consequences

- Validity rules stay in the domain; transport translation stays in the handler — clean separation.
- `errors.Is` unwraps the chain, so intermediate wrapping (e.g. use-case adding context) does
  not break the mapping.
- The 404 for "no stats recorded" is semantically correct: there is no "most-frequent" resource
  to return yet.
- Adding a new domain error requires only: define the sentinel, wrap it, add an `errors.Is`
  branch in the handler.
