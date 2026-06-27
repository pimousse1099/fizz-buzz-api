# 0020. Pin CI Tool and Runner Versions (No Floating `latest`)

- **Status:** Accepted
- **Date:** 2026-06-27
- **Related:** [0012](0012-linting-golangci-lint.md), [0019](0019-container-and-ci-cd.md)

## Context

The CI workflows initially used floating tags: `runs-on: ubuntu-latest` on every job and
`golangci-lint-action` with `version: latest`. Floating `latest` makes builds non-reproducible —
the same commit can pass today and fail tomorrow with no code change — and couples the build to
upstream release timing. This is not hypothetical here: a `golangci-lint` `latest` built against an
older Go than our pinned toolchain (1.26) refused to load the config (`the Go language version used
to build golangci-lint is lower than the targeted Go version`). `ubuntu-latest` carries the same
risk: it silently migrates to a new major image (e.g. 22.04 → 24.04) on GitHub's schedule, which can
shift preinstalled tool versions under the build.

## Decision

Pin every external version the CI depends on; ban floating `latest` for build inputs.

- **Runners:** `ubuntu-24.04` (not `ubuntu-latest`) in all workflows (lint, test, build-pull-request,
  build-release, dependabot).
- **golangci-lint:** pinned to `v2.12.2` via a `GOLANGCI_LINT_VERSION` env var, not `latest`. The
  version is chosen to match the Go toolchain (see the context above).
- **Go toolchain:** already pinned via `GO_VERSION: "1.26"`.
- **GitHub Actions:** pinned by major version (`actions/checkout@v4`, `actions/setup-go@v5`,
  `golangci/golangci-lint-action@v7`, `docker/build-push-action@v6`, …) and kept current by
  Dependabot's `github-actions` ecosystem ([0019](0019-container-and-ci-cd.md)).

**Explicit exception — the Docker release image `:latest` tag.** `build-release` still publishes
`type=raw,value=latest`. That is a *consumer-facing release alias* ("the latest stable release"),
not a build input, so it does not affect reproducibility and is kept.

## Consequences

- CI is reproducible: a green run stays green until a pin is changed in a reviewable diff.
- Tool/runner upgrades become deliberate, visible PRs — Dependabot proposes Actions bumps; the
  runner image and golangci-lint pins are bumped by hand (a one-line change), which is also when the
  golangci-lint↔Go-version compatibility is re-checked.
- Small ongoing cost: pins must be advanced on purpose rather than drifting forward automatically —
  an acceptable trade for not being broken by an upstream release at an arbitrary time.
