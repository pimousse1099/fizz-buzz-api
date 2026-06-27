# 0012. Linting and Formatting — golangci-lint v2

- **Status:** Accepted
- **Date:** 2026-06-23

## Context

The service must be maintainable by the wider reezoback team. Code style, correctness checks,
and formatting must be enforced automatically. The team uses golangci-lint; we need a config that
is current (v2 schema, Go 1.26) and coherent with the house philosophy.

## Decision

Use **golangci-lint v2** with a `.golangci.yaml` that adopts the reezoback house philosophy:
`disable-all: true` plus an explicit curated enable list — scalable and future-proof.

The config is authored for the **v2 schema** (not a verbatim copy of the reezoback v1 file).

**Linters carried over from reezoback:**

| Category | Linters |
|---|---|
| Formatting | `gofumpt` (extra rules), `gci` (import sections: standard / default / `prefix(github.com/Pimousse1099)`) |
| Complexity | `cyclop`, `gocognit`, `funlen`, `nestif` |
| Error discipline | `errname`, `err113` (sentinel errors + wrapping) |
| Security | `gosec` |
| Global/init hygiene | `gochecknoglobals`, `gochecknoinits` |
| Test discipline | `testpackage` (black-box `_test` packages), `paralleltest`, `tparallel` |
| General quality | `revive`, `staticcheck`, `gocritic`, `godot`, `misspell` |
| Whitespace | `wsl`, `nlreturn`, `whitespace` |
| Magic numbers | `mnd` |
| slog consistency | `sloglint` |

**Adaptations for v2 / this service:**
- Renames: `gomnd` → `mnd`, `goerr113` → `err113`.
- Dropped (removed/obsolete in Go 1.22+): `execinquery`, `exportloopref`.
- Dropped (irrelevant to a stdlib HTTP service): `protogetter`, `rowserrcheck`, `sqlclosecheck`,
  `zerologlint`, and reezoback-specific `skip-dirs`.
- `gci` / `goimports` local prefix: `github.com/Pimousse1099`.

## Consequences

- The explicit enable list means new linters are opt-in — no surprise failures after golangci-lint
  upgrades.
- `sloglint` enforces consistent `log/slog` usage, complementing [0015](0015-request-logging-httplog.md).
- `err113` + `errname` enforce the sentinel-error pattern from [0005](0005-error-to-http-status-mapping.md).
- `testpackage` enforces black-box test packages, consistent with the testing strategy in
  [0001](0001-clean-hexagonal-architecture.md).
- Parity with reezoback linting means PRs from team members will feel familiar.
