# 0016. Rate Limiting — `go-chi/httprate`

- **Status:** Accepted
- **Date:** 2026-06-27
- **Supersedes:** [0011](0011-rate-limiting-inhouse.md)

## Context

The in-house rate limiter ([0011](0011-rate-limiting-inhouse.md)) had a stub Redis backend
(`panic("implement me")`) and used a single global token bucket instead of per-IP limiting.
When chi was adopted ([0014](0014-adopt-chi-router-and-middleware.md)), a more correct and
production-ready replacement was evaluated.

## Decision

Use **`github.com/go-chi/httprate`** for rate limiting, keyed **per client IP**.

**Key properties:**
- `httprate.LimitByRealIP` limits N requests per window per client IP, where the IP is resolved by
  `httprate.KeyByRealIP`: `True-Client-IP`, then `X-Real-IP`, then the leftmost `X-Forwarded-For`
  entry, falling back to the socket address (`RemoteAddr`).
- Returns `429 Too Many Requests` with `Retry-After` and `X-RateLimit-*` headers automatically.
- The distributed backend is **real, not a stub**: `github.com/go-chi/httprate-redis` provides a
  `httprate.LimitCounter` that drops in via `httprate.Limit(..., httprate.WithLimitCounter(redisCounter))`.

**Default:** in-memory limiter (runs with no Redis dependency). Redis counter is the documented
scale-out path.

**Configuration:**
- `RATE_LIMIT_REQUESTS` — number of requests allowed per window.
- `RATE_LIMIT_WINDOW` — duration of the window.

### Trust assumption (security)

`KeyByRealIP` derives the client IP from request headers, which are **client-spoofable**. A client
can send `X-Forwarded-For` / `X-Real-IP` / `True-Client-IP` to land in a different bucket — bypassing
the limit, or poisoning another address's bucket. `LimitByRealIP` is therefore only sound **behind a
trusted proxy / ingress that sets and overwrites those headers and strips any inbound copies** (the
normal production topology). In a deployment with no such proxy, switch to `httprate.LimitByIP`,
which keys on the TCP socket address and cannot be spoofed.

This is the same class of issue that led chi to **deprecate its `middleware.RealIP`** (GHSA-3fxj-6jh8-hvhx
and related): we deliberately do **not** rewrite `RemoteAddr` with that middleware, and instead key
the limiter directly via `LimitByRealIP`. A spoofing-resistant alternative (trust only the N
right-most `X-Forwarded-For` entries, e.g. `realclientip-go`) was considered but not adopted, to
avoid an extra dependency and a "trusted proxy depth" config knob for what the edge already provides.

**Edge-is-authoritative stance preserved:** same guidance as [0011](0011-rate-limiting-inhouse.md)
— the authoritative rate limit for a horizontally scalable service belongs at the infrastructure /
edge layer. The in-app limiter remains a coarse per-instance safety net.

**Removed:** the entire `infrastructure/ratelimiter` package (in-memory + Redis stub) and the
`httpmiddleware.RateLimit`/`RateLimiter` interface seam.

**Dependencies added:** `github.com/go-chi/httprate`.
**Dependencies removed:** `golang.org/x/time/rate`.

## Consequences

- Per-IP limiting is semantically more correct than a global bucket: a single aggressive client
  is limited without throttling other clients, and behind a proxy the limit applies to the real
  client rather than collapsing onto the proxy's address.
- **The header-based IP resolution is only trustworthy behind a sanitising proxy** (see above). The
  deployment contract must guarantee it; otherwise `LimitByIP` is the safe choice.
- The `Retry-After` and `X-RateLimit-*` headers let clients implement proper back-off.
- The Redis backend is genuinely pluggable (not a panic stub), enabling real distributed limiting
  without code changes — only DI wiring.
- The hand-rolled `RateLimiter` interface and `infrastructure/ratelimiter` package are gone,
  reducing maintenance surface.
