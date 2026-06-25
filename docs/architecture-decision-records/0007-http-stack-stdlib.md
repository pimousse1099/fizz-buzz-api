# 0007. HTTP Stack — stdlib `net/http` with Hand-Rolled Middleware

- **Status:** Superseded by [0014](0014-adopt-chi-router-and-middleware.md)
- **Date:** 2026-06-23

## Context

The service needs an HTTP transport. We must choose between the Go standard library and a
third-party router/framework. The `pure-go` branch intent favours minimal external dependencies.

## Decision

Use the standard library `net/http` (Go 1.22+ `ServeMux` with method+path routing) as the
sole HTTP layer. Write middlewares by hand (recovery, request-id, logging, rate-limit chain).
No web framework (gin/echo) and no third-party router (chi/gorilla).

**Rationale at the time:**
- With only five simple parameters, hand-written validation is clearer than struct-tag magic.
- Go 1.22 `ServeMux` with method routing makes chi/gin largely redundant.
- "Production-ready" is better demonstrated by a mastered, zero-magic stdlib stack.
- Consistent with the `pure-go` intent of the branch.

## Consequences

- Small dependency footprint.
- More hand-written plumbing: `Chain` helper, custom `Recovery` (did not correctly re-panic
  `http.ErrAbortHandler`), custom `RequestID`, manual route wiring.
- The hand-rolled `Recovery` middleware had an edge case (double-write risk on
  `http.ErrAbortHandler`). This became a direct motivation for [0014](0014-adopt-chi-router-and-middleware.md).
- **Superseded:** after review, chi + httplog + httprate replaced the hand-rolled transport layer
  (see [0014](0014-adopt-chi-router-and-middleware.md)). The domain, use-cases, and stat store were
  unaffected.
