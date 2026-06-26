FROM golang:1.26-alpine AS builder
WORKDIR /project
RUN apk update && apk add --no-cache ca-certificates && update-ca-certificates
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o fizz-buzz-api ./main.go

FROM scratch
COPY --from=builder /project/fizz-buzz-api .
ENTRYPOINT ["/fizz-buzz-api"]
