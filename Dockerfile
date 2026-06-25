# syntax=docker/dockerfile:1

# ----------------------------------------------------------------------------
# Build stage: compile a static, stripped, version-stamped binary.
# ----------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION=dev

WORKDIR /src

# Download dependencies first (cached unless go.mod/go.sum change).
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Build.
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath \
      -ldflags="-s -w -X github.com/Pimousse1099/fizz-buzz-api/config.AppVersion=${APP_VERSION}" \
      -o /out/fizz-buzz-api ./cmd

# ----------------------------------------------------------------------------
# Runtime stage: distroless static — no shell/package manager, non-root,
# CA certs + tzdata included (needed for OTLP/TLS export).
# ----------------------------------------------------------------------------
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /out/fizz-buzz-api /fizz-buzz-api

# Liveness/readiness are HTTP probes (GET /healthz, /readyz) driven by the
# orchestrator; distroless has no shell for a Docker HEALTHCHECK.

USER nonroot:nonroot
ENTRYPOINT ["/fizz-buzz-api"]
