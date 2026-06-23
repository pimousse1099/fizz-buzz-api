.PHONY: build test race lint fix run tidy

build:
	go build ./...

test:
	go test ./...

race:
	go test -race ./...

lint:
	golangci-lint run

fix:
	golangci-lint run --fix

run:
	go run ./cmd

tidy:
	go mod tidy
