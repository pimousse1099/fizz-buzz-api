# 0008. Operational Endpoints — Health and Readiness

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

Production services running in a container orchestrator (Kubernetes, ECS) need liveness and
readiness probes. We must decide which endpoints to expose and what they return.

## Decision

Expose the following routes:

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/fizzbuzz` | Main fizz-buzz endpoint (see [0002](0002-api-shape-get-query-params.md)) |
| `GET` | `/fizzbuzz/stats` | Statistics sub-resource (see [0002](0002-api-shape-get-query-params.md)) |
| `GET` | `/healthz` | Liveness probe — returns `200 OK` immediately |
| `GET` | `/readyz` | Readiness probe — returns `200 OK` when the server is ready to serve |

`/healthz` (liveness) and `/readyz` (readiness) follow the Kubernetes convention. Both return
`200 OK` with a minimal JSON body (e.g. `{"status":"ok"}`). They are cheap, synchronous checks
that do not touch the stat store or other I/O.

Route path constants are defined next to their handlers (`GenerateFizzBuzzRoute = "/fizzbuzz"`)
and wired via the chi router (see [0014](0014-adopt-chi-router-and-middleware.md)).

## Consequences

- Orchestrators can distinguish "process alive" (liveness) from "ready to receive traffic"
  (readiness), enabling correct rolling-deploy behaviour.
- The endpoints are trivially testable with `httptest`.
- No authentication on health endpoints — they must be reachable before the service is fully
  initialised and must not add latency to probes.
