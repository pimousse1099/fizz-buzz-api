FROM golang:1.26-alpine AS builder
WORKDIR /src

# Cache module downloads independently of source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# CGO disabled -> fully static binary, runnable on a distroless/scratch base.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /fizz-buzz-api ./cmd/fizz-buzz-api

# distroless static + nonroot: no shell/package manager (small attack surface),
# but unlike scratch it ships ca-certificates, tzdata and a nonroot user (uid 65532).
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /fizz-buzz-api /fizz-buzz-api
ENTRYPOINT ["/fizz-buzz-api"]
