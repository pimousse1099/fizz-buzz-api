# fizz-buzz-api

A small, production-ready REST API that generates fizz-buzz sequences and reports
the most-requested call. Built in Go with a clean/hexagonal architecture.

> **The task.** Expose a REST endpoint that accepts five parameters — three
> integers `int1`, `int2`, `limit` and two strings `str1`, `str2` — and returns
> the numbers from 1 to `limit` where multiples of `int1` become `str1`,
> multiples of `int2` become `str2`, and multiples of both become `str1str2`.
> Bonus: a statistics endpoint (no parameters) returning the parameters of the
> most frequent request and its hit count. The server must be production-ready
> and easy to maintain.

For the architecture and the rationale behind every choice, see the
[developer guide](docs/developer-guide.md) and the
[Architecture Decision Records](docs/architecture-decision-records/README.md).

## Endpoints

| Method & path | Description |
|---|---|
| `GET /fizzbuzz` | Generate the sequence (see params below) |
| `GET /fizzbuzz/stats` | Most frequent successful request + its hit count (`404` if none yet) |
| `GET /healthz` | Liveness probe (always `200` while the process is up) |
| `GET /readyz` | Readiness probe (`200` when ready to serve) |

### `GET /fizzbuzz`

All query parameters are required:

| Param | Type | Rule |
|---|---|---|
| `int1` | int | `> 0` |
| `int2` | int | `> 0` |
| `limit` | int | `1 .. FIZZBUZZ_MAX_SEQUENCE_LENGTH` |
| `str1` | string | non-empty, ≤ 100 chars |
| `str2` | string | non-empty, ≤ 100 chars |

```sh
curl 'http://localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz'
# ["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]
```

Success returns a raw JSON array of strings. Invalid/missing parameters return
`400` with a JSON string message; an unexpected failure returns `500`.

### `GET /fizzbuzz/stats`

```sh
curl 'http://localhost:8080/fizzbuzz/stats'
# {"request":{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"buzz"},"total_hits":1}
```

Only successful (`200`) generations are counted. On a tie, the request that
reached the maximum first is reported. Returns `404` until at least one
successful generation has happened.

## Limits

- **Input validation** — the rules in the table above; `limit` is bounded by
  `FIZZBUZZ_MAX_SEQUENCE_LENGTH` to cap response size (DoS guard).
- **Rate limiting** — per client IP, `HTTP_RATE_LIMIT_REQUESTS` per
  `HTTP_RATE_LIMIT_WINDOW`; over the limit returns `429` with `Retry-After` and
  `X-RateLimit-*` headers. The client IP comes from `X-Forwarded-For` /
  `X-Real-IP` / `True-Client-IP`, which are spoofable — so this is only sound
  **behind a trusted proxy** that sets and overwrites those headers; without one,
  use socket-based limiting. (In a scaled deployment the authoritative limit
  belongs at the edge / a shared store — see
  [ADR 0016](docs/architecture-decision-records/0016-rate-limiting-httprate.md).)
- **Response compression** — gzip/deflate for compressible types (incl. JSON),
  negotiated via `Accept-Encoding`.
- **Timeouts** — read-header / write / idle timeouts on the HTTP server.

## Configuration

All configuration comes from environment variables (12-factor). Required
variables have **no default on purpose** — they must be provided explicitly.

| Variable | Required | Default | Description |
|---|---|---|---|
| `ENV_TYPE` | yes | — | Deployment environment (e.g. `production`, `dev`) |
| `ENV_NAME` | no | — | Free-form instance label |
| `HTTP_ADDR` | yes | — | Listen address, e.g. `:8080` |
| `HTTP_READ_HEADER_TIMEOUT` | no | `2s` | Read-header timeout |
| `HTTP_WRITE_TIMEOUT` | no | `10s` | Write timeout |
| `HTTP_IDLE_TIMEOUT` | no | `120s` | Idle timeout |
| `HTTP_RATE_LIMIT_REQUESTS` | no | `100` | Per-IP request allowance per window |
| `HTTP_RATE_LIMIT_WINDOW` | no | `1m` | Rate-limit window length |
| `FIZZBUZZ_MAX_SEQUENCE_LENGTH` | yes | — | Upper bound for the `limit` parameter |
| `LOG_LEVEL` | yes | — | `debug` / `info` / `warn` / `error` |
| `TRACING_ENABLED` | no | `false` | Enable OpenTelemetry tracing (OTLP/HTTP) |
| `TRACING_SAMPLE_RATIO` | no | `1` | Trace sampling ratio (0..1) |
| `TRACING_OTLP_ENDPOINT` | no | — | OTLP/HTTP collector `host:port` (else `OTEL_EXPORTER_OTLP_ENDPOINT`) |

## Running

Requires Go 1.26+.

```sh
# From source — required env vars must be provided
ENV_TYPE=dev HTTP_ADDR=:8080 FIZZBUZZ_MAX_SEQUENCE_LENGTH=10000 LOG_LEVEL=info \
  go run ./cmd
```

With Docker (the image sets no defaults — pass the env explicitly):

```sh
docker build -t fizz-buzz-api .
docker run --rm -p 8080:8080 \
  -e ENV_TYPE=dev -e HTTP_ADDR=:8080 \
  -e FIZZBUZZ_MAX_SEQUENCE_LENGTH=10000 -e LOG_LEVEL=info \
  fizz-buzz-api
```

## Production-readiness

Details and rationale in the [developer guide](docs/developer-guide.md) and the
[ADRs](docs/architecture-decision-records/README.md).

### The twelve factors

| # | Factor | How it is applied here |
|---|---|---|
| 1 | Codebase | One codebase in Git, many deploys |
| 2 | Dependencies | Declared/locked in `go.mod`/`go.sum`; CGO-free static binary — no system deps |
| 3 | Config | All config from environment variables (`go-envconfig`); required vars have no default |
| 4 | Backing services | The stat store and rate-limit counter sit behind interfaces; a Redis backend attaches via config with no code change (in-memory by default) |
| 5 | Build, release, run | Build = the immutable distroless image (CI); run = that same image configured purely by env — stages are separate |
| 6 | Processes | Stateless **except** the in-memory stats counter, which is **per-instance and reset on restart** — a deliberate trade-off; a shared store (Redis) is the documented path to true statelessness |
| 7 | Port binding | Self-contained binary binds `HTTP_ADDR`; no external web server |
| 8 | Concurrency | Scales out as identical processes; note the per-instance state from factor 6 — stats and the in-memory rate limit are **not shared across replicas**, so authoritative limits/stats belong at the edge or a shared store |
| 9 | Disposability | Fast startup; graceful shutdown on `SIGINT`/`SIGTERM` with a bounded deadline |
| 10 | Dev/prod parity | The same binary/image everywhere; behaviour driven only by env |
| 11 | Logs | Structured JSON to stdout as an event stream; no log files or rotation in-app |
| 12 | Admin processes | None required; one-off tasks would run as separate invocations of the same binary |

### Design principles

- **Clean / hexagonal architecture** with the dependency rule
  `presentation → usecase → domain`; the domain depends on nothing.
- **SOLID** — single-responsibility packages; interface-**segregated** ports
  (`StatRecorder` / `StatReader`); **dependency inversion** via the DI container.
- **Law of Demeter** — components talk only to their immediate collaborators
  (handlers → use-cases → ports; a handler never reaches through to the store),
  **and a function receives only what it needs**, never a god-object: e.g.
  `Validate(maxLimit int)` takes the single bound it checks — not the whole
  `Config` — and the rate-limit middleware is handed a `RateLimiter`, not the
  container. (You hand the baker the coins, not your whole wallet.)

### Observability & safety

- Separate liveness (`/healthz`) and readiness (`/readyz`) probes.
- Structured request/application logs with a request id; opt-in OpenTelemetry
  **tracing**. HTTP **golden-signal metrics** are **delegated to the
  infrastructure layer** (mesh / ingress / LB / eBPF), not instrumented in-app —
  see [ADR 0017](docs/architecture-decision-records/0017-metrics-delegated-to-infra.md).
- Input validation, bounded response size, per-IP rate limiting, the race
  detector in CI, and a strict linter.
