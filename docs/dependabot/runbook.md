# Dependabot Operations Runbook

Last updated: 2026-06-20

This runbook covers the recurring operational patterns when handling Dependabot
PRs in this repository. The triage history lives alongside this file (e.g.
`triage-2026-05-24.md`).

## Docker base image bumps (digests maintained by Dependabot)

### Current behaviour — verified PR #114 (2026-06)
Dependabot's `docker` ecosystem updates **both the version tag and the `@sha256` digest**, preserving the pin:

```dockerfile
- FROM golang:1.26.3-alpine@sha256:<old> AS builder
+ FROM golang:1.26.4-alpine@sha256:<new> AS builder
```

There is **no `Pinned-Dependencies` regression and no re-pin commit is needed**. Handle Docker PRs like `gomod`/`github-actions`: confirm the diff touches only the `FROM` line(s), then merge.

### Two things to know (and not "fix")
- **The pinned digest may lag the registry by a few days.** Docker Official Images use mutable tags rebuilt for base-OS patches, so by merge time the tag may point to a newer digest than the PR pins. This is expected — the PR's digest is the one CI validated. Do **not** manually refresh it; Dependabot opens a follow-up digest bump when the tag is rebuilt.
- **The `golang` image is the discarded builder stage.** The deployed image is the final `alpine:*` stage (multi-stage build copies only the static binary), so the builder's base-OS freshness never reaches production. Runtime base patches are governed by the final stage, not by `golang` bumps. What matters from a `golang` bump is the Go toolchain version.

## Fallback: re-pin only if a digest is ever stripped

Current Dependabot maintains digests (see above), so this is **not** the normal path. It applies only if a future Docker PR ever lands a bare `FROM golang:<v>` with no `@sha256` — the behaviour observed historically at the 1.26.2 → 1.26.3 bump, before the FROM lines carried digests. Steps 1–7 below are the fallback.

### 1. Sync local main

```bash
git checkout main && git pull --ff-only origin main
```

### 2. Identify affected Dockerfiles

```bash
find . -name 'Dockerfile*' -not -path './node_modules/*' -not -path './.git/*'
```

Current files: `Dockerfile`, `Dockerfile.dev`.

### 3. Resolve the new digests

For each base image variant in use, pull and read the digest. Example for the
two variants this repo uses:

```bash
docker pull golang:<new-version>-alpine
docker inspect --format='{{index .RepoDigests 0}}' golang:<new-version>-alpine

docker pull golang:<new-version>
docker inspect --format='{{index .RepoDigests 0}}' golang:<new-version>
```

Copy the `sha256:<digest>` portion of each output.

### 4. Re-pin the Dockerfiles

Edit each `FROM` line to restore the digest pin. Format:

```dockerfile
FROM golang:<version>-alpine@sha256:<digest> AS builder
FROM golang:<version>@sha256:<digest>
```

Match the existing file conventions — do not introduce new patterns.

### 5. Sanity build

```bash
docker build --no-cache -f Dockerfile . -t asam-backend:repin-test
docker image rm asam-backend:repin-test
```

### 6. Commit and push to main

Direct push to `main` is acceptable here (CI-only / security-pin change, no
business logic). Use the conventional commit format:

```
security(ci): re-pin Docker base images to <new-version> digests

Restores SHA digest pins after Dependabot PR #<num> merged with bare
version tags. Preserves OpenSSF Pinned-Dependencies: 10/10.
```

### 7. Verify

After CI passes:

- Check the next OpenSSF Scorecard run (or trigger it) to confirm
  `Pinned-Dependencies: 10/10`.
- Cross-check: `git log --oneline -1` should show the security commit on top.

## Other Dependabot ecosystems

For reference, the following ecosystems behave correctly and do not need this
runbook:

- `gomod` — Dependabot edits `go.mod`/`go.sum` cleanly; standard PR review.
- `github-actions` — Dependabot preserves `uses: foo/bar@<sha> # v<version>`
  pins correctly. Verified in PR #102 (2026-05-24 triage).
- `docker` — Dependabot updates the version tag and `@sha256` digest together, preserving the pin. Verified in PR #114 (2026-06).

## Related documents

- [triage-2026-05-24.md](./triage-2026-05-24.md) — the cycle that established
  this runbook.
- [../../.github/dependabot.yml](../../.github/dependabot.yml) — current
  Dependabot configuration.
- [../CI-CD.md](../CI-CD.md) — overall CI/CD documentation.
