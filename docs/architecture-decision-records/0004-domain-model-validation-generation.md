# 0004. Domain Model, Validation, and Generation Placement

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The fizz-buzz business objects need a home. The generation algorithm and input validation rules
must be placed in the layer that owns them — and that placement has consequences for testability,
cohesion, and dependency direction.

## Decision

### Domain model (`domain/fizzbuzz`)

The domain package holds **data structures** and **validation** only — no generation logic:

- `GenerateRequest{Int1, Int2, Limit int; Str1, Str2 string}` — value object, used as a map key
  in the stat store (therefore comparable; no pointer fields).
- `GenerateResponse{Result []string}`
- `GetStatsResponse{Request GenerateRequest; TotalHits int}`
- Sentinel errors in `error.go` (see [0005](0005-error-to-http-status-mapping.md)).

### Validation lives in the domain, triggered by the use-case

`func (r GenerateRequest) Validate(maxLimit int) error` (in `validator.go`) encodes the business
invariants:

- `int1 > 0` and `int2 > 0` (reject non-positive).
- `1 <= limit <= maxLimit` (`maxLimit` comes from config, default 10 000) — bounds response size
  to prevent memory exhaustion.
- `str1` and `str2` non-empty and `<= 100` characters.

The use-case calls `req.Validate(maxLimit)` at the very start of `Execute()`, so invariants are
enforced regardless of caller. The handler stays thin: parse query params → build
`GenerateRequest` → call use-case.

**Value receiver:** `Validate` uses a value receiver. `GenerateRequest` is a small (~56-byte)
immutable value object used as a map key. A value receiver copies a few bytes on the stack with
no allocation; a pointer receiver buys nothing here and can force the struct to escape to the
heap. Pointer receivers are reserved for mutation or large structs — neither applies.

### Generation logic lives in the use-case

The domain carries **no `Generate()` method**. The fizz-buzz loop lives in
`usecase.GenerateFizzBuzz.Execute`:

```
validate → generate (loop + switch in Execute) → record stats
```

This keeps the domain as a pure data/validation package, consistent with the layering rule: the
domain does not know which use-cases exist.

## Consequences

- The domain depends on nothing and is trivially portable and fast to test.
- Validation rules are co-located with the data they constrain (domain package), not scattered
  across handlers.
- The use-case is the single authoritative place that orchestrates validate → generate → record;
  no handler can bypass validation.
- If validation ever needs injected dependencies (e.g. a blocklist lookup), it can be refactored
  into a dedicated validator type without changing the use-case's call site.
