# 0019. Container and CI/CD

- **Status:** Accepted
- **Date:** 2026-06-24

## Context

The service must be deployable as a container image and verified continuously. We need a
container strategy that produces a minimal, secure, non-root image and a CI/CD pipeline that
enforces quality gates on every commit and publishes images on release.

## Decision

### Container image — multi-stage Dockerfile

**Build stage:** compiles a static, CGO-free binary:
```
-trimpath
-ldflags="-s -w -X github.com/Pimousse1099/fizz-buzz-api/config.AppVersion=${APP_VERSION}"
CGO_ENABLED=0
```

**Runtime stage:** `gcr.io/distroless/static:nonroot`
- No shell, no package manager — minimal attack surface.
- Runs as a non-root user by default (`nonroot`).
- Includes CA certificates for outbound TLS (e.g. OTLP/HTTP to a collector).

**Multi-arch:** images are built for `linux/amd64` and `linux/arm64`.

**No `ENV` defaults and no `EXPOSE` in the Dockerfile** — environment variables are provided
explicitly at runtime (12-factor, see [0009](0009-configuration-and-lifecycle.md)); port
exposure is declared at the orchestrator level, not baked into the image.

### CI — GitHub Actions

Triggered on every push / pull request:

1. **`golangci-lint` v2** — lint and format check (see [0012](0012-linting-golangci-lint.md)).
2. **`go test -race` + coverage** — tests with the race detector enabled.
3. **`go vet`** — static analysis.

### CD — GitHub Actions

Triggered on push to `master` or a tag:

- Build and push multi-arch image to **GitHub Container Registry (GHCR)**.
- Image tagged with the Git SHA and, on tags, the semantic version.

### Dependency updates — Dependabot

Dependabot is configured for:
- `gomod` — Go module dependencies.
- `docker` — base image updates.
- `github-actions` — Actions version pinning.

## Consequences

- The distroless runtime image has no shell, so `docker exec` into a running container is
  impossible — debugging relies on structured logs and traces.
- Static binary + distroless = minimal image size and near-zero OS-level CVE surface.
- No `ENV` defaults enforce the explicit-configuration discipline; a misconfigured deployment
  fails at `go-envconfig` parse time, not silently at first use.
- Multi-arch images support ARM-based nodes (AWS Graviton, Apple Silicon in dev) without
  cross-compilation complexity at deploy time.
- Dependabot keeps the dependency graph fresh with minimal manual effort.
