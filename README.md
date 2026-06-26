# fizz-buzz-api

A straightforward REST API implementation of fizz-buzz.

It is built on the [echo v5](https://echo.labstack.com/) web framework, so request binding,
validation and HTTP middlewares come out of the box rather than being hand-rolled. Logging is
structured JSON via the standard library [`log/slog`](https://pkg.go.dev/log/slog) (echo v5 logs
through slog natively) and is written to stdout. The server listens on a fixed `:8080` (see
limitations) and wires the following middleware stack (outermost first): `Recover`, `RequestID`, a
slog-backed request logger, `Secure` (security headers), `CORS` (permissive, origin `*`), `Gzip`,
`BodyLimit` (1 MiB), a per-IP `RateLimiter` (20 req/s) and a per-request `ContextTimeout` (10s).
Shutdown is graceful: `SIGINT`/`SIGTERM` drains in-flight requests (10s budget) before exiting.

The codebase targets Go 1.26 and is organised into small, single-responsibility packages (see
[project layout](#project-layout)).

## usage

### prod

```sh
docker run --rm -p 8080:8080 pimousse1099/fizz-buzz-api-go:v0.1.0
```

### dev

```sh
make run
# or
go run ./cmd/fizz-buzz-api
```

## example

The main endpoint is `GET /fizz-buzz` and takes its five required parameters as query-string
values: three unsigned integers `int1`, `int2`, `limit` and two strings `str1`, `str2`. It returns
a JSON array of strings from 1 to `limit` where multiples of `int1` become `str1`, multiples of
`int2` become `str2`, multiples of both become `str1str2`, and every other number is itself. Every
parameter is required (non-empty strings, non-zero integers); a missing or invalid parameter
returns HTTP `400` with the validator error message.

### CURL

```sh
curl --location 'localhost:8080/fizz-buzz?int1=3&int2=4&limit=20&str1=fizz&str2=buzz'
```

### Go

```go
package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	url := "http://localhost:8080/fizz-buzz?int1=3&int2=4&limit=20&str1=fizz&str2=buzz"

	res, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}
```

Both snippets return the following response body:

```json
["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13","14","fizz","buzz","17","fizz","19","buzz"]
```

## metrics and logs

`GET /metrics/top-hits` returns the statistics for the **most frequently requested** fizz-buzz call:
its parameters and how many times it was made, as `{"request_params": { ... }, "nb_hits": N}`
(`nb_hits` is `0` with zero-valued params until a request has been served). The counters are kept in a
mutex-guarded map keyed by the request parameters, so the endpoint is safe under concurrency. This is
custom JSON and **not** Prometheus exposition format — see the limitations below.

Logs are structured JSON (`log/slog`) written to stdout via echo's `RequestLogger` middleware — one
line per request (`method`, `uri`, `status`, `latency`, a correlation `request_id`, and more), at
`INFO` on success and `ERROR` on failure (with the error attached). The `request_id` is also returned
as the `X-Request-Id` response header. A sidecar (fluentd, filebeat, ...) is the suggested way to
ship the logs into a log service (graylog, logstash, ...).

## CI/CD

GitHub Actions runs four workflows: `lint` (golangci-lint v2), `test` (`go build` + `go vet` +
`go test -race`), `build-pull-request` (push a branch-name-tagged image to GHCR on PRs) and
`build-release` (push semver + `latest` tags to GHCR on release). Dependabot
(`.github/dependabot.yml`: gomod, docker and github-actions, weekly) is paired with an auto-merge
workflow that auto-approves every update and auto-merges patch/minor bumps, leaving major bumps for
human review.

The container image is a multi-stage build: a `golang:1.26-alpine` builder produces a static,
CGO-disabled, `-trimpath` binary that runs on `gcr.io/distroless/static:nonroot` (non-root uid
65532, ships ca-certs, no shell or package manager). The Dockerfile sets no env defaults and does
not `EXPOSE` a port — the operator provides everything at runtime.

## project layout

```
cmd/fizz-buzz-api/      composition root: builds dependencies, handles signals, runs the server
internal/
  domain/              fizz-buzz request/response types and the Generate logic (no transport/storage)
  http/                echo server setup, handlers, middlewares, and the StatsStorer interface
  statsstorer/         in-memory StatsStorer implementation
```

Dependencies point inward: `domain` has no internal imports; `http` and `statsstorer` depend on
`domain`; `cmd` wires everything together. The `http` package declares the `StatsStorer` interface it
needs (consumer-side), and `statsstorer.InMemory` satisfies it — so the store can be swapped (e.g.
for a Redis-backed one) without touching the HTTP layer.

## limitations

This is the **simple** implementation: small and intentionally lean. The gaps below are inherent
trade-offs of that choice, documented so they are explicit. A fully hardened design — with bounded
inputs, a persistent stats store, env-based config, health/readiness probes and distributed tracing —
lives on the `clean-archi-2026` branch.

### Production-readiness

| # | Gap | Impact |
|---|-----|--------|
| 1 | **`limit` is unbounded** — the response slice is sized directly from `limit` (a `uint`) | A single very large `limit` allocates a correspondingly huge slice → memory exhaustion / DoS. No upper bound, pagination or streaming. The 1 MiB body cap does not constrain a query-string `limit`. |
| 2 | **Unbounded stats growth** — the stats map keeps one entry per distinct parameter tuple, never evicted (the most-frequent lookup itself is O(1)) | High-cardinality clients drive memory growth over time. |
| 3 | **Hardcoded configuration** — the listen address is `:8080` with no environment-variable configuration | Conflicts with 12-factor config; the port cannot change and multiple instances cannot run without recompiling. |
| 4 | **No health / readiness probes** | Orchestrators (Kubernetes, ECS) cannot gauge liveness or readiness. |
| 5 | **Limited observability** — structured slog request logs with a correlation id are emitted, but there is no distributed tracing and statistics are custom JSON, not Prometheus exposition format | Per-request latency/errors are visible in logs, but there is no trace-level insight and the stats endpoint won't scrape. |
| 6 | **Stats are per-instance and in-memory** | Behind a load balancer the "most frequent request" is per-replica and incorrect overall, and all counts are lost on restart. |

### Maintainability

| # | Gap | Impact |
|---|-----|--------|
| 1 | **Brittle tests** — the integration tests assert on exact validator error strings | Breaks across library versions. |
