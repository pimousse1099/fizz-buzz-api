# 0003. Response and Error Format

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

Both the fizz-buzz and stats endpoints must return structured responses. Error cases must also
carry enough information for clients to act on them. We need a consistent, minimal wire format.

## Decision

**Success — fizz-buzz:**
```json
["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz"]
```
A raw JSON array of strings with `Content-Type: application/json`. This is the most direct
representation of "a list of strings" as stated in the spec.

**Success — stats:**
```json
{"request":{"int1":3,"int2":5,"limit":100,"str1":"fizz","str2":"buzz"},"total_hits":42}
```

**Error:** a dedicated HTTP status code (400/404/429/500) plus a valid JSON body. A JSON string
is sufficient:
```json
"int1 must be a positive integer"
```

All responses (success and error) carry `Content-Type: application/json` so clients can always
parse the body unconditionally.

## Consequences

- The raw array is the simplest valid encoding of the spec output; no wrapper object required.
- Always-valid JSON on errors prevents clients from needing two parsing paths.
- A plain JSON string for errors is intentionally simple — no error-object schema to version or
  document. If richer error metadata is needed (error codes, field pointers), the format can be
  evolved to a JSON object without breaking the "always JSON" contract.
