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

_(#142 / PR #144 fixed all three original holes. History kept on purpose — if a
symptom reappears, you'll recognise it.)_

- **gosec runs for real now — but know the failure mode.** The "Security Scan
  (SAST)" job runs gosec **on the runner** (Go from setup-go) and gates on findings
  that are HIGH severity **and** HIGH confidence. The full report (all severities)
  is still printed; only HIGH+HIGH fails the build. The current backlog (~58
  findings, mostly `test/seed` + `cmdtemp` + false positives) is tracked in issue
  #145. **Broken before #144:** it used the `securego/gosec` **Docker action**,
  whose image ships an older Go — with `GOTOOLCHAIN=local` it couldn't load a
  `go 1.26.x` module and silently scanned **0 files**, and `-no-fail` hid that its
  green tick meant nothing. If gosec ever reports 0 findings again, suspect the
  toolchain/loader, not clean code.
- **`govulncheck` runs in CI** (added in #142), no flags. Its default source mode
  fails (exit 3) only on **reachable** vulnerabilities and passes (exit 0) on vulns
  merely present but not called (e.g. GO-2026-5932, `x/crypto/openpgp`). A red
  govulncheck means our call graph is actually affected — worth acting on.
- **The Docker image IS built on every PR now** (job added in #142): build only,
  no push, Buildx + GHA cache. `cache-to` is scoped to `push` events so Dependabot
  / fork PRs (read-only token) don't 403 on cache export — they only read the cache
  warmed from `main`. A green CI now does exercise the Dockerfile.
