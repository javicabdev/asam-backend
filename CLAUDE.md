# CLAUDE.md

Guidance for AI assistants (and humans) working in this repo. Non-obvious,
durable constraints only — the kind of thing that causes a silent regression or
wastes a review if you don't know it.

## Dependency invariants

- **`gorm.io/gorm` is pinned to 1.31.0.** 1.31.1 breaks Preload
  (go-gorm/gorm#7686). It is deliberately ignored in `.github/dependabot.yml`.
  Do not bump it.
- **Floor for `google.golang.org/grpc`: >= 1.82.1** (GHSA-hrxh-6v49-42gf). Any
  PR that lowers it is a security regression, not a bump.
- **The Go version lives in 5 production locations and they move TOGETHER:**
  `go.mod`, `.github/workflows/ci.yml` (`GO_VERSION`), `Dockerfile` (tag AND
  digest), `.github/workflows/release.yml`, `.github/workflows/cloud-run-deploy.yml`.
  Non-production (kept separate): `Dockerfile.dev`, `.github/workflows/examples/`.
- **In the Dockerfile the digest wins over the tag.** Changing only the tag
  leaves the build silently using the old image. Update both, and resolve the
  digest from the multi-arch index (not a single platform).
- **`gqlgen` is pinned in TWO places:** `go.mod` and the
  `go install github.com/99designs/gqlgen@vX` in the `Dockerfile`. If they drift,
  code gets generated with one version and compiled with another.
- **`lib/pq` is a direct require that nobody imports,** so every `go mod tidy`
  wants to relocate it (direct <-> indirect). Revert that noise until PR #122
  lands.

## CI caveats

- **The "Security Scan (SAST)" job runs gosec with `-no-fail`: it CANNOT fail.**
  Its green tick is not evidence of anything when reviewing a PR.
- **`govulncheck` is not in CI.** Run it by hand before treating a Dependabot
  alert as urgent — it distinguishes "outdated" from "actually affects us"
  (reachability).
- **The Docker image is not built on PRs** (only `release.yml` on tag push and
  the manual deploy). A green CI does NOT mean the Dockerfile works.
