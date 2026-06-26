# fizz-buzz-api

straight-forward implementation of a fizz-buzz REST API.

It uses the [echo v5](https://echo.labstack.com/) web framework so we don't have to rewrite request
binding, validation or HTTP middlewares by hand. Logging is structured JSON via the standard
library [`log/slog`](https://pkg.go.dev/log/slog) (echo v5 logs through slog natively).

It only allows users to make JSON `POST` requests.

The server listens on a **fixed** `:8080` (see limitations) and wires the following middlewares:
`Recover`, `RequestID`, a slog-backed request logger, `Secure` (security headers), `CORS`, `Gzip`,
`BodyLimit` (1 MiB), a per-IP `RateLimiter` (20 req/s) and a per-request `ContextTimeout` (10s).
Shutdown is graceful: `SIGINT`/`SIGTERM` drains in-flight requests (10s budget) before exiting.

## usage

### prod

`docker run --rm -p 8080:8080 rmasclef/fizz-buzz-api-go:v0.1.0`

### dev
`make run`
or
`go run .`

> Note: the listen port is hardcoded to `:8080`; the previously advertised `HTTP_PORT` argument is
> ignored (see limitations).

## example

## CURL 
```
curl --location --request POST 'localhost:8080/fizz-buzz' \
--header 'Content-Type: application/json' \
--data-raw '{
	"int1": 3,
	"int2": 4,
	"limit": 20,
	"str1": "fizz",
	"str2": "buzz"
}
'
```

## Go
```
package main

import (
  "fmt"
  "strings"
  "net/http"
  "io/ioutil"
)

func main() {

  url := "localhost:8080/fizz-buzz"
  method := "POST"
  payload := strings.NewReader(`{"int1": 3, "int2": 4, "limit": 20, "str1": "fizz", "str2": "buzz"}`)

  req, err := http.NewRequest(method, url, payload)
  if err != nil {
    fmt.Println(err)
  }
  req.Header.Add("Content-Type", "application/json")

  client := &http.Client {}
  res, err := client.Do(req)
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body)

  fmt.Println(string(body))
}
```

the two above code snippets will return the following response body:

`["1","2","fizz","buzz","5","fizz","7","buzz","fizz","10","11","fizzbuzz","13","14","fizz","buzz","17","fizz","19","buzz"]`

## metrics and logs

domain request counters are exposed on `GET /metrics` (sorted by hit count). Note this is **not**
Prometheus exposition format despite the conventional path name — see the limitations below.

logs are structured JSON (`log/slog`) written to stdout, one line per request with `method`, `uri`,
`status`, `latency` and a correlation `request_id` (also returned as the `X-Request-ID` header). We
suggest a sidecar (fluentd, filebeats, ...) to aggregate them in a log service (graylog, logstash, ...).

## limitations (this version is NOT production-ready)

This branch is the **simple / naive** implementation. Some hardening is now in place (request
timeout, body-size cap, per-IP rate limiting, security headers, structured logging with correlation
ids and graceful shutdown — see the intro), but it still should not be deployed as-is. The gaps
below are intentional and documented so the trade-offs are explicit; a fully hardened, layered
design lives on the `clean-archi-2026` branch.

### Production-readiness

| # | Gap | Impact |
|---|-----|--------|
| 1 | **`limit` is unbounded** — `make(fizzBuzzResponse, limit)` with `limit` validated only as `required` (non-zero `uint`) | A single `limit=4000000000` request allocates billions of strings → OOM / denial of service. No upper bound, pagination or streaming. (Mitigated, not solved, by the 1 MiB body cap on the *request*.) |
| 2 | **Data races in the stats collector** — `IncRequestCounter` ranges over the slice outside the mutex; `/metrics` calls `sort.Sort` (in-place) with no lock | Under real concurrency: corrupted counts or crash. The current tests don't exercise concurrency, so `-race` stays green — a false sense of safety. |
| 3 | **Unbounded stats growth** — every distinct parameter tuple appends an entry that is never evicted, looked up via O(n) linear scan | High-cardinality / malicious clients leak memory and degrade latency over time. |
| 4 | **Hardcoded configuration** — the listen address is hardcoded `:8080`; `HTTP_PORT` (previously advertised in the Makefile/README) is ignored by `main` | Cannot change ports or run multiple instances without recompiling. Violates 12-factor config. |
| 5 | **No health / readiness probes** | Orchestrators (Kubernetes, ECS) cannot tell whether the instance is alive or ready to serve. |
| 6 | **Partial observability** — structured slog logs with `request_id` are now emitted, but there is still no distributed tracing, and `/metrics` is custom JSON, not Prometheus exposition format | Per-request latency/errors are visible in logs, but there is no trace-level insight and the `/metrics` path won't scrape despite its name. |
| 7 | **Stats are per-instance, in-memory, lost on restart** | Behind a load balancer the "most frequent request" is wrong, and all counts vanish on redeploy. |

### Maintainability

| # | Gap | Impact |
|---|-----|--------|
| 1 | **No layering** — HTTP wiring, validation, domain logic and the stats store all live in `package main` across flat files | Hard to unit-test in isolation, hard to swap implementations (e.g. a Redis-backed store), dependencies are implicit. |
| 2 | **Stringly-typed stats key** — stats are keyed by `fizzBuzzRequest.String()` (`"int1:3_int2:4_..."`) | Fragile (a format tweak silently breaks aggregation); the endpoint returns this opaque string instead of the structured parameters the spec asks for. |
| 3 | **`/metrics` naming collision** — conventionally Prometheus, here domain counters | Confusing for operators and scraping tooling. |
| 4 | **Brittle tests** — assert on exact echo/validator internal error strings | Broke on this very dependency bump (content-type charset) and will break again on any upgrade; no unit tests for the stats collector or its concurrency. |
| 5 | **Spec only partially met** — the bonus statistics endpoint should return the params of the *single* most-frequent call plus its hit count; `/metrics` instead dumps every entry keyed by an opaque string | Diverges from the requested contract. |

See the `clean-archi-2026` branch for a hexagonal architecture that addresses these:
bounded inputs, mutex-safe and swappable stats store, env-based config, `/healthz` + `/readyz`,
structured logging, OpenTelemetry tracing, rate limiting, and graceful shutdown.
