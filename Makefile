DOCKER_REPOSITORY ?= pimousse1099/fizz-buzz-api-go

# Required runtime config for `make run` (override on the command line if needed).
HTTP_ADDR ?= :8080
LOG_LEVEL ?= info
FIZZBUZZ_MAX_LIMIT ?= 100000

lint:
	@echo "> Launch linter..."
	docker run --rm -v $(PWD):/project -w /project golangci/golangci-lint:v2.12.2 golangci-lint run -v

test:
	@echo "> running tests ..."
	# we can't use alpine as we want to use cgo to check race conditions
	docker run --rm -v $(PWD):/project -w /project golang:1.26 sh -c 'go test --race -v ./... 2>&1'

build-image:
	@echo "> start building docker image..."
	DOCKER_BUILDKIT=1 docker build -t $(DOCKER_REPOSITORY):$(TAG) .

push-image:
	@echo "> start pushing docker image..."
	docker push $(DOCKER_REPOSITORY):$(TAG)

run:
	@echo "> running fizz-buzz-api on $(HTTP_ADDR)"
	docker run --rm -it -v $(PWD):/project -w /project -p8080:8080 \
		-e HTTP_ADDR=$(HTTP_ADDR) -e LOG_LEVEL=$(LOG_LEVEL) -e FIZZBUZZ_MAX_LIMIT=$(FIZZBUZZ_MAX_LIMIT) \
		golang:1.26-alpine go run ./cmd/fizz-buzz-api

.PHONY: lint test build-image push-image run
