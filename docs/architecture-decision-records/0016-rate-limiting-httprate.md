# 0016. Rate Limiting — `go-chi/httprate`

- **Status:** Accepted
- **Date:** 2026-06-24
- **Supersedes:** [0011](0011-rate-limiting-inhouse.md)

## Context

The in-house rate limiter ([0011](0011-rate-limiting-inhouse.md)) had a stub Redis backend
(`panic("implement me")`) and used a single global token bucket instead of per-IP limiting.
When chi was adopted ([0014](0014-adopt-chi-router-and-middleware.md)), a more correct and
production-ready replacement was evaluated.

## Decision

Use **`github.com/go-chi/httprate`** for rate limiting.

**Key properties:**
- `httprate.LimitByIP` limits N requests per window **per client IP** — more correct than the
  previous single global token-bucket.
- Returns `429 Too Many Requests` with `Retry-After` and `X-RateLimit-*` headers automatically.
- The distributed backend is **real, not a stub**: `github.com/go-chi/httprate-redis` provides a
  `httprate.LimitCounter` that drops in via `httprate.Limit(..., httprate.WithLimitCounter(redisCounter))`.

**Default:** in-memory limiter (runs with no Redis dependency). Redis counter is the documented
scale-out path.

**Configuration:**
- `RATE_LIMIT_REQUESTS` — number of requests allowed per window.
- `RATE_LIMIT_WINDOW` — duration of the window.

**Edge-is-authoritative stance preserved:** same guidance as [0011](0011-rate-limiting-inhouse.md)
— the authoritative rate limit for a horizontally scalable service belongs at the infrastructure /
edge layer. The in-app limiter remains a coarse per-instance safety net.

**Removed:** the entire `infrastructure/ratelimiter` package (in-memory + Redis stub) and the
`httpmiddleware.RateLimit`/`RateLimiter` interface seam.

**Dependencies added:** `github.com/go-chi/httprate`.
**Dependencies removed:** `golang.org/x/time/rate`.

## Consequences

- Per-IP limiting is semantically more correct than a global bucket: a single aggressive client
  is limited without throttling other clients.
- The `Retry-After` and `X-RateLimit-*` headers let clients implement proper back-off.
- The Redis backend is genuinely pluggable (not a panic stub), enabling real distributed limiting
  without code changes — only DI wiring.
- The hand-rolled `RateLimiter` interface and `infrastructure/ratelimiter` package are gone,
  reducing maintenance surface.
