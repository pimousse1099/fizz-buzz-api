# 0017. Metrics — Delegated to Infrastructure

- **Status:** Accepted
- **Date:** 2026-06-24

## Context

Production services need observability. HTTP golden signals (latency p50/p95/p99, throughput,
error rate by route/status) are the most common starting point. We must decide whether the
application instruments these signals itself or delegates to the platform.

## Decision

**Golden-signal HTTP metrics are delegated entirely to the infrastructure layer.** The application
does not expose a `/metrics` endpoint and contains no application metrics code.

**Infrastructure options (any one suffices):**
- Service mesh sidecar (Envoy / Istio / Linkerd)
- Ingress / API gateway
- Cloud load balancer
- eBPF auto-instrumentation (Grafana Beyla, Pixie)

**Per the 12-factor "logs as event streams" principle:** the application writes only **structured
JSON logs to stdout** (including per-request duration via httplog — see [0015](0015-request-logging-httplog.md))
and leaves shipping, aggregation, and dashboarding to the platform. It does not manage log files,
rotation, or a metrics scrape endpoint.

**Application-level instrumentation is reserved for what infra cannot synthesise:**
- **Distributed tracing** — intra-request breakdown (handler vs use-case vs store vs
  serialisation) is invisible to the infra layer and is instrumented in-app with OpenTelemetry
  (see [0018](0018-distributed-tracing-opentelemetry.md)).
- Future **business metrics** (e.g. most-popular fizz-buzz combos) if product analytics require them.

HTTP performance metrics are explicitly not instrumented in-app.

## Consequences

- Zero metrics code in the service — nothing to maintain, no scrape endpoint to secure or version.
- Consistent observability across all services in the platform regardless of language or framework:
  the infra layer sees every service the same way.
- Golden signals are available from day one if the infra layer is already in place; no deployment
  coordination required.
- If the infra layer is absent (e.g. local development), HTTP metrics are simply not collected —
  this is acceptable; structured logs and traces cover most debugging needs.
