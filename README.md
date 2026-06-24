# fizz-buzz-api
implementation of a fizz-buzz REST ~~ful~~ API

test instructions

> Please share a Git repository containing your solution, including any setup instructions
required to run and evaluate it.
> 
> The original fizz-buzz consists in writing all numbers from 1 to 100, and just replacing all
multiples of 3 by "fizz", all multiples of 5 by "buzz", and all multiples of 15 by "fizzbuzz".
> 
> The output would look like this:
```"1,2,fizz,4,buzz,fizz,7,8,fizz,buzz,11,fizz,13,14,fizzbuzz,16,..."```.
> 
> 
> Your goal is to implement a web server that will expose a REST API endpoint that:
> - Accepts five parameters: three integers int1, int2 and limit, and two strings str1 and str2.
> - Returns a list of strings with numbers from 1 to limit, where: all multiples of int1 are
  replaced by str1, all multiples of int2 are replaced by str2, all multiples of int1 and int2 are
  replaced by str1str2.The server needs to be:
> - Ready for production
> - Easy to maintain by other developers
> - Bonus: add a statistics endpoint allowing users to
  know what the most frequent request has been. This endpoint should:
>   - Accept no parameter
>   - Return the parameters corresponding to the most used request, as well as the number of
    hits for this request
---

## Run & evaluate

Requirements: Go 1.26+.

```sh
# Run the server (required env vars must be provided)
ENV_TYPE=dev HTTP_ADDR=:8080 FIZZBUZZ_MAX_SEQUENCE_LENGTH=10000 LOG_LEVEL=info go run ./cmd

# Generate
curl 'http://localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz'

# Most frequent request (404 until at least one successful generation)
curl 'http://localhost:8080/fizzbuzz/stats'

# Health
curl http://localhost:8080/healthz
```

Configuration (environment variables), grouped by concern:

| Variable | Required | Default | Description |
|---|---|---|---|
| `ENV_TYPE` | yes | — | Deployment environment (e.g. `production`, `dev`) |
| `ENV_NAME` | no | — | Free-form instance label |
| `HTTP_ADDR` | yes | — | Listen address, e.g. `:8080` |
| `HTTP_READ_HEADER_TIMEOUT` | no | `2s` | Read-header timeout |
| `HTTP_WRITE_TIMEOUT` | no | `10s` | Write timeout |
| `HTTP_IDLE_TIMEOUT` | no | `120s` | Idle timeout |
| `HTTP_RATE_LIMIT_PER_SEC` | no | `50` | Per-instance rate limit (req/s) |
| `HTTP_RATE_LIMIT_BURST` | no | `100` | Rate-limit burst |
| `FIZZBUZZ_MAX_SEQUENCE_LENGTH` | yes | — | Upper bound for the `limit` parameter |
| `LOG_LEVEL` | yes | — | `debug` / `info` / `warn` / `error` |

Development:

```sh
make build   # go build ./...
make race    # go test -race ./...
make lint    # golangci-lint run
```

Architecture decisions: see `docs/architecture-decision-records/2026-06-23-fizz-buzz-api.md`.
