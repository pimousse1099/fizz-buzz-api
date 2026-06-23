# ADR — Fizz-Buzz REST API

- **Date:** 2026-06-23
- **Branch:** `pure-go-2026`
- **Module path:** `github.com/Pimousse1099/fizz-buzz-api`
- **Status:** Accepted
- **Context source:** `README.md` (test instructions)

## 1. Context

We must implement a production-ready, maintainable REST API exposing fizz-buzz:

- **Main endpoint** — accepts five parameters: three integers `int1`, `int2`, `limit`, and two
  strings `str1`, `str2`. Returns the list of values from 1 to `limit` where multiples of `int1`
  become `str1`, multiples of `int2` become `str2`, and multiples of both become `str1str2`.
- **Statistics endpoint (bonus)** — accepts no parameter and returns the parameters of the most
  frequently requested fizz-buzz call plus its hit count.

The server must be production-ready (input validation, bounded resource usage, safe concurrency,
correct HTTP status codes, graceful shutdown) and easy for other developers to maintain.

This solution is designed **from the README spec alone** and stands on its own (no dependency on
other implementation branches). It is a **standalone module**, so it cannot reuse the private
internal libraries of the reference monorepo (`backend-libs`, `gol4ng/httpware`, `gorilla/mux`);
instead it reproduces their spirit with the Go standard library plus a minimal set of dependencies.

## 2. Decisions

### 2.1 HTTP stack — stdlib `net/http` + `golang.org/x/time/rate`

**Decision:** Use the standard library `net/http` (Go 1.22+ `ServeMux` with method+path routing) as
the base, with hand-written middlewares. No web framework (gin/echo) and no third-party router
(chi/gorilla). Rate limiting is addressed separately (see §2.11).

**Rationale:**
- With only five simple parameters, hand-written validation is clearer and more maintainable than
  struct-tag magic or framework binding, and it yields precise, fully controlled error messages.
- Middleware composition and `ServeMux` (since Go 1.22) make chi/gin largely redundant here.
- "Production-ready" is better demonstrated by a mastered, zero-magic stdlib stack than by a
  framework that hides the logic. Consistent with the `pure-go` intent.

### 2.2 API shape — `GET` with query parameters

**Decision:** Expose the main endpoint as `GET /fizzbuzz?int1=3&int2=5&limit=100&str1=fizz&str2=buzz`.

**Rationale:** The operation is a pure read; `GET` is RESTful, cacheable, and trivially testable
with curl/browser. Only `GET` is supported (no `POST`) to keep the validation/test surface small.

### 2.3 Response & error format

**Decision:**
- **Success:** raw JSON array of strings — `["1","2","fizz","4","buzz",...]` — with
  `Content-Type: application/json`.
- **Error:** a dedicated HTTP status code plus a valid JSON body. A simple JSON string is
  acceptable, e.g. `"int1 must be a positive integer"`.

**Rationale:** The spec asks for "a list of strings"; a raw array is the most direct representation.
Errors still return valid JSON so clients can always parse the body.

### 2.4 Clean / hexagonal architecture

**Decision:** Layered architecture with the dependency rule `presentation → usecase → domain` and
`infrastructure` wiring everything via a DI container. The **domain never imports the use-case or
any outer layer.**

The business request/response objects (`GenerateRequest`, `GenerateResponse`, `GetStatsRequest`,
`GetStatsResponse`) are **domain/business concepts** and live in the `domain/fizzbuzz` package.
Use-cases consume them. Domain independence is preserved as a *dependency-direction* property: the
names evoke operations, but the `domain` package depends on nothing above it. Type names are kept
free of the `FizzBuzz` prefix to avoid stuttering with the package name (`fizzbuzz.GenerateRequest`,
not `fizzbuzz.GenerateFizzBuzzRequest`).

The **generation logic lives in the use-case** (`Execute`), not on the domain type: the domain holds
data structures and validation only; the use-case applies the business logic (validate → generate →
record). The domain model carries no `Generate()` method.

**Rationale:** Mirrors the in-house convention (reezoback `agencies-api`) for familiarity and
maintainability, while fixing its smell of naming domain sub-packages after use-cases — the domain
must not know which use-cases exist.

#### Package layout

```
domain/fizzbuzz/         # package fizzbuzz — named after the business concept, never a use-case
  model.go     # GenerateRequest{Int1,Int2,Limit,Str1,Str2} (data only, no methods)
               # GenerateResponse{Result []string}
               # GetStatsRequest{} (empty); GetStatsResponse{Request GenerateRequest, Hits int}
  error.go     # sentinel errors: ErrFailedToValidateGenerateRequest, ErrNoStatsRecorded
  validator.go # func (r *GenerateRequest) Validate(maxLimit int) error   (business invariants)

usecase/
  generate_fizzbuzz.go   # defines StatRecorder; Execute(fizzbuzz.GenerateRequest) (fizzbuzz.GenerateResponse, error)
                         #   = req.Validate(maxLimit) -> generate (loop+switch here) -> recorder.Record(req)
  get_fizzbuzz_stats.go  # defines StatReader;  Execute(fizzbuzz.GetStatsRequest) (fizzbuzz.GetStatsResponse, error)

presentation/http/
  server/server.go       # Server{Srv *http.Server, Logger *slog.Logger}: Start(stopChan) / Stop(ctx, cancel)
  server/middleware.go   # recovery, request-id, logging, rate-limit — hand-written; defines RateLimiter iface
  server/routes.go       # ServeMux: GET /fizzbuzz, GET /fizzbuzz/stats, GET /healthz
  handler/generate_fizzbuzz.go   # parse query -> fizzbuzz.GenerateRequest -> uc.Execute -> JSON
  handler/get_fizzbuzz_stats.go  # uc.Execute -> JSON

infrastructure/
  di/container.go        # IoC container, lazy memoized getters (if c.x == nil)
  di/http_server.go      # build *http.Server (timeouts), wire middlewares + routes + handlers
  di/logger.go           # slog (JSON handler)
  stat_storer/in_memory.go  # implements StatRecorder + StatReader (map + sync.Mutex, memoized top)
  stat_storer/redis.go      # struct + methods panic("implement me") — extensibility placeholder
  rate_limiter/in_memory.go # implements RateLimiter (x/time/rate token bucket) — local per-instance guard
  rate_limiter/redis.go     # struct + methods panic("implement me") — distributed-limiter placeholder

config/config.go         # sethvargo/go-envconfig: Addr, MaxLimit, RateLimit, timeouts; AppName/AppVersion
cmd/main.go              # config -> di.NewContainer -> Start -> intercept SIGINT/SIGTERM -> Stop (graceful)
```

### 2.5 Domain owns validation; validation triggered inside the use-case

**Decision:** Validation rules live in the `domain/fizzbuzz` package as a method on the value object:
`func (r GenerateRequest) Validate(maxLimit int) error` (in `validator.go`). The use-case calls
`req.Validate(maxLimit)` at the very start of `Execute()`, so the use-case guarantees its invariants
regardless of the caller. The handler stays thin.

A method (not a `RequestValidator` struct) is chosen because the rules are pure: the only external
input is `maxLimit`, passed as an argument rather than via an injected dependency. If validation
ever needs injected dependencies, this can be refactored into a dedicated validator type.

**Receiver type — value, not pointer.** `Validate` (and any other method on the domain value
objects) uses a **value receiver**. `GenerateRequest` is a small (~56-byte), immutable value object
that is also used as a map key in the stat store. For such a type a value receiver copies a few
bytes on the stack with no allocation, whereas a pointer receiver buys nothing and can force the
struct to escape to the heap (extra allocation + GC pressure). Pointer receivers are reserved for
mutation or large structs — neither applies here — so value receivers are both idiomatic and at
least as fast.

**Validation rules:**
- `int1 > 0` and `int2 > 0` (reject non-positive).
- `1 <= limit <= MaxLimit` (configurable, default 10000) — bounds the response size to prevent
  memory exhaustion / DoS.
- `str1` and `str2` non-empty and length-bounded (e.g. `<= 100`).

### 2.6 Error handling — domain errors mapped to HTTP codes

**Decision:** The domain defines descriptive **sentinel errors** (prefixed `Err`, per the `errname`
linter) that callers wrap with `%w` (per the `err113` linter). The handler classifies them via
`errors.Is` and maps them to HTTP status codes.
- `fizzbuzz.ErrFailedToValidateGenerateRequest` — `Validate` returns it wrapped with a field-level
  detail, e.g. `fmt.Errorf("int1 must be a positive integer: %w", ErrFailedToValidateGenerateRequest)`.
  This and query-parsing failures (non-integer, missing parameter) → **`400 Bad Request`** with the
  message as a JSON string.
- `fizzbuzz.ErrNoStatsRecorded` (no request counted yet) → **`404 Not Found`**.
- Any other/unexpected error → **`500 Internal Server Error`**.

No `ValidationError` struct is used: a wrapped sentinel is simpler and satisfies `err113` (which
discourages dynamic `errors.New`/`fmt.Errorf` without a wrapped sentinel). The field detail travels
in the wrap message; the sentinel carries the classification for `errors.Is`.

**Rationale:** Keeps validity rules in the domain and lets the transport layer translate them into
the correct, semantically meaningful HTTP codes. The HTTP verb/code carries the meaning (404 = no
"most frequent" resource to return).

### 2.7 Statistics — in-memory store behind segregated interfaces

**Decision:**
- The counter is **in-memory**, thread-safe (`map[fizzbuzz.GenerateRequest]int` guarded by a
  `sync.Mutex`), reset on restart.
- It sits behind **segregated interfaces** (interface segregation, defined where consumed):
  - `StatRecorder` (write — `Record(fizzbuzz.GenerateRequest)`), defined in the generate use-case.
  - `StatReader` (read — `MostFrequent() (fizzbuzz.GenerateRequest, int, bool)`), defined in the
    stats use-case.
- The in-memory store implements both. A `redis.go` placeholder (struct + `panic("implement me")`)
  demonstrates how a durable backend would plug in without touching the use-cases.
- **Only successful (`200 OK`) requests are counted** — an invalid (`400`) request is not a real
  fizz-buzz request.
- **Tie-breaking:** on equal hit counts, return the combination that reached the max **first**.
  Implemented in O(1) by memoizing the current top and updating it only when a count *strictly*
  exceeds the current maximum — deterministic and test-friendly.

**Rationale:** In-memory is sufficient for the spec; the interface boundary and redis stub show the
extensibility expected of production code without pulling in real infrastructure.

### 2.8 Routing & operational endpoints

**Decision:**
- `GET /fizzbuzz` — main endpoint.
- `GET /fizzbuzz/stats` — statistics as a sub-resource of fizzbuzz.
- `GET /healthz` — liveness/readiness probe returning `200`.

### 2.9 Configuration & lifecycle

**Decision:**
- Configuration loaded from environment variables via `sethvargo/go-envconfig` (no stdlib
  equivalent for struct↔env mapping). Config covers: listen address, `MaxLimit`, rate-limit
  parameters, and HTTP server timeouts. `AppName`/`AppVersion` as constants/build-time vars.
- `cmd/main.go`: load config → build DI container → `httpSrv.Start(stopChan)` → intercept
  `SIGINT`/`SIGTERM` → `Stop(ctx, cancel)` for graceful shutdown.
- `http.Server` configured with `ReadHeaderTimeout`, `WriteTimeout`, `IdleTimeout`,
  `MaxHeaderBytes`.

### 2.10 Logging — stdlib `log/slog`

**Decision:** Use the standard library `log/slog` (JSON handler) rather than `logrus`.

**Rationale:** Structured, zero-dependency, modern (Go 1.21+), consistent with the `pure-go` intent.

> **To refine.** Logging, observability (metrics/tracing), and the precise middleware stack
> (recovery, request-id, logging, rate-limit ordering) are intentionally left open at this stage and
> will be decided in a follow-up discussion. This section will be updated then.

### 2.11 Rate limiting — infra/edge is authoritative; app keeps a local guard

**Decision:**
- For a genuinely production-ready, horizontally scalable app, the **authoritative** rate limit
  belongs at the **infrastructure / edge layer** (API gateway, ingress, reverse-proxy, cloud load
  balancer) or, failing that, on a **shared store** (e.g. Redis sliding-window / token-bucket). This
  is the recommended production approach.
- The application ships only a **local, per-instance guard** to keep a single instance from being
  saturated. It is implemented with `golang.org/x/time/rate` (token bucket) wrapped in a
  `func(http.Handler) http.Handler` middleware that returns **`429 Too Many Requests`** when the
  bucket is empty.
- The limiter sits behind a `RateLimiter` interface (defined in `server/middleware.go`, where it is
  consumed), symmetric with the stat store: an in-memory implementation now, plus a `redis.go`
  placeholder (`panic("implement me")`) showing the path to a distributed limiter.

**Known limitation (explicit):** an in-process limiter is **per instance**. With N replicas behind a
load balancer the effective global limit becomes N × the configured rate, and a single client spread
across instances is limited inconsistently. The in-process limiter is therefore a coarse safety net,
**not** a correct global limit — that responsibility is delegated to the edge or a shared store.

**Note (`x/time/rate` vs `go.uber.org/ratelimit`):** `x/time/rate` is a *token bucket* with
non-blocking `Allow()` semantics — ideal to **reject** surplus traffic (429), which is what an
ingress guard wants. `go.uber.org/ratelimit` is a *leaky bucket* whose `Take()` **blocks** to
**smooth** an outbound throughput to a steady rate. Different goals: protect-and-reject vs
shape-and-smooth. We adopt `x/time/rate` for inbound protection.

### 2.12 Linting & formatting — golangci-lint (v2), reezoback intent

**Decision:** Lint with **golangci-lint** using a `.golangci.yaml` that adopts the reezoback house
philosophy — `disable-all: true` plus an explicit curated enable list (the scalable, future-proof
approach) — but authored for the **v2 config schema** (current with Go 1.26), not a verbatim copy of
the reezoback v1 file.

Carried over from reezoback: `gofumpt` (extra-rules), `gci` (sections: standard / default /
`prefix(github.com/Pimousse1099)`), complexity bounds (`cyclop`, `gocognit`, `funlen`, `nestif`),
`errname` + `err113` (sentinel errors, wrapping), `gosec`, `gochecknoglobals`, `gochecknoinits`,
`testpackage` (black-box `_test` packages), `paralleltest` + `tparallel`, `revive`, `staticcheck`,
`gocritic`, `godot`, `misspell`, `wsl`/`nlreturn`/`whitespace`, `mnd` (magic numbers), and notably
**`sloglint`** (consistent `log/slog` usage — matches §2.10).

Adapted for v2 / this service:
- Renames: `gomnd` → `mnd`, `goerr113` → `err113`.
- Dropped (removed/obsolete): `execinquery`, `exportloopref` (loop-var capture fixed in Go 1.22).
- Dropped (irrelevant to a stdlib HTTP service): `protogetter`, `rowserrcheck`, `sqlclosecheck`,
  `zerologlint`, and the reezoback-specific `skip-dirs` (graphql/gateway).
- `gci` / `goimports` local prefix points to the module path `github.com/Pimousse1099`.

The concrete `.golangci.yaml` is written during implementation; parity with reezoback can be
reviewed afterward.

### 2.13 Context propagation

**Decision:** `context.Context` is threaded from the HTTP handler down through the use-case to every
**I/O boundary**. The handler passes `r.Context()`; the affected signatures take `ctx` as their
first parameter:
- `usecase.GenerateFizzBuzz.Execute(ctx, req)` and `usecase.GetFizzBuzzStats.Execute(ctx, req)`.
- `StatRecorder.Record(ctx, req)` and `StatReader.MostFrequent(ctx)`.
- `RateLimiter.Allow(ctx)`.

**Pure domain compute stays ctx-free:** `GenerateRequest.Validate(maxLimit)` (and the in-use-case
generation loop) perform no I/O and gain nothing from a context, so they keep no `ctx` parameter —
adding context to pure functions is non-idiomatic.

**Rationale:** Production-readiness — cancellation, deadlines and trace propagation must flow with
the request. This matters most at the seams the Redis placeholders represent: a durable stat store
or a distributed rate limiter makes network calls that must honor the request's context. In-memory
implementations accept `ctx` to satisfy the interface contract even though they ignore it, so the
network-bound implementations drop in without signature changes.

## 3. Dependencies

**Runtime:**
- `golang.org/x/time/rate` — rate limiting (local per-instance guard, see §2.11).
- `github.com/sethvargo/go-envconfig` — env-var configuration.
- Everything else from the Go standard library (`net/http`, `log/slog`, `sync`, `encoding/json`).

**Dev tooling:**
- `golangci-lint` (v2) — linting/formatting (see §2.12).

## 4. Testing strategy

- Pure unit tests on `domain/fizzbuzz` (generation incl. edge cases: `int1==int2`, `str1str2`
  combination, limit boundaries; validation rules and wrapped sentinel errors via `errors.Is`).
- Unit tests on use-cases with a fake store.
- Concurrency test on the in-memory store, run under `go test -race` (critical for the counter).
- Handler/integration tests via `httptest` covering status codes (200/400/404), response shapes,
  and the stats counting/tie-breaking semantics.
- Black-box `_test` packages (per the `testpackage` linter) and `t.Parallel()` where applicable.
- Verification gate before "done": `go build ./...`, `go test -race ./...`, `golangci-lint run`.

## 5. Consequences

- Zero web framework: small, transparent, easy to reason about and maintain; slightly more
  hand-written plumbing (middlewares, query parsing).
- Stats reset on restart (acceptable per spec); the `StatReader`/`StatRecorder` seam allows a
  durable backend later with no use-case changes.
- The clean-architecture layering adds files but keeps each unit small, single-purpose, and
  independently testable.
```
