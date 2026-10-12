# CLAUDE.md

Guidance for AI assistants (and humans) working in this repo. Non-obvious,
durable constraints only — the kind of thing that causes a silent regression or
wastes a review if you don't know it.

## Dependency invariants

- **`gorm.io/gorm`: never 1.31.1** — it breaks Preload upstream
  (go-gorm/gorm#7686) and is ignored in `.github/dependabot.yml`. We were
  pinned to 1.31.0 until 1.31.2 was verified (with `gorm.io/driver/postgres`
  1.6.2, which requires gorm >= 1.31.2). Our Preload patterns (belongs-to,
  polymorphic has-many, nested, `Select`+`Joins`+`Preload`) are guarded by
  `test/integration/preload_relations_test.go`: a gorm bump that turns it red
  must not be merged. Extend that test when you add a new Preload shape.
- **Floor for `google.golang.org/grpc`: >= 1.83.2** (GHSA-vp52-pcj8-j9qc,
  GHSA-2v4p-qf9q-27wj; previous floor 1.82.1 for GHSA-hrxh-6v49-42gf). Any
  PR that lowers it is a security regression, not a bump.
- **The Go version lives in 5 production locations and they move TOGETHER:**
  `go.mod`, `.github/workflows/ci.yml` (`GO_VERSION`), `Dockerfile` (tag AND
  digest), `.github/workflows/release.yml`, `.github/workflows/cloud-run-deploy.yml`.
  Non-production (kept separate): `Dockerfile.dev`, `.github/workflows/examples/`.
  Dependabot **ignores the `golang` Docker image on purpose** (see
  `.github/dependabot.yml`): it can only bump the Dockerfiles, which silently
  desyncs the other locations. A Go bump is a manual PR touching all 5 locations
  (in the Dockerfile: the tag AND the multi-arch digest).
  **A Go minor bump is gated by the tools, not just `go build`.** gqlgen, gosec
  and golangci-lint are built with the new Go but analyse code through the
  `golang.org/x/tools` they pin, which must read the new export data format, so
  run all three under `GOTOOLCHAIN=go1.X.Y` before bumping. Go 1.27 needs
  x/tools >= v0.50.0. Checked on go1.27.2, every pinned tool failed:
  - gqlgen v0.17.95 generated broken code (`undefined: invalid`);
  - gosec v2.29.0 failed to load 68 files *and still exited green*;
  - golangci-lint v2.5.0 failed `typecheck`; v2.14.0 works.
- **In the Dockerfile the digest wins over the tag.** Changing only the tag
  leaves the build silently using the old image. Update both, and resolve the
  digest from the multi-arch index (not a single platform).
- **`gqlgen` is pinned in SIX places and they must all match:** `go.mod` (the
  runtime library), `Dockerfile` and `Dockerfile.dev` (`go install …@vX`), the
  `go run/install …@vX generate` steps in `.github/workflows/ci.yml` **and**
  `.github/workflows/release.yml`, and the `Makefile` (`generate` and `tools`
  targets — it sat at 0.17.91 while the rest was on 0.17.93 until the 0.17.95
  bump). **Tool version and runtime version must be the
  same** — if they drift, the committed `generated.go` does NOT compile against the
  other version. It's not a warning, it's a build break, e.g.
  `unknown field Label in struct literal of type graphql.DeferredGroup`. (ci.yml /
  release.yml were stuck at 0.17.73 while the runtime moved to 0.17.91–0.17.93 —
  twenty versions behind; the repo compiling was luck, not design. Fixed in the
  0.17.93 sync PR.) When bumping gqlgen: change all six, then `gqlgen generate`
  and commit the regenerated code. Dependabot only bumps `go.mod`, so close its
  gqlgen PRs and do the bump by hand.
- **The only Postgres driver is `pgx/v5`** — GORM uses it, and `cmd/migrate`
  uses golang-migrate's `database/pgx/v5` driver, so migration URLs use the
  `pgx5://` scheme (a `postgres://` URL fails with "unknown driver"). `lib/pq`
  was removed in PR #122; do not reintroduce it. `go.sum` may still list it as
  part of other deps' module graph — that is expected, nothing compiles it.

## CI caveats

_(#142 / PR #144 fixed all three original holes. History kept on purpose — if a
symptom reappears, you'll recognise it.)_

- **gosec runs for real now — but know the failure mode.** The "Security Scan
  (SAST)" job runs gosec **on the runner** (Go from setup-go), excluding generated
  code, and fails on any **HIGH severity** finding (any confidence).
  The full report (all severities) is still printed. The HIGH backlog was triaged to
  zero in #145; what remains is MEDIUM/LOW and reviewed. Justify a false positive
  inline with `// #nosec G101 -- why`, and the `#nosec` must **start** the comment:
  gosec ignores it after `//nolint:gosec //`, though golangci honours plain `#nosec`.
  **Broken before #144:** it used the `securego/gosec` **Docker action**,
  whose image ships an older Go — with `GOTOOLCHAIN=local` it couldn't load a
  `go 1.26.x` module and silently scanned **0 files**, and `-no-fail` hid that its
  green tick meant nothing. A separate gate step now fails when gosec scans 0
  files **or** reports load errors (`"Golang errors"` in its JSON). Don't trust
  the finding count alone: on go1.27, gosec failed to load 68 files yet still
  reported 2 findings over "117 files".
- **`govulncheck` runs in CI** (added in #142), no flags. Its default source mode
  fails (exit 3) only on **reachable** vulnerabilities and passes (exit 0) on vulns
  merely present but not called (e.g. GO-2026-5932, `x/crypto/openpgp`). A red
  govulncheck means our call graph is actually affected — worth acting on.
- **The Docker image IS built on every PR now** (job added in #142): build only,
  no push, Buildx + GHA cache. `cache-to` is scoped to `push` events so Dependabot
  / fork PRs (read-only token) don't 403 on cache export — they only read the cache
  warmed from `main`. A green CI now does exercise the Dockerfile.
