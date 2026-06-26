# fizz-buzz-api

A fizz-buzz REST API. The endpoint takes five parameters — `int1`, `int2`, `limit`, `str1`, `str2` —
and returns the list from 1 to `limit` where multiples of `int1` become `str1`, multiples of `int2`
become `str2`, multiples of both become `str1str2`, and every other number is itself. A statistics
endpoint reports the most frequently requested call.

It comes in **two implementations** — pick the one that fits your needs.

## simple

- PR: https://github.com/pimousse1099/fizz-buzz-api/pull/1
- Branch: https://github.com/pimousse1099/fizz-buzz-api/tree/simple-version

A lean, pragmatic implementation on the [echo v5](https://echo.labstack.com/) framework, organised
into small `internal/` packages (`domain`, `http`, `statsstorer`). Structured JSON logging via
`log/slog` (with a per-request `request_id`), a sensible middleware stack (rate limiting, body-size
cap, request timeout, gzip, security headers), graceful shutdown, and a distroless container image.
CI runs lint, race tests and image builds. It reads top to bottom in one sitting; its remaining
trade-offs are documented honestly in the branch README.

## clean architecture

- PR: https://github.com/pimousse1099/fizz-buzz-api/pull/3
- Branch: https://github.com/pimousse1099/fizz-buzz-api/tree/clean-archi-2026

A hexagonal / clean-architecture implementation: dependencies flow inward (domain → use cases →
presentation), wired by a dependency-injection container. Adds OpenTelemetry tracing, liveness and
readiness probes, stores hidden behind interfaces (swappable, e.g. in-memory or Redis), and an
architecture-decision-record per design choice. More moving parts, but the most production-hardened
and extensible option.
