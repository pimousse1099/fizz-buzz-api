# 0018. Distributed Tracing — OpenTelemetry

- **Status:** Accepted
- **Date:** 2026-06-24

## Context

Metrics are delegated to infra ([0017](0017-metrics-delegated-to-infra.md)). Distributed tracing
is the one signal the infra layer cannot synthesise — intra-request breakdown (handler vs
use-case vs store vs serialisation) is only visible from inside the application. We need a
tracing strategy that is vendor-neutral and has zero overhead when disabled.

## Decision

Instrument distributed tracing with the **vanilla OpenTelemetry SDK** over **OTLP/HTTP**.

### SDK and exporter

- `go.opentelemetry.io/otel` + `/sdk` + `otlptracehttp` — vanilla OTel SDK, not a vendor distro.
  OTLP works with any backend (Tempo, Jaeger, Datadog, Splunk, …). (reezoback uses the Splunk
  distro pinned to OTel v1.28; this service uses vanilla at v1.44.)

### Server span

- **`github.com/riandyrn/otelchi`** middleware creates the server span, named by chi route
  pattern, added outermost so it covers the whole request. It is a no-op when no provider is
  configured.

### Use-case spans

```go
// tracerName = "github.com/Pimousse1099/fizz-buzz-api/usecase" (instrumentation scope)
ctx, span := otel.Tracer(tracerName).Start(ctx, "usecase.generate_fizzbuzz")
defer span.End()
```

The OTel trace **API** is a cross-cutting instrumentation API that is a no-op without a provider,
so importing it in the use-case layer is legitimate — unlike an HTTP logging library, it carries
no transport-layer semantics.

### Propagation and sampling

- **W3C** propagation (`tracecontext` + `baggage`).
- Sampler: parent-based ratio (configurable via `TRACING_SAMPLE_RATIO`).

### Disabled by default

`TRACING_ENABLED=false` (default) — the global tracer stays the no-op; the app runs with
**no collector dependency and zero overhead**.

**Configuration:**

| Variable | Purpose |
|---|---|
| `TRACING_ENABLED` | Enable/disable (default: `false`) |
| `TRACING_SAMPLE_RATIO` | Sampling ratio (default: `1.0` when enabled) |
| `TRACING_OTLP_ENDPOINT` | Collector endpoint (or standard `OTEL_EXPORTER_OTLP_ENDPOINT`) |

### Lifecycle

The OTel provider is owned by the DI container (`GetTracerProvider`) and **flushed and stopped**
on graceful shutdown (`cmd/main.go`) to ensure in-flight spans are exported before the process
exits.

## Consequences

- Vendor-neutral OTLP avoids backend lock-in; switching from Tempo to Jaeger is a config change.
- Disabled by default means new deployments work without a collector — no operational surprise.
- The use-case layer imports only the OTel API (not the SDK or any exporter), preserving the
  dependency rule: use-cases are not coupled to any infrastructure concern.
- The DI container owns provider lifecycle, keeping `cmd/main.go` as the single shutdown
  orchestrator.
