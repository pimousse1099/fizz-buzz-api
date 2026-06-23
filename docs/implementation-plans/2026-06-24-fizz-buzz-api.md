# Fizz-Buzz API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a production-ready fizz-buzz REST API in Go with a generation endpoint and a statistics endpoint, following clean/hexagonal architecture.

**Architecture:** Layered with dependency rule `presentation → usecase → domain`, wired by an `infrastructure/di` container. The domain (`domain/fizzbuzz`) holds the business value objects, generation logic, validation, and sentinel errors. Use-cases orchestrate (validate → generate → record stats). Transport is stdlib `net/http` with hand-written middleware. Stats and rate-limiting live behind segregated interfaces with in-memory implementations and Redis placeholders.

**Tech Stack:** Go 1.26, stdlib `net/http` (ServeMux), `log/slog`, `golang.org/x/time/rate`, `github.com/sethvargo/go-envconfig`, `golangci-lint` v2.

## Global Constraints

- **Module path:** `github.com/Pimousse1099/fizz-buzz-api` — every import uses this prefix.
- **Go version:** 1.26 (`go 1.26` in `go.mod`).
- **Runtime dependencies (only):** `golang.org/x/time/rate`, `github.com/sethvargo/go-envconfig`. Everything else from the standard library.
- **Errors:** package-level **sentinels** prefixed `Err`, wrapped with `%w` (satisfies `errname` + `err113`). Classify with `errors.Is`. No `==` error comparisons, no inline dynamic `errors.New` returned for classification.
- **Validation rules:** `int1 > 0`, `int2 > 0`, `1 <= limit <= MaxLimit` (default 10000), `str1`/`str2` non-empty and `<= 100` chars.
- **Stats semantics:** count **only successful (200)** generations; on tie return the combo that reached the max **first** (update top only on a *strictly greater* count).
- **HTTP responses:** success of `/fizzbuzz` is a raw JSON array of strings; errors carry a dedicated status code + a JSON string body. Stats success is a JSON object; stats-when-empty is `404`.
- **Package names:** no underscores (lint `staticcheck`/`revive`). Directories `statstorer/` and `ratelimiter/` (the ADR's `stat_storer`/`rate_limiter` sketch is renamed here for Go package-naming compliance).
- **Lined/formatted via** `golangci-lint run` (v2 schema). Run `golangci-lint run --fix` to auto-resolve formatter findings (gci/gofumpt) and `wsl`/`whitespace` before committing.
- **Verification gate before any "done" claim:** `go build ./...`, `go test -race ./...`, `golangci-lint run` all pass.

---

## File Structure

```
go.mod
.golangci.yaml
Makefile
config/config.go                          # env-var config (go-envconfig)
domain/fizzbuzz/model.go                   # GenerateRequest/Response, GetStatsRequest/Response, Generate()
domain/fizzbuzz/error.go                   # sentinel errors
domain/fizzbuzz/validator.go              # GenerateRequest.Validate(maxLimit)
usecase/generate_fizzbuzz.go               # StatRecorder iface + GenerateFizzBuzz UC
usecase/get_fizzbuzz_stats.go              # StatReader iface + GetFizzBuzzStats UC
infrastructure/statstorer/in_memory.go     # in-memory StatRecorder + StatReader
infrastructure/statstorer/redis.go         # Redis placeholder (panic("implement me"))
infrastructure/ratelimiter/in_memory.go    # in-memory rate limiter (x/time/rate)
infrastructure/ratelimiter/redis.go        # Redis placeholder
presentation/http/handler/request.go       # query parsing -> fizzbuzz.GenerateRequest
presentation/http/handler/response.go      # writeJSON / writeError helpers
presentation/http/handler/generate_fizzbuzz.go
presentation/http/handler/get_fizzbuzz_stats.go
presentation/http/server/middleware.go     # RateLimiter iface + Recovery/RequestID/Logging/RateLimit/Chain
presentation/http/server/routes.go         # ServeMux wiring
presentation/http/server/server.go         # Server{Start,Stop}
infrastructure/di/logger.go                 # slog JSON logger
infrastructure/di/container.go              # IoC container, lazy getters
infrastructure/di/http_server.go            # build *http.Server + wire middleware/routes/handlers
cmd/main.go                                 # bootstrap + graceful shutdown
README.md                                   # run/eval instructions (append)
```

---

### Task 1: Module bootstrap & tooling

**Files:**
- Create: `go.mod`
- Create: `.golangci.yaml`
- Create: `Makefile`

**Interfaces:**
- Consumes: nothing.
- Produces: a buildable module at `github.com/Pimousse1099/fizz-buzz-api`; `make lint/test/build` targets.

- [ ] **Step 1: Initialize the module**

Run:
```bash
go mod init github.com/Pimousse1099/fizz-buzz-api
go mod edit -go=1.26
```

- [ ] **Step 2: Add the two runtime dependencies**

Run:
```bash
go get golang.org/x/time/rate@latest
go get github.com/sethvargo/go-envconfig@latest
```
Expected: `go.mod` lists both under `require`.

- [ ] **Step 3: Create `.golangci.yaml` (v2 schema, reezoback intent)**

```yaml
version: "2"

run:
  timeout: 5m

linters:
  default: none
  enable:
    - bodyclose
    - contextcheck
    - cyclop
    - dogsled
    - dupl
    - durationcheck
    - errcheck
    - errname
    - errorlint
    - err113
    - exhaustive
    - forbidigo
    - forcetypeassert
    - funlen
    - gochecknoglobals
    - gochecknoinits
    - gocognit
    - goconst
    - gocritic
    - godot
    - goprintffuncname
    - gosec
    - govet
    - importas
    - inamedparam
    - ineffassign
    - lll
    - makezero
    - mirror
    - misspell
    - mnd
    - nakedret
    - nestif
    - nilerr
    - nilnil
    - nlreturn
    - noctx
    - nolintlint
    - paralleltest
    - prealloc
    - predeclared
    - reassign
    - revive
    - sloglint
    - staticcheck
    - testableexamples
    - testpackage
    - thelper
    - tparallel
    - unconvert
    - unparam
    - unused
    - usestdlibvars
    - usetesting
    - wastedassign
    - whitespace
    - wsl
  settings:
    cyclop:
      max-complexity: 15
    dupl:
      threshold: 100
    funlen:
      lines: 100
      statements: 50
    goconst:
      min-len: 2
      min-occurrences: 2
    gocritic:
      enabled-tags:
        - diagnostic
        - opinionated
        - performance
        - style
    lll:
      line-length: 150
    misspell:
      locale: US
  exclusions:
    rules:
      - linters: [funlen, dupl, err113, gochecknoglobals]
        path: _test\.go

formatters:
  enable:
    - gci
    - gofumpt
  settings:
    gci:
      sections:
        - standard
        - default
        - prefix(github.com/Pimousse1099)
    gofumpt:
      extra-rules: true
```

> Notes vs reezoback v1: `gosimple`+`stylecheck` are folded into `staticcheck` in v2; `gomnd`→`mnd`, `goerr113`→`err113`, `tenv`→`usetesting`; `execinquery`/`exportloopref` removed; SQL/proto/prometheus/zerolog linters dropped as irrelevant. Formatters moved under `formatters:`.

- [ ] **Step 4: Create `Makefile`**

```makefile
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
```

- [ ] **Step 5: Verify it builds and commit**

Run:
```bash
go build ./...
```
Expected: no output, exit 0 (no packages yet is fine).

```bash
git add go.mod go.sum .golangci.yaml Makefile
git commit -m "chore: bootstrap go module, linter config and Makefile"
```

---

### Task 2: Domain models

> **Amendment (post-review):** generation logic moved OUT of the domain into the use-case (Task 4),
> per ADR §2.4. `model.go` now holds **data structs only** (no `Generate()`/`value()` methods, no
> `strconv` import). The generation behaviour is TDD-tested in Task 4. Domain methods use **value
> receivers** (ADR §2.5).

**Files:**
- Create: `domain/fizzbuzz/model.go`

**Interfaces:**
- Consumes: nothing.
- Produces:
  - `type GenerateRequest struct { Int1, Int2, Limit int; Str1, Str2 string }`
  - `type GenerateResponse struct { Result []string }`
  - `type GetStatsRequest struct{}`
  - `type GetStatsResponse struct { Request GenerateRequest; Hits int }`

- [ ] **Step 1: Write the failing test**

Create `domain/fizzbuzz/model_test.go`:
```go
package fizzbuzz_test

import (
	"reflect"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

func TestGenerateRequest_Generate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  fizzbuzz.GenerateRequest
		want []string
	}{
		{
			name: "classic fizzbuzz up to 15",
			req:  fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 15, Str1: "fizz", Str2: "buzz"},
			want: []string{"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz"},
		},
		{
			name: "concatenation order is str1 then str2",
			req:  fizzbuzz.GenerateRequest{Int1: 2, Int2: 3, Limit: 6, Str1: "a", Str2: "b"},
			want: []string{"1", "a", "b", "a", "5", "ab"},
		},
		{
			name: "int1 == int2 always concatenates on multiples",
			req:  fizzbuzz.GenerateRequest{Int1: 2, Int2: 2, Limit: 4, Str1: "x", Str2: "y"},
			want: []string{"1", "xy", "3", "xy"},
		},
		{
			name: "limit of 1",
			req:  fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 1, Str1: "fizz", Str2: "buzz"},
			want: []string{"1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.req.Generate()
			if !reflect.DeepEqual(got.Result, tt.want) {
				t.Fatalf("Generate() = %v, want %v", got.Result, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./domain/fizzbuzz/ -run TestGenerateRequest_Generate -v`
Expected: build failure / FAIL — `fizzbuzz` package or `GenerateRequest` undefined.

- [ ] **Step 3: Write the implementation**

Create `domain/fizzbuzz/model.go`:
```go
// Package fizzbuzz holds the fizz-buzz business model: its value objects,
// generation logic, validation rules and sentinel errors. It depends on
// nothing above it (no use-case, transport or infrastructure imports).
package fizzbuzz

import "strconv"

// GenerateRequest is the set of parameters of a fizz-buzz generation.
type GenerateRequest struct {
	Int1  int
	Int2  int
	Limit int
	Str1  string
	Str2  string
}

// GenerateResponse is the result of a fizz-buzz generation.
type GenerateResponse struct {
	Result []string
}

// GetStatsRequest carries no parameter: the statistics query takes none.
type GetStatsRequest struct{}

// GetStatsResponse is the most frequently requested generation and its hit count.
type GetStatsResponse struct {
	Request GenerateRequest
	Hits    int
}

// Generate produces the fizz-buzz sequence from 1 to Limit. It assumes the
// request has been validated (Int1 and Int2 strictly positive); callers must
// call Validate first.
func (r GenerateRequest) Generate() GenerateResponse {
	result := make([]string, 0, r.Limit)

	for n := 1; n <= r.Limit; n++ {
		result = append(result, r.value(n))
	}

	return GenerateResponse{Result: result}
}

func (r GenerateRequest) value(n int) string {
	multipleOf1 := n%r.Int1 == 0
	multipleOf2 := n%r.Int2 == 0

	switch {
	case multipleOf1 && multipleOf2:
		return r.Str1 + r.Str2
	case multipleOf1:
		return r.Str1
	case multipleOf2:
		return r.Str2
	default:
		return strconv.Itoa(n)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./domain/fizzbuzz/ -run TestGenerateRequest_Generate -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add domain/fizzbuzz/model.go domain/fizzbuzz/model_test.go
git commit -m "feat(domain): fizz-buzz value objects and generation logic"
```

---

### Task 3: Domain sentinel errors & validation

**Files:**
- Create: `domain/fizzbuzz/error.go`
- Create: `domain/fizzbuzz/validator.go`
- Test: `domain/fizzbuzz/validator_test.go`

**Interfaces:**
- Consumes: `GenerateRequest` (Task 2).
- Produces:
  - `var ErrFailedToValidateGenerateRequest error`
  - `var ErrNoStatsRecorded error`
  - `func (r GenerateRequest) Validate(maxLimit int) error`

- [ ] **Step 1: Write the failing test**

Create `domain/fizzbuzz/validator_test.go`:
```go
package fizzbuzz_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

func TestGenerateRequest_Validate(t *testing.T) {
	t.Parallel()

	const maxLimit = 1000

	valid := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}

	tests := []struct {
		name    string
		mutate  func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest
		wantErr bool
	}{
		{"valid", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { return r }, false},
		{"int1 zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int1 = 0; return r }, true},
		{"int1 negative", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int1 = -1; return r }, true},
		{"int2 zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Int2 = 0; return r }, true},
		{"limit zero", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = 0; return r }, true},
		{"limit above max", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = maxLimit + 1; return r }, true},
		{"limit at max", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Limit = maxLimit; return r }, false},
		{"str1 empty", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str1 = ""; return r }, true},
		{"str2 empty", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str2 = ""; return r }, true},
		{"str1 too long", func(r fizzbuzz.GenerateRequest) fizzbuzz.GenerateRequest { r.Str1 = strings.Repeat("a", 101); return r }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.mutate(valid).Validate(maxLimit)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if !errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
					t.Fatalf("error %v does not wrap ErrFailedToValidateGenerateRequest", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./domain/fizzbuzz/ -run TestGenerateRequest_Validate -v`
Expected: build failure — `Validate` / `ErrFailedToValidateGenerateRequest` undefined.

- [ ] **Step 3: Write the sentinel errors**

Create `domain/fizzbuzz/error.go`:
```go
package fizzbuzz

import "errors"

// ErrFailedToValidateGenerateRequest is wrapped by every validation failure of
// a GenerateRequest. Callers classify validation errors with errors.Is.
var ErrFailedToValidateGenerateRequest = errors.New("invalid fizz-buzz parameters")

// ErrNoStatsRecorded is returned when no successful request has been recorded
// yet, so there is no "most frequent" request to report.
var ErrNoStatsRecorded = errors.New("no statistics recorded yet")
```

- [ ] **Step 4: Write the validator**

Create `domain/fizzbuzz/validator.go`:
```go
package fizzbuzz

import "fmt"

const maxStrLen = 100

// Validate checks the business invariants of a GenerateRequest. Every failure
// wraps ErrFailedToValidateGenerateRequest. maxLimit is supplied by the caller
// (configuration) rather than hard-coded.
func (r GenerateRequest) Validate(maxLimit int) error {
	switch {
	case r.Int1 <= 0:
		return fmt.Errorf("int1 must be a positive integer, got %d: %w", r.Int1, ErrFailedToValidateGenerateRequest)
	case r.Int2 <= 0:
		return fmt.Errorf("int2 must be a positive integer, got %d: %w", r.Int2, ErrFailedToValidateGenerateRequest)
	case r.Limit < 1 || r.Limit > maxLimit:
		return fmt.Errorf("limit must be between 1 and %d, got %d: %w", maxLimit, r.Limit, ErrFailedToValidateGenerateRequest)
	case r.Str1 == "":
		return fmt.Errorf("str1 must not be empty: %w", ErrFailedToValidateGenerateRequest)
	case r.Str2 == "":
		return fmt.Errorf("str2 must not be empty: %w", ErrFailedToValidateGenerateRequest)
	case len(r.Str1) > maxStrLen:
		return fmt.Errorf("str1 must be at most %d characters: %w", maxStrLen, ErrFailedToValidateGenerateRequest)
	case len(r.Str2) > maxStrLen:
		return fmt.Errorf("str2 must be at most %d characters: %w", maxStrLen, ErrFailedToValidateGenerateRequest)
	default:
		return nil
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./domain/fizzbuzz/ -v`
Expected: PASS (both Generate and Validate tests).

- [ ] **Step 6: Commit**

```bash
git add domain/fizzbuzz/error.go domain/fizzbuzz/validator.go domain/fizzbuzz/validator_test.go
git commit -m "feat(domain): sentinel errors and request validation"
```

---

### Task 4: Use-case — GenerateFizzBuzz

> **Amendment (post-review):** the generation loop (`for n := 1..limit` with the fizz-buzz `switch`)
> lives HERE in `Execute`, not in the domain. `Execute` does: `req.Validate(maxLimit)` → build the
> `[]string` result inline → `recorder.Record(req)`. The test asserts the actual fizz-buzz output
> (e.g. `["1","2","fizz","4","buzz"]`), not just the length. Add `import "strconv"`.

**Files:**
- Create: `usecase/generate_fizzbuzz.go`
- Test: `usecase/generate_fizzbuzz_test.go`

**Interfaces:**
- Consumes: `fizzbuzz.GenerateRequest`, `fizzbuzz.GenerateResponse`, `GenerateRequest.Validate`, `GenerateRequest.Generate`, `fizzbuzz.ErrFailedToValidateGenerateRequest`.
- Produces:
  - `type StatRecorder interface { Record(req fizzbuzz.GenerateRequest) }`
  - `type GenerateFizzBuzz struct { ... }`
  - `func NewGenerateFizzBuzz(maxLimit int, recorder StatRecorder) *GenerateFizzBuzz`
  - `func (uc *GenerateFizzBuzz) Execute(req fizzbuzz.GenerateRequest) (fizzbuzz.GenerateResponse, error)`

- [ ] **Step 1: Write the failing test**

Create `usecase/generate_fizzbuzz_test.go`:
```go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type spyRecorder struct {
	recorded []fizzbuzz.GenerateRequest
}

func (s *spyRecorder) Record(req fizzbuzz.GenerateRequest) {
	s.recorded = append(s.recorded, req)
}

func TestGenerateFizzBuzz_Execute_Valid(t *testing.T) {
	t.Parallel()

	rec := &spyRecorder{}
	uc := usecase.NewGenerateFizzBuzz(1000, rec)
	req := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 5, Str1: "fizz", Str2: "buzz"}

	resp, err := uc.Execute(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Result) != 5 {
		t.Fatalf("expected 5 results, got %d", len(resp.Result))
	}

	if len(rec.recorded) != 1 || rec.recorded[0] != req {
		t.Fatalf("expected request recorded once, got %v", rec.recorded)
	}
}

func TestGenerateFizzBuzz_Execute_Invalid(t *testing.T) {
	t.Parallel()

	rec := &spyRecorder{}
	uc := usecase.NewGenerateFizzBuzz(1000, rec)
	req := fizzbuzz.GenerateRequest{Int1: 0, Int2: 5, Limit: 5, Str1: "fizz", Str2: "buzz"}

	_, err := uc.Execute(req)
	if !errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if len(rec.recorded) != 0 {
		t.Fatalf("invalid request must not be recorded, got %v", rec.recorded)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./usecase/ -run TestGenerateFizzBuzz -v`
Expected: build failure — `usecase` package / `NewGenerateFizzBuzz` undefined.

- [ ] **Step 3: Write the implementation**

Create `usecase/generate_fizzbuzz.go`:
```go
// Package usecase orchestrates the application logic on top of the fizzbuzz
// domain. It defines the (segregated) interfaces it needs from infrastructure.
package usecase

import "github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"

// StatRecorder records a successful generation request for statistics.
type StatRecorder interface {
	Record(req fizzbuzz.GenerateRequest)
}

// GenerateFizzBuzz validates a request, generates the sequence, and records the
// request only on success.
type GenerateFizzBuzz struct {
	maxLimit int
	recorder StatRecorder
}

// NewGenerateFizzBuzz builds the use-case with its max-limit bound and recorder.
func NewGenerateFizzBuzz(maxLimit int, recorder StatRecorder) *GenerateFizzBuzz {
	return &GenerateFizzBuzz{maxLimit: maxLimit, recorder: recorder}
}

// Execute validates the request, generates the result and records the request.
func (uc *GenerateFizzBuzz) Execute(req fizzbuzz.GenerateRequest) (fizzbuzz.GenerateResponse, error) {
	if err := req.Validate(uc.maxLimit); err != nil {
		return fizzbuzz.GenerateResponse{}, err
	}

	resp := req.Generate()
	uc.recorder.Record(req)

	return resp, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./usecase/ -run TestGenerateFizzBuzz -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add usecase/generate_fizzbuzz.go usecase/generate_fizzbuzz_test.go
git commit -m "feat(usecase): generate fizzbuzz orchestration"
```

---

### Task 5: Use-case — GetFizzBuzzStats

**Files:**
- Create: `usecase/get_fizzbuzz_stats.go`
- Test: `usecase/get_fizzbuzz_stats_test.go`

**Interfaces:**
- Consumes: `fizzbuzz.GenerateRequest`, `fizzbuzz.GetStatsRequest`, `fizzbuzz.GetStatsResponse`, `fizzbuzz.ErrNoStatsRecorded`.
- Produces:
  - `type StatReader interface { MostFrequent() (fizzbuzz.GenerateRequest, int, bool) }`
  - `type GetFizzBuzzStats struct { ... }`
  - `func NewGetFizzBuzzStats(reader StatReader) *GetFizzBuzzStats`
  - `func (uc *GetFizzBuzzStats) Execute(_ fizzbuzz.GetStatsRequest) (fizzbuzz.GetStatsResponse, error)`

- [ ] **Step 1: Write the failing test**

Create `usecase/get_fizzbuzz_stats_test.go`:
```go
package usecase_test

import (
	"errors"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type stubReader struct {
	req  fizzbuzz.GenerateRequest
	hits int
	ok   bool
}

func (s stubReader) MostFrequent() (fizzbuzz.GenerateRequest, int, bool) {
	return s.req, s.hits, s.ok
}

func TestGetFizzBuzzStats_Execute_WithData(t *testing.T) {
	t.Parallel()

	want := fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"}
	uc := usecase.NewGetFizzBuzzStats(stubReader{req: want, hits: 7, ok: true})

	resp, err := uc.Execute(fizzbuzz.GetStatsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Request != want || resp.Hits != 7 {
		t.Fatalf("got %+v, want request %+v hits 7", resp, want)
	}
}

func TestGetFizzBuzzStats_Execute_Empty(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetFizzBuzzStats(stubReader{ok: false})

	_, err := uc.Execute(fizzbuzz.GetStatsRequest{})
	if !errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
		t.Fatalf("expected ErrNoStatsRecorded, got %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./usecase/ -run TestGetFizzBuzzStats -v`
Expected: build failure — `NewGetFizzBuzzStats` undefined.

- [ ] **Step 3: Write the implementation**

Create `usecase/get_fizzbuzz_stats.go`:
```go
package usecase

import "github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"

// StatReader reads the most frequently recorded generation request.
type StatReader interface {
	MostFrequent() (req fizzbuzz.GenerateRequest, hits int, ok bool)
}

// GetFizzBuzzStats returns the most frequent request and its hit count.
type GetFizzBuzzStats struct {
	reader StatReader
}

// NewGetFizzBuzzStats builds the use-case with its reader.
func NewGetFizzBuzzStats(reader StatReader) *GetFizzBuzzStats {
	return &GetFizzBuzzStats{reader: reader}
}

// Execute returns the most frequent request, or ErrNoStatsRecorded if none.
func (uc *GetFizzBuzzStats) Execute(_ fizzbuzz.GetStatsRequest) (fizzbuzz.GetStatsResponse, error) {
	req, hits, ok := uc.reader.MostFrequent()
	if !ok {
		return fizzbuzz.GetStatsResponse{}, fizzbuzz.ErrNoStatsRecorded
	}

	return fizzbuzz.GetStatsResponse{Request: req, Hits: hits}, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./usecase/ -v`
Expected: PASS (all use-case tests).

- [ ] **Step 5: Commit**

```bash
git add usecase/get_fizzbuzz_stats.go usecase/get_fizzbuzz_stats_test.go
git commit -m "feat(usecase): get fizzbuzz stats"
```

---

### Task 6: Infrastructure — in-memory stat store + Redis stub

**Files:**
- Create: `infrastructure/statstorer/in_memory.go`
- Create: `infrastructure/statstorer/redis.go`
- Test: `infrastructure/statstorer/in_memory_test.go`

**Interfaces:**
- Consumes: `fizzbuzz.GenerateRequest`.
- Produces (satisfies `usecase.StatRecorder` and `usecase.StatReader`):
  - `func NewInMemory() *InMemory`
  - `func (s *InMemory) Record(req fizzbuzz.GenerateRequest)`
  - `func (s *InMemory) MostFrequent() (fizzbuzz.GenerateRequest, int, bool)`
  - `type Redis struct{}` + `NewRedis()` + same two methods (panic placeholders).

- [ ] **Step 1: Write the failing test**

Create `infrastructure/statstorer/in_memory_test.go`:
```go
package statstorer_test

import (
	"sync"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
)

func req(int1 int) fizzbuzz.GenerateRequest {
	return fizzbuzz.GenerateRequest{Int1: int1, Int2: 5, Limit: 10, Str1: "fizz", Str2: "buzz"}
}

func TestInMemory_Empty(t *testing.T) {
	t.Parallel()

	_, _, ok := statstorer.NewInMemory().MostFrequent()
	if ok {
		t.Fatal("expected ok=false on empty store")
	}
}

func TestInMemory_MostFrequent(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	s.Record(a)
	s.Record(a)
	s.Record(a)
	s.Record(b)

	got, hits, ok := s.MostFrequent()
	if !ok || got != a || hits != 3 {
		t.Fatalf("got %+v hits=%d ok=%v, want %+v hits=3", got, hits, ok, a)
	}
}

func TestInMemory_TieBreakFirstToReachMax(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a, b := req(3), req(7)
	s.Record(a) // a=1
	s.Record(a) // a=2  -> top a
	s.Record(b) // b=1
	s.Record(b) // b=2  -> not strictly greater, top stays a

	got, hits, _ := s.MostFrequent()
	if got != a || hits != 2 {
		t.Fatalf("tie must keep first to reach max: got %+v hits=%d, want %+v hits=2", got, hits, a)
	}
}

func TestInMemory_ConcurrentRecord(t *testing.T) {
	t.Parallel()

	s := statstorer.NewInMemory()
	a := req(3)

	const goroutines, perGoroutine = 50, 100

	var wg sync.WaitGroup

	wg.Add(goroutines)

	for range goroutines {
		go func() {
			defer wg.Done()

			for range perGoroutine {
				s.Record(a)
			}
		}()
	}

	wg.Wait()

	_, hits, ok := s.MostFrequent()
	if !ok || hits != goroutines*perGoroutine {
		t.Fatalf("got hits=%d ok=%v, want %d", hits, ok, goroutines*perGoroutine)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./infrastructure/statstorer/ -v`
Expected: build failure — `statstorer.NewInMemory` undefined.

- [ ] **Step 3: Write the in-memory store**

Create `infrastructure/statstorer/in_memory.go`:
```go
// Package statstorer holds implementations of the use-case stat interfaces
// (StatRecorder, StatReader): an in-memory store and a Redis placeholder.
package statstorer

import (
	"sync"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

// InMemory is a concurrency-safe, process-local stat counter. The current most
// frequent request is memoized: it is updated only when a count becomes
// strictly greater than the current maximum, so on ties the first request to
// reach the maximum is kept. State is lost on restart.
type InMemory struct {
	mu      sync.Mutex
	counts  map[fizzbuzz.GenerateRequest]int
	topReq  fizzbuzz.GenerateRequest
	topHits int
}

// NewInMemory builds an empty in-memory stat store.
func NewInMemory() *InMemory {
	return &InMemory{counts: make(map[fizzbuzz.GenerateRequest]int)}
}

// Record increments the counter for req.
func (s *InMemory) Record(req fizzbuzz.GenerateRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[req]++

	if s.counts[req] > s.topHits {
		s.topHits = s.counts[req]
		s.topReq = req
	}
}

// MostFrequent returns the most frequent request, its hit count, and whether
// any request has been recorded.
func (s *InMemory) MostFrequent() (fizzbuzz.GenerateRequest, int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.topHits == 0 {
		return fizzbuzz.GenerateRequest{}, 0, false
	}

	return s.topReq, s.topHits, true
}
```

- [ ] **Step 4: Write the Redis placeholder**

Create `infrastructure/statstorer/redis.go`:
```go
package statstorer

import "github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"

// Redis is a placeholder for a durable, shared stat store. It demonstrates how
// a distributed backend would plug into the same StatRecorder/StatReader
// interfaces without touching the use-cases.
type Redis struct{}

// NewRedis builds the placeholder.
func NewRedis() *Redis {
	return &Redis{}
}

// Record is not implemented yet.
func (s *Redis) Record(_ fizzbuzz.GenerateRequest) {
	panic("implement me: durable stat recording via Redis")
}

// MostFrequent is not implemented yet.
func (s *Redis) MostFrequent() (fizzbuzz.GenerateRequest, int, bool) {
	panic("implement me: durable stat reading via Redis")
}
```

- [ ] **Step 5: Run tests (with race detector)**

Run: `go test -race ./infrastructure/statstorer/ -v`
Expected: PASS, no race reported.

- [ ] **Step 6: Commit**

```bash
git add infrastructure/statstorer/
git commit -m "feat(infra): in-memory stat store with Redis placeholder"
```

---

### Task 7: Infrastructure — in-memory rate limiter + Redis stub

**Files:**
- Create: `infrastructure/ratelimiter/in_memory.go`
- Create: `infrastructure/ratelimiter/redis.go`
- Test: `infrastructure/ratelimiter/in_memory_test.go`

**Interfaces:**
- Consumes: `golang.org/x/time/rate`.
- Produces (satisfies the `server.RateLimiter` interface defined in Task 11 — `Allow() bool`):
  - `func NewInMemory(perSecond float64, burst int) *InMemory`
  - `func (l *InMemory) Allow() bool`
  - `type Redis struct{}` + `NewRedis()` + `Allow()` (panic placeholder).

- [ ] **Step 1: Write the failing test**

Create `infrastructure/ratelimiter/in_memory_test.go`:
```go
package ratelimiter_test

import (
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/ratelimiter"
)

func TestInMemory_AllowsUpToBurstThenRejects(t *testing.T) {
	t.Parallel()

	// 0 refills/sec, burst 3: exactly 3 allowed, the 4th rejected.
	l := ratelimiter.NewInMemory(0, 3)

	for i := range 3 {
		if !l.Allow() {
			t.Fatalf("call %d should be allowed within burst", i+1)
		}
	}

	if l.Allow() {
		t.Fatal("call beyond burst should be rejected")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./infrastructure/ratelimiter/ -v`
Expected: build failure — `ratelimiter.NewInMemory` undefined.

- [ ] **Step 3: Write the in-memory limiter**

Create `infrastructure/ratelimiter/in_memory.go`:
```go
// Package ratelimiter holds rate-limiter implementations used by the HTTP
// middleware: a process-local token-bucket limiter and a Redis placeholder.
//
// A process-local limiter is a per-instance guard only. In a horizontally
// scaled deployment the authoritative limit belongs at the edge (gateway /
// ingress / load balancer) or a shared store; see the project ADR.
package ratelimiter

import "golang.org/x/time/rate"

// InMemory is a token-bucket limiter backed by golang.org/x/time/rate.
type InMemory struct {
	limiter *rate.Limiter
}

// NewInMemory builds a limiter allowing perSecond sustained requests with the
// given burst capacity.
func NewInMemory(perSecond float64, burst int) *InMemory {
	return &InMemory{limiter: rate.NewLimiter(rate.Limit(perSecond), burst)}
}

// Allow reports whether a request may proceed now (non-blocking).
func (l *InMemory) Allow() bool {
	return l.limiter.Allow()
}
```

- [ ] **Step 4: Write the Redis placeholder**

Create `infrastructure/ratelimiter/redis.go`:
```go
package ratelimiter

// Redis is a placeholder for a distributed rate limiter backed by a shared
// store, which is what an authoritative global limit requires when scaling out.
type Redis struct{}

// NewRedis builds the placeholder.
func NewRedis() *Redis {
	return &Redis{}
}

// Allow is not implemented yet.
func (l *Redis) Allow() bool {
	panic("implement me: distributed rate limiting via a shared store")
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./infrastructure/ratelimiter/ -v`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add infrastructure/ratelimiter/
git commit -m "feat(infra): in-memory rate limiter with Redis placeholder"
```

---

### Task 8: Configuration

**Files:**
- Create: `config/config.go`
- Test: `config/config_test.go`

**Interfaces:**
- Consumes: `github.com/sethvargo/go-envconfig`.
- Produces:
  - `const AppName = "fizz-buzz-api"`; `var AppVersion string`
  - `type Config struct { HTTPAddr string; MaxLimit int; RateLimitPerSec float64; RateLimitBurst int; ReadHeaderTimeout, WriteTimeout, IdleTimeout time.Duration; LogLevel string }`
  - `func New(ctx context.Context) (*Config, error)`

- [ ] **Step 1: Write the failing test**

Create `config/config_test.go`:
```go
package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

func TestNew_Defaults(t *testing.T) {
	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}

	if cfg.MaxLimit != 10000 {
		t.Errorf("MaxLimit = %d, want 10000", cfg.MaxLimit)
	}
}

func TestNew_FromEnv(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("MAX_LIMIT", "42")
	t.Setenv("WRITE_TIMEOUT", "30s")

	cfg, err := config.New(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.HTTPAddr != ":9090" || cfg.MaxLimit != 42 || cfg.WriteTimeout != 30*time.Second {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}
```

> Note: `TestNew_Defaults` does not call `t.Parallel()` because `TestNew_FromEnv` uses `t.Setenv`, which forbids parallelism in the same package.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./config/ -v`
Expected: build failure — `config.New` undefined.

- [ ] **Step 3: Write the implementation**

Create `config/config.go`:
```go
// Package config loads the service configuration from environment variables.
package config

import (
	"context"
	"fmt"
	"time"

	"github.com/sethvargo/go-envconfig"
)

// AppName is the service name used in logs.
const AppName = "fizz-buzz-api"

// AppVersion is set at build time via -ldflags; defaults to DEV.
//
//nolint:gochecknoglobals // build-time version, the single allowed global var
var AppVersion = "DEV"

// Config holds all runtime configuration, populated from the environment.
type Config struct {
	HTTPAddr          string        `env:"HTTP_ADDR, default=:8080"`
	MaxLimit          int           `env:"MAX_LIMIT, default=10000"`
	RateLimitPerSec   float64       `env:"RATE_LIMIT_PER_SEC, default=50"`
	RateLimitBurst    int           `env:"RATE_LIMIT_BURST, default=100"`
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT, default=2s"`
	WriteTimeout      time.Duration `env:"WRITE_TIMEOUT, default=10s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT, default=120s"`
	LogLevel          string        `env:"LOG_LEVEL, default=info"`
}

// New loads configuration from the environment, applying defaults.
func New(ctx context.Context) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(ctx, cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
	}

	return cfg, nil
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./config/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add config/config.go config/config_test.go
git commit -m "feat(config): env-var configuration with defaults"
```

---

### Task 9: HTTP handler — generate (with parsing & response helpers)

**Files:**
- Create: `presentation/http/handler/response.go`
- Create: `presentation/http/handler/request.go`
- Create: `presentation/http/handler/generate_fizzbuzz.go`
- Test: `presentation/http/handler/generate_fizzbuzz_test.go`

**Interfaces:**
- Consumes: `usecase.GenerateFizzBuzz`, `usecase.NewGenerateFizzBuzz`, `statstorer.NewInMemory`, `fizzbuzz.GenerateRequest`, `fizzbuzz.ErrFailedToValidateGenerateRequest`.
- Produces:
  - `func GenerateFizzBuzz(uc *usecase.GenerateFizzBuzz, logger *slog.Logger) http.HandlerFunc`
  - `func writeJSON(w http.ResponseWriter, status int, v any)` (unexported, package-internal)
  - `func writeError(w http.ResponseWriter, status int, message string)` (unexported)
  - `func parseGenerateRequest(r *http.Request) (fizzbuzz.GenerateRequest, error)` (unexported)

- [ ] **Step 1: Write the failing test**

Create `presentation/http/handler/generate_fizzbuzz_test.go`:
```go
package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

func newGenerateHandler() http.HandlerFunc {
	uc := usecase.NewGenerateFizzBuzz(10000, statstorer.NewInMemory())

	return handler.GenerateFizzBuzz(uc, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestGenerateFizzBuzz_OK(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=3&int2=5&limit=5&str1=fizz&str2=buzz", nil)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("content-type = %q, want application/json", ct)
	}

	var got []string
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("body is not a JSON array of strings: %v", err)
	}

	want := []string{"1", "2", "fizz", "4", "buzz"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestGenerateFizzBuzz_ValidationError(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=0&int2=5&limit=5&str1=fizz&str2=buzz", nil)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGenerateFizzBuzz_MalformedQuery(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/fizzbuzz?int1=abc&int2=5&limit=5&str1=fizz&str2=buzz", nil)

	newGenerateHandler().ServeHTTP(rec, r)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

var _ = io.Discard // keep import if unused after edits
```

> Remove the trailing `var _ = io.Discard` line if `io` is already used (it is, via `io.Discard`).

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./presentation/http/handler/ -run TestGenerateFizzBuzz -v`
Expected: build failure — `handler.GenerateFizzBuzz` undefined.

- [ ] **Step 3: Write the response helpers**

Create `presentation/http/handler/response.go`:
```go
// Package handler contains the HTTP handlers and their request/response helpers.
package handler

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes the message as a JSON string body with the given status.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, message)
}
```

- [ ] **Step 4: Write the request parser**

Create `presentation/http/handler/request.go`:
```go
package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
)

var errInvalidQueryParam = errors.New("invalid or missing query parameter")

func parseGenerateRequest(r *http.Request) (fizzbuzz.GenerateRequest, error) {
	q := r.URL.Query()

	int1, err := parseIntParam(q, "int1")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	int2, err := parseIntParam(q, "int2")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	limit, err := parseIntParam(q, "limit")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	str1, err := parseStringParam(q, "str1")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	str2, err := parseStringParam(q, "str2")
	if err != nil {
		return fizzbuzz.GenerateRequest{}, err
	}

	return fizzbuzz.GenerateRequest{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}, nil
}

func parseIntParam(q url.Values, name string) (int, error) {
	raw := q.Get(name)
	if raw == "" {
		return 0, fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", name, errInvalidQueryParam)
	}

	return v, nil
}

func parseStringParam(q url.Values, name string) (string, error) {
	raw := q.Get(name)
	if raw == "" {
		return "", fmt.Errorf("%s is required: %w", name, errInvalidQueryParam)
	}

	return raw, nil
}
```

- [ ] **Step 5: Write the handler**

Create `presentation/http/handler/generate_fizzbuzz.go`:
```go
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// GenerateFizzBuzz parses the query, runs the use-case and writes the result.
// Parsing and validation failures map to 400; unexpected failures to 500.
func GenerateFizzBuzz(uc *usecase.GenerateFizzBuzz, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parseGenerateRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())

			return
		}

		resp, err := uc.Execute(req)
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrFailedToValidateGenerateRequest) {
				writeError(w, http.StatusBadRequest, err.Error())

				return
			}

			logger.Error("generate fizzbuzz failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		writeJSON(w, http.StatusOK, resp.Result)
	}
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `go test ./presentation/http/handler/ -run TestGenerateFizzBuzz -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add presentation/http/handler/response.go presentation/http/handler/request.go presentation/http/handler/generate_fizzbuzz.go presentation/http/handler/generate_fizzbuzz_test.go
git commit -m "feat(http): generate fizzbuzz handler with query parsing"
```

---

### Task 10: HTTP handler — stats

**Files:**
- Create: `presentation/http/handler/get_fizzbuzz_stats.go`
- Test: `presentation/http/handler/get_fizzbuzz_stats_test.go`

**Interfaces:**
- Consumes: `usecase.GetFizzBuzzStats`, `usecase.NewGetFizzBuzzStats`, `statstorer.InMemory`, `fizzbuzz.GetStatsResponse`, `fizzbuzz.ErrNoStatsRecorded`, `writeJSON`/`writeError` (Task 9).
- Produces:
  - `func GetFizzBuzzStats(uc *usecase.GetFizzBuzzStats, logger *slog.Logger) http.HandlerFunc`

- [ ] **Step 1: Write the failing test**

Create `presentation/http/handler/get_fizzbuzz_stats_test.go`:
```go
package handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

func TestGetFizzBuzzStats_Empty404(t *testing.T) {
	t.Parallel()

	uc := usecase.NewGetFizzBuzzStats(statstorer.NewInMemory())
	h := handler.GetFizzBuzzStats(uc, slog.New(slog.NewTextHandler(io.Discard, nil)))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fizzbuzz/stats", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestGetFizzBuzzStats_OK(t *testing.T) {
	t.Parallel()

	store := statstorer.NewInMemory()
	store.Record(fizzbuzz.GenerateRequest{Int1: 3, Int2: 5, Limit: 100, Str1: "fizz", Str2: "buzz"})

	uc := usecase.NewGetFizzBuzzStats(store)
	h := handler.GetFizzBuzzStats(uc, slog.New(slog.NewTextHandler(io.Discard, nil)))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/fizzbuzz/stats", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Request struct {
			Int1  int    `json:"int1"`
			Int2  int    `json:"int2"`
			Limit int    `json:"limit"`
			Str1  string `json:"str1"`
			Str2  string `json:"str2"`
		} `json:"request"`
		Hits int `json:"hits"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if body.Hits != 1 || body.Request.Int1 != 3 || body.Request.Str2 != "buzz" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./presentation/http/handler/ -run TestGetFizzBuzzStats -v`
Expected: build failure — `handler.GetFizzBuzzStats` undefined.

- [ ] **Step 3: Write the handler**

Create `presentation/http/handler/get_fizzbuzz_stats.go`:
```go
package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/domain/fizzbuzz"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

type statsDTO struct {
	Request statsRequestDTO `json:"request"`
	Hits    int             `json:"hits"`
}

type statsRequestDTO struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

func newStatsDTO(resp fizzbuzz.GetStatsResponse) statsDTO {
	return statsDTO{
		Request: statsRequestDTO{
			Int1:  resp.Request.Int1,
			Int2:  resp.Request.Int2,
			Limit: resp.Request.Limit,
			Str1:  resp.Request.Str1,
			Str2:  resp.Request.Str2,
		},
		Hits: resp.Hits,
	}
}

// GetFizzBuzzStats returns the most frequent request as JSON, or 404 if none.
func GetFizzBuzzStats(uc *usecase.GetFizzBuzzStats, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		resp, err := uc.Execute(fizzbuzz.GetStatsRequest{})
		if err != nil {
			if errors.Is(err, fizzbuzz.ErrNoStatsRecorded) {
				writeError(w, http.StatusNotFound, "no statistics recorded yet")

				return
			}

			logger.Error("get fizzbuzz stats failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")

			return
		}

		writeJSON(w, http.StatusOK, newStatsDTO(resp))
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./presentation/http/handler/ -v`
Expected: PASS (all handler tests).

- [ ] **Step 5: Commit**

```bash
git add presentation/http/handler/get_fizzbuzz_stats.go presentation/http/handler/get_fizzbuzz_stats_test.go
git commit -m "feat(http): get fizzbuzz stats handler"
```

---

### Task 11: HTTP middleware

**Files:**
- Create: `presentation/http/server/middleware.go`
- Test: `presentation/http/server/middleware_test.go`

**Interfaces:**
- Consumes: `log/slog`.
- Produces:
  - `type RateLimiter interface { Allow() bool }`
  - `func Recovery(logger *slog.Logger) func(http.Handler) http.Handler`
  - `func RequestID(next http.Handler) http.Handler`
  - `func Logging(logger *slog.Logger) func(http.Handler) http.Handler`
  - `func RateLimit(limiter RateLimiter) func(http.Handler) http.Handler`
  - `func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler`

- [ ] **Step 1: Write the failing test**

Create `presentation/http/server/middleware_test.go`:
```go
package server_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

type stubLimiter struct{ allow bool }

func (s stubLimiter) Allow() bool { return s.allow }

func TestRecovery_TurnsPanicInto500(t *testing.T) {
	t.Parallel()

	panicky := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	h := server.Recovery(discardLogger())(panicky)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
}

func TestRequestID_SetsHeader(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := server.RequestID(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRateLimit_Rejects(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	h := server.RateLimit(stubLimiter{allow: false})(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
}

func TestRateLimit_Allows(t *testing.T) {
	t.Parallel()

	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := server.RateLimit(stubLimiter{allow: true})(ok)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418 (passed through)", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./presentation/http/server/ -v`
Expected: build failure — `server.Recovery` etc. undefined.

- [ ] **Step 3: Write the middleware**

Create `presentation/http/server/middleware.go`:
```go
// Package server wires the HTTP router, middleware stack and server lifecycle.
package server

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

const requestIDHeader = "X-Request-ID"

// RateLimiter decides whether a request may proceed now.
type RateLimiter interface {
	Allow() bool
}

// statusRecorder captures the response status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Recovery converts panics into a 500 response and logs them.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "panic", rec, "path", r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RequestID assigns a request id (honouring an inbound X-Request-ID) and echoes
// it on the response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(requestIDHeader)
		if id == "" {
			id = newRequestID()
		}

		w.Header().Set(requestIDHeader, id)
		next.ServeHTTP(w, r)
	})
}

// Logging logs one structured line per request with method, path, status and duration.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(sr, r)

			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sr.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", w.Header().Get(requestIDHeader),
			)
		})
	}
}

// RateLimit rejects requests with 429 when the limiter denies them.
func RateLimit(limiter RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "too many requests", http.StatusTooManyRequests)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Chain wraps h with the given middlewares so the first listed is outermost.
func Chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	return h
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}

	return hex.EncodeToString(b)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./presentation/http/server/ -v`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add presentation/http/server/middleware.go presentation/http/server/middleware_test.go
git commit -m "feat(http): recovery, request-id, logging and rate-limit middleware"
```

---

### Task 12: Router & server lifecycle

**Files:**
- Create: `presentation/http/server/routes.go`
- Create: `presentation/http/server/server.go`
- Test: `presentation/http/server/routes_test.go`

**Interfaces:**
- Consumes: `log/slog`.
- Produces:
  - `func NewRouter(generate, stats http.HandlerFunc) http.Handler`
  - `type Server struct { ... }`
  - `func New(srv *http.Server, logger *slog.Logger) *Server`
  - `func (s *Server) Start(errChan chan<- error)`
  - `func (s *Server) Stop(ctx context.Context) error`

- [ ] **Step 1: Write the failing test**

Create `presentation/http/server/routes_test.go`:
```go
package server_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

func TestNewRouter_RoutesAndHealthz(t *testing.T) {
	t.Parallel()

	generate := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("gen")) }
	stats := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK); _, _ = w.Write([]byte("stats")) }

	router := server.NewRouter(generate, stats)

	cases := []struct {
		path     string
		wantCode int
		wantBody string
	}{
		{"/healthz", http.StatusOK, ""},
		{"/fizzbuzz", http.StatusOK, "gen"},
		{"/fizzbuzz/stats", http.StatusOK, "stats"},
		{"/unknown", http.StatusNotFound, ""},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rec.Code != tc.wantCode {
				t.Fatalf("path %s: status = %d, want %d", tc.path, rec.Code, tc.wantCode)
			}

			if tc.wantBody != "" && rec.Body.String() != tc.wantBody {
				t.Fatalf("path %s: body = %q, want %q", tc.path, rec.Body.String(), tc.wantBody)
			}
		})
	}
}

func TestNewRouter_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	router := server.NewRouter(
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) },
	)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/fizzbuzz", nil))

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./presentation/http/server/ -run TestNewRouter -v`
Expected: build failure — `server.NewRouter` undefined.

- [ ] **Step 3: Write the router**

Create `presentation/http/server/routes.go`:
```go
package server

import "net/http"

// NewRouter builds the ServeMux with the application routes plus a health
// probe. Method+path patterns (Go 1.22+) yield 405 on method mismatch and 404
// on unknown paths automatically.
func NewRouter(generate, stats http.HandlerFunc) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /fizzbuzz", generate)
	mux.HandleFunc("GET /fizzbuzz/stats", stats)

	return mux
}
```

- [ ] **Step 4: Write the server lifecycle**

Create `presentation/http/server/server.go`:
```go
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Server wraps an *http.Server with start/stop lifecycle helpers.
type Server struct {
	srv    *http.Server
	logger *slog.Logger
}

// New builds a Server around a configured *http.Server.
func New(srv *http.Server, logger *slog.Logger) *Server {
	return &Server{srv: srv, logger: logger}
}

// Start runs ListenAndServe in a goroutine, forwarding any non-graceful error
// to errChan.
func (s *Server) Start(errChan chan<- error) {
	go func() {
		s.logger.Info("starting HTTP server", "addr", s.srv.Addr)

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("http server: %w", err)
		}
	}()
}

// Stop gracefully shuts the server down, respecting ctx's deadline.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	return nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./presentation/http/server/ -v`
Expected: PASS (middleware + routes tests).

- [ ] **Step 6: Commit**

```bash
git add presentation/http/server/routes.go presentation/http/server/server.go presentation/http/server/routes_test.go
git commit -m "feat(http): router and graceful server lifecycle"
```

---

### Task 13: DI container & HTTP server wiring (with integration test)

**Files:**
- Create: `infrastructure/di/logger.go`
- Create: `infrastructure/di/container.go`
- Create: `infrastructure/di/http_server.go`
- Test: `infrastructure/di/http_server_test.go`

**Interfaces:**
- Consumes: `config.Config`, `statstorer.NewInMemory`, `ratelimiter.NewInMemory`, `usecase.NewGenerateFizzBuzz`, `usecase.NewGetFizzBuzzStats`, `handler.GenerateFizzBuzz`, `handler.GetFizzBuzzStats`, `server.NewRouter`, `server.Chain`, `server.Recovery/RequestID/Logging/RateLimit`, `server.New`.
- Produces:
  - `func NewContainer(ctx context.Context, cfg *config.Config) *Container`
  - `func (c *Container) GetLogger() *slog.Logger`
  - `func (c *Container) GetHTTPServer() *server.Server`
  - `func (c *Container) httpHandler() http.Handler` (unexported, used by the test)

- [ ] **Step 1: Write the failing integration test**

Create `infrastructure/di/http_server_test.go`:
```go
package di_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/di"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg := &config.Config{
		HTTPAddr:        ":0",
		MaxLimit:        10000,
		RateLimitPerSec: 1000,
		RateLimitBurst:  1000,
		LogLevel:        "error",
	}

	c := di.NewContainer(context.Background(), cfg)
	ts := httptest.NewServer(c.HTTPHandler())

	t.Cleanup(ts.Close)

	return ts
}

func TestIntegration_GenerateThenStats(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/fizzbuzz?int1=3&int2=5&limit=5&str1=fizz&str2=buzz")
	if err != nil {
		t.Fatalf("GET /fizzbuzz: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("generate status = %d, want 200", resp.StatusCode)
	}

	var arr []string
	if err := json.NewDecoder(resp.Body).Decode(&arr); err != nil {
		t.Fatalf("decode generate body: %v", err)
	}

	if len(arr) != 5 {
		t.Fatalf("got %d items, want 5", len(arr))
	}

	statsResp, err := http.Get(ts.URL + "/fizzbuzz/stats")
	if err != nil {
		t.Fatalf("GET /fizzbuzz/stats: %v", err)
	}

	defer func() { _ = statsResp.Body.Close() }()

	if statsResp.StatusCode != http.StatusOK {
		t.Fatalf("stats status = %d, want 200", statsResp.StatusCode)
	}
}

func TestIntegration_StatsEmpty404(t *testing.T) {
	t.Parallel()

	ts := newTestServer(t)

	resp, err := http.Get(ts.URL + "/fizzbuzz/stats")
	if err != nil {
		t.Fatalf("GET /fizzbuzz/stats: %v", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("stats status = %d, want 404", resp.StatusCode)
	}
}
```

> Note: the test calls `c.HTTPHandler()` (exported) to get the middleware-wrapped handler without binding a port. Expose it as exported in Step 4.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./infrastructure/di/ -v`
Expected: build failure — `di.NewContainer` / `HTTPHandler` undefined.

- [ ] **Step 3: Write the logger**

Create `infrastructure/di/logger.go`:
```go
package di

import (
	"log/slog"
	"os"

	"github.com/Pimousse1099/fizz-buzz-api/config"
)

// GetLogger returns the memoized structured JSON logger.
func (c *Container) GetLogger() *slog.Logger {
	if c.logger == nil {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(c.config.LogLevel)})
		c.logger = slog.New(handler).With(
			"app", config.AppName,
			"version", config.AppVersion,
		)
	}

	return c.logger
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

- [ ] **Step 4: Write the container & HTTP wiring**

Create `infrastructure/di/container.go`:
```go
// Package di is the inversion-of-control container: it constructs and memoizes
// the application's dependencies. Getters are lazy (built on first use).
package di

import (
	"context"
	"log/slog"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/statstorer"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
	"github.com/Pimousse1099/fizz-buzz-api/usecase"
)

// Container holds configuration and the lazily-built singletons.
type Container struct {
	ctx    context.Context //nolint:containedctx // base context for server lifecycle
	config *config.Config

	logger     *slog.Logger
	statStore  *statstorer.InMemory
	httpServer *server.Server

	generateUC *usecase.GenerateFizzBuzz
	statsUC    *usecase.GetFizzBuzzStats
}

// NewContainer builds a container from the base context and configuration.
func NewContainer(ctx context.Context, cfg *config.Config) *Container {
	return &Container{ctx: ctx, config: cfg}
}

func (c *Container) getStatStore() *statstorer.InMemory {
	if c.statStore == nil {
		c.statStore = statstorer.NewInMemory()
	}

	return c.statStore
}

func (c *Container) getGenerateUseCase() *usecase.GenerateFizzBuzz {
	if c.generateUC == nil {
		c.generateUC = usecase.NewGenerateFizzBuzz(c.config.MaxLimit, c.getStatStore())
	}

	return c.generateUC
}

func (c *Container) getStatsUseCase() *usecase.GetFizzBuzzStats {
	if c.statsUC == nil {
		c.statsUC = usecase.NewGetFizzBuzzStats(c.getStatStore())
	}

	return c.statsUC
}
```

Create `infrastructure/di/http_server.go`:
```go
package di

import (
	"net/http"

	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/ratelimiter"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/handler"
	"github.com/Pimousse1099/fizz-buzz-api/presentation/http/server"
)

// HTTPHandler builds the fully-wired HTTP handler (router + middleware stack).
// Exposed so tests can exercise it via httptest without binding a port.
func (c *Container) HTTPHandler() http.Handler {
	logger := c.GetLogger()

	router := server.NewRouter(
		handler.GenerateFizzBuzz(c.getGenerateUseCase(), logger),
		handler.GetFizzBuzzStats(c.getStatsUseCase(), logger),
	)

	limiter := ratelimiter.NewInMemory(c.config.RateLimitPerSec, c.config.RateLimitBurst)

	return server.Chain(
		router,
		server.Recovery(logger),
		server.RequestID,
		server.Logging(logger),
		server.RateLimit(limiter),
	)
}

// GetHTTPServer builds the memoized HTTP server with timeouts from config.
func (c *Container) GetHTTPServer() *server.Server {
	if c.httpServer == nil {
		srv := &http.Server{
			Addr:              c.config.HTTPAddr,
			Handler:           c.HTTPHandler(),
			ReadHeaderTimeout: c.config.ReadHeaderTimeout,
			WriteTimeout:      c.config.WriteTimeout,
			IdleTimeout:       c.config.IdleTimeout,
			MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		}

		c.httpServer = server.New(srv, c.GetLogger())
	}

	return c.httpServer
}
```

- [ ] **Step 5: Run tests (with race detector)**

Run: `go test -race ./infrastructure/di/ -v`
Expected: PASS, no race.

- [ ] **Step 6: Commit**

```bash
git add infrastructure/di/
git commit -m "feat(di): container, logger and HTTP server wiring"
```

---

### Task 14: Entrypoint, README & full verification

**Files:**
- Create: `cmd/main.go`
- Modify: `README.md` (append a "Run & evaluate" section)

**Interfaces:**
- Consumes: `config.New`, `di.NewContainer`, `Container.GetLogger`, `Container.GetHTTPServer`, `Server.Start`, `Server.Stop`.
- Produces: the `main` package / runnable binary.

- [ ] **Step 1: Write the entrypoint**

Create `cmd/main.go`:
```go
// Command fizz-buzz-api starts the HTTP server.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pimousse1099/fizz-buzz-api/config"
	"github.com/Pimousse1099/fizz-buzz-api/infrastructure/di"
)

const shutdownTimeout = 10 * time.Second

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	container := di.NewContainer(ctx, cfg)
	logger := container.GetLogger()

	httpSrv := container.GetHTTPServer()
	errChan := make(chan error, 1)
	httpSrv.Start(errChan)

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case startErr := <-errChan:
		logger.Error("server failed", "error", startErr)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if stopErr := httpSrv.Stop(shutdownCtx); stopErr != nil {
		logger.Error("graceful shutdown failed", "error", stopErr)
	}
}
```

- [ ] **Step 2: Verify it builds and runs**

Run:
```bash
go build ./...
go run ./cmd &
sleep 1
curl -s 'http://localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz'
curl -s 'http://localhost:8080/fizzbuzz/stats'
curl -s -o /dev/null -w '%{http_code}\n' 'http://localhost:8080/fizzbuzz?int1=0&int2=5&limit=5&str1=a&str2=b'
kill %1
```
Expected: the array `["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]`, then a stats JSON object with `"hits":1`, then `400`.

- [ ] **Step 3: Append run instructions to `README.md`**

Add this section to the end of `README.md`:
```markdown
## Run & evaluate

Requirements: Go 1.26+.

```sh
# Run the server (defaults to :8080)
go run ./cmd

# Generate
curl 'http://localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz'

# Most frequent request (404 until at least one successful generation)
curl 'http://localhost:8080/fizzbuzz/stats'

# Health
curl http://localhost:8080/healthz
```

Configuration (environment variables): `HTTP_ADDR` (`:8080`), `MAX_LIMIT` (`10000`),
`RATE_LIMIT_PER_SEC` (`50`), `RATE_LIMIT_BURST` (`100`), `READ_HEADER_TIMEOUT` (`2s`),
`WRITE_TIMEOUT` (`10s`), `IDLE_TIMEOUT` (`120s`), `LOG_LEVEL` (`info`).

Development:

```sh
make build   # go build ./...
make race    # go test -race ./...
make lint    # golangci-lint run
```

Architecture decisions: see `docs/architecture-decision-records/2026-06-23-fizz-buzz-api.md`.
```

- [ ] **Step 4: Run the full verification gate**

Run:
```bash
go mod tidy
go build ./...
go test -race ./...
golangci-lint run
```
Expected: all pass. If `golangci-lint` reports formatter findings, run `golangci-lint run --fix` and re-run.

- [ ] **Step 5: Commit**

```bash
git add cmd/main.go README.md go.mod go.sum
git commit -m "feat(cmd): server entrypoint with graceful shutdown; docs: run instructions"
```

---

## Self-Review

**Spec coverage (against the ADR):**
- §2.1 stdlib net/http + hand-written middleware → Tasks 11–13. ✅
- §2.2 GET + query params → Task 9 (parsing), Task 12 (route). ✅
- §2.3 raw JSON array success / JSON string error → Task 9 (`writeJSON`/`writeError`). ✅
- §2.4 clean architecture, package layout, no stuttering names → Tasks 2–13. ✅
- §2.5 domain `Validate(maxLimit)` called inside the use-case → Tasks 3, 4. ✅
- §2.6 sentinel errors → HTTP codes (400/404/500) → Tasks 3, 9, 10. ✅
- §2.7 in-memory store, segregated interfaces, only-200 counted, tie-break-first, race test → Tasks 4, 6. ✅
- §2.8 routes `/fizzbuzz`, `/fizzbuzz/stats`, `/healthz` → Task 12. ✅
- §2.9 go-envconfig, timeouts, graceful shutdown → Tasks 8, 13, 14. ✅
- §2.10 slog JSON logger + minimal middleware (obs left "to refine") → Tasks 11, 13. ✅
- §2.11 in-memory rate-limit guard + Redis stub, 429 → Task 7, 11, 13. ✅
- §2.12 golangci-lint v2 config → Task 1. ✅
- Testing strategy (unit, race, httptest, black-box `_test` packages) → throughout. ✅

**Placeholder scan:** No "TBD/TODO" in implementation code. The `panic("implement me")` strings in `statstorer.Redis` and `ratelimiter.Redis` are deliberate, spec-mandated extensibility placeholders (§2.7, §2.11), not plan gaps. The `var _ = io.Discard` line in Task 9's test is flagged with a removal note.

**Type consistency:** Verified across tasks — `fizzbuzz.GenerateRequest/GenerateResponse/GetStatsRequest/GetStatsResponse`; `usecase.StatRecorder.Record`, `usecase.StatReader.MostFrequent`; `statstorer.InMemory` satisfies both; `server.RateLimiter.Allow`; `ratelimiter.InMemory.Allow`; handler constructors take `(*usecase.X, *slog.Logger)`; container exposes `GetLogger`, `HTTPHandler`, `GetHTTPServer`.
