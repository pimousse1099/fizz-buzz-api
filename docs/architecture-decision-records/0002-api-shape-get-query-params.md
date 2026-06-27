# 0002. API Shape — GET with Query Parameters

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The fizz-buzz main endpoint accepts five parameters (`int1`, `int2`, `limit`, `str1`, `str2`).
We need to decide how to expose them over HTTP.

## Decision

Expose the main endpoint as:

```
GET /fizzbuzz?int1=3&int2=5&limit=100&str1=fizz&str2=buzz
```

Only `GET` is supported (no `POST`). The statistics endpoint is a sub-resource:

```
GET /fizzbuzz/stats
```

## Consequences

- `GET` is semantically correct for a pure read/compute operation; it is RESTful, cacheable, and
  trivially testable with curl or a browser.
- Restricting to `GET` keeps the validation and test surface small.
- Query parameters are the natural encoding for optional/named inputs that do not identify a
  resource by path.
- No request body parsing, no `Content-Type` negotiation on the input side.
