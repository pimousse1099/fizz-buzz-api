# 0011. Rate Limiting — In-House `x/time/rate` Guard

- **Status:** Superseded by [0016](0016-rate-limiting-httprate.md)
- **Date:** 2026-06-23

## Context

The service must not be saturated by a single client. We need an inbound rate-limiting strategy
that is honest about its limitations in a horizontally-scaled deployment.

## Decision

Ship a **local, per-instance guard** using `golang.org/x/time/rate` (token bucket) wrapped in a
`func(http.Handler) http.Handler` middleware that returns `429 Too Many Requests` when the bucket
is empty.

**Architecture stance:** the *authoritative* rate limit for a horizontally scalable service
belongs at the **infrastructure / edge layer** (API gateway, ingress, reverse-proxy, cloud load
balancer) or on a **shared store** (Redis sliding-window / token-bucket). The in-process limiter
is an explicit, coarse safety net for a single instance — not a correct global limit.

The limiter sat behind a `RateLimiter` interface (defined in the middleware package, where it is
consumed):
- `infrastructure/ratelimiter/in_memory.go` — `x/time/rate` implementation.
- `infrastructure/ratelimiter/redis.go` — placeholder with `panic("implement me")`.

**Known limitation (documented):** with N replicas, the effective global limit is N × the
configured rate; a client spread across instances is limited inconsistently.

**`x/time/rate` vs `go.uber.org/ratelimit`:** `x/time/rate` is a *token bucket* with
non-blocking `Allow()` — ideal to *reject* surplus traffic (429). `go.uber.org/ratelimit` is a
*leaky bucket* whose `Take()` *blocks* to smooth outbound throughput. Different goals:
protect-and-reject vs shape-and-smooth.

## Consequences

- **Superseded:** the entire `infrastructure/ratelimiter` package (including the Redis stub) was
  removed when `go-chi/httprate` was adopted (see [0016](0016-rate-limiting-httprate.md)).
  `httprate` provides per-IP limiting and a real (not stub) Redis backend via `httprate-redis`.
- The edge-is-authoritative stance is preserved in [0016](0016-rate-limiting-httprate.md).
