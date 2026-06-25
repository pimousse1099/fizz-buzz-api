# Architecture Decision Records

An Architecture Decision Record (ADR) captures a significant architectural choice: the context
that made it necessary, what was decided, and what the consequences are. Each file records one
immutable decision. Decisions are never edited retroactively — when a decision is reversed or
superseded, a new ADR is created and the original is marked **Superseded by NNNN**.

## Index

| # | Decision | Status |
|---|---|---|
| [0001](0001-clean-hexagonal-architecture.md) | Clean / Hexagonal Architecture | Accepted |
| [0002](0002-api-shape-get-query-params.md) | API Shape — GET with Query Parameters | Accepted |
| [0003](0003-response-and-error-format.md) | Response and Error Format | Accepted |
| [0004](0004-domain-model-validation-generation.md) | Domain Model, Validation, and Generation Placement | Accepted |
| [0005](0005-error-to-http-status-mapping.md) | Error-to-HTTP-Status Mapping | Accepted |
| [0006](0006-statistics-store.md) | Statistics Store | Accepted |
| [0007](0007-http-stack-stdlib.md) | HTTP Stack — stdlib `net/http` with Hand-Rolled Middleware | Superseded by [0014](0014-adopt-chi-router-and-middleware.md) |
| [0008](0008-operational-endpoints-health-readiness.md) | Operational Endpoints — Health and Readiness | Accepted |
| [0009](0009-configuration-and-lifecycle.md) | Configuration and Lifecycle | Accepted |
| [0010](0010-structured-logging-slog.md) | Structured Logging — stdlib `log/slog` | Superseded by [0015](0015-request-logging-httplog.md) |
| [0011](0011-rate-limiting-inhouse.md) | Rate Limiting — In-House `x/time/rate` Guard | Superseded by [0016](0016-rate-limiting-httprate.md) |
| [0012](0012-linting-golangci-lint.md) | Linting and Formatting — golangci-lint v2 | Accepted |
| [0013](0013-context-propagation.md) | Context Propagation | Accepted |
| [0014](0014-adopt-chi-router-and-middleware.md) | Adopt chi Router and Middleware | Accepted — Supersedes [0007](0007-http-stack-stdlib.md) |
| [0015](0015-request-logging-httplog.md) | Request Logging — `go-chi/httplog/v2` | Accepted — Supersedes [0010](0010-structured-logging-slog.md) |
| [0016](0016-rate-limiting-httprate.md) | Rate Limiting — `go-chi/httprate` | Accepted — Supersedes [0011](0011-rate-limiting-inhouse.md) |
| [0017](0017-metrics-delegated-to-infra.md) | Metrics — Delegated to Infrastructure | Accepted |
| [0018](0018-distributed-tracing-opentelemetry.md) | Distributed Tracing — OpenTelemetry | Accepted |
| [0019](0019-container-and-ci-cd.md) | Container and CI/CD | Accepted |
