# 0014. Adopt chi Router and Middleware

- **Status:** Accepted
- **Date:** 2026-06-24
- **Supersedes:** [0007](0007-http-stack-stdlib.md)

## Context

After the initial implementation ([0007](0007-http-stack-stdlib.md)) was reviewed, the hand-rolled
transport layer showed friction points: the `Recovery` middleware did not correctly re-panic
`http.ErrAbortHandler` and risked double-writing a response; the `Chain` helper, custom
`RequestID`, and manual route wiring were meaningful amounts of edge-case-prone plumbing that chi
provides for free. The chi ecosystem is lightweight, `net/http`-native, and widely used in
production Go services.

## Decision

Replace the stdlib-only transport layer with the chi ecosystem:

**Adopted:**
- **`github.com/go-chi/chi/v5`** as the router (`chi.NewRouter`, `r.Get(path, handler)`), replacing
  the stdlib `ServeMux`, the `Chain` helper, and the `NewRouter`/`Route` wiring.
- **`chi/middleware.RequestID`** — request-ID injection, replacing the hand-rolled equivalent.
- **`chi/middleware.Recoverer`** — panic recovery that correctly re-panics `http.ErrAbortHandler`
  and avoids double-writes, fixing the known bug in the hand-rolled `Recovery`.

**Removed packages:**
- `presentation/http/middleware/{chain,recovery,request_id,logging}.go`
- `presentation/http/server/routes.go` (`Route`/`NewRouter`)
- `presentation/http/reqctx` (entire package)

**Kept in-house:** `httpserver/server.go` — owns the `http.Server` lifecycle (Start/Stop,
timeouts, BaseContext, ErrorLog). chi is a router only and does not manage the server.

**Route constants:** business route patterns are now path-only constants defined next to their
handlers (`GenerateFizzBuzzRoute = "/fizzbuzz"`) and wired via `r.Get`.

**Dependencies added:** `github.com/go-chi/chi/v5`.
**Dependencies removed:** hand-rolled middleware code (no net change in external deps for the
router itself).

The structured request logging and rate limiting decisions that accompanied this change are
recorded separately in [0015](0015-request-logging-httplog.md) and [0016](0016-rate-limiting-httprate.md).

## Consequences

- The `Recovery` edge case (`http.ErrAbortHandler`) is fixed by a battle-tested implementation.
- The presentation layer is simpler: the `middleware`, `reqctx`, and hand-rolled route-wiring
  packages are gone.
- chi is `net/http`-native — handlers remain plain `http.HandlerFunc`, no framework lock-in on
  the handler side.
- The domain, use-cases, and stat store were entirely unaffected by this change.
- `golang.org/x/time/rate` was removed as a direct dependency (replaced by httprate, see [0016](0016-rate-limiting-httprate.md)).
