# 0006. Statistics Store

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The stats endpoint must return the parameters of the most frequently requested fizz-buzz call
plus its hit count. This requires a shared, thread-safe counter that tracks requests across the
lifetime of the process. We need to decide on the data structure, concurrency strategy,
interface design, and tie-breaking behaviour.

## Decision

### Storage

In-memory counter: `map[fizzbuzz.GenerateRequest]int` guarded by a `sync.Mutex` in
`infrastructure/statstorer/in_memory.go`. Stats reset on restart (acceptable per spec).

### Interface segregation

Two interfaces, defined where consumed (dependency-inversion, see [0001](0001-clean-hexagonal-architecture.md)):

- `StatRecorder` (in `usecase/generate_fizzbuzz.go`) — write side:
  ```go
  RecordFizzBuzzStat(ctx context.Context, req fizzbuzz.GenerateRequest) error
  ```
- `StatReader` (in `usecase/get_fizzbuzz_stats.go`) — read side:
  ```go
  GetMostFrequentFizzbuzzRequest(ctx context.Context) (*fizzbuzz.GetStatsResponse, error)
  ```

The in-memory store implements both. A `redis.go` file (struct + `panic("implement me")`)
demonstrates how a durable backend plugs in without touching the use-cases.

### Only successful requests are counted

A request that fails validation (`400`) is not a real fizz-buzz request; only `200 OK` responses
increment the counter.

### Tie-breaking: first-to-max wins

When two parameter sets have equal hit counts, return the one that reached the maximum first.
Implemented in O(1) by memoising the current top entry and updating it only when a new count
**strictly exceeds** the current maximum — deterministic and test-friendly.

## Consequences

- In-memory is sufficient for the spec; the interface boundary and Redis stub show the
  extensibility expected of production code.
- Swapping to Redis (or any durable backend) requires no changes to the use-cases — only a new
  `infrastructure/statstorer/redis.go` implementation wired in the DI container.
- The segregated interfaces enforce the minimum capability each use-case actually needs (write vs
  read), making fakes in tests minimal and precise.
- Stats are lost on restart; this is a known, accepted trade-off.
