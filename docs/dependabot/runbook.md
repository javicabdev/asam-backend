# Dependabot Operations Runbook

Last updated: 2026-05-24

This runbook covers the recurring operational patterns when handling Dependabot
PRs in this repository. The triage history lives alongside this file (e.g.
`triage-2026-05-24.md`).

## Known issue: Docker base image SHA pin regression

### Symptom
When Dependabot bumps the Go Docker base image (e.g. `golang:1.26.2 → 1.26.3`),
the generated PR replaces our existing digest pin

```dockerfile
FROM golang:1.26.2-alpine@sha256:<old-digest> AS builder
```

with a bare version tag:

```dockerfile
FROM golang:1.26.3-alpine
```

Merging that PR as-is regresses the repo's OpenSSF Scorecard
`Pinned-Dependencies` score from `10/10`. This is a known Dependabot behaviour
(it does not yet resolve registry digests when generating Docker PRs).

### Why we accept the regression window
- Go base image bumps are roughly monthly.
- The exposure window is the time between merging Dependabot's PR and pushing
  the digest-restoring commit — minutes if handled promptly.
- Building a custom workflow that resolves digests before opening the PR is
  over-engineering for the current bump frequency (YAGNI).

If the bump frequency or compliance pressure ever changes, revisit and
consider replacing Dependabot's `docker` ecosystem with a scheduled custom
workflow.

## Procedure: re-pin after a Docker base image bump

Execute immediately after merging a Dependabot Docker PR.

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

## Related documents

- [triage-2026-05-24.md](./triage-2026-05-24.md) — the cycle that established
  this runbook.
- [../../.github/dependabot.yml](../../.github/dependabot.yml) — current
  Dependabot configuration.
- [../CI-CD.md](../CI-CD.md) — overall CI/CD documentation.
