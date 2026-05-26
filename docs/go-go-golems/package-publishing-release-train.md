# go-go-golems package publishing and dependency rollout playbook

This playbook captures the reusable operational workflow used during the logcopter rollout across Glazed, Clay, Geppetto, and Pinocchio. It is for publishing upstream go-go-golems packages, bumping downstream modules, validating against published dependencies, and merging PRs only after CI and Codex review are satisfied.

## Tooling in this repo

Scripts:

- `scripts/go-go-golems/00-pr-ready-check.sh` — one-shot PR readiness check.
- `scripts/go-go-golems/01-pr-ready-check.py` — GitHub GraphQL implementation for checks + Codex signals.
- `scripts/go-go-golems/02-trigger-codex-review.sh` — posts `@codex review`.
- `scripts/go-go-golems/03-watch-codex-reactions.py` — watches for Codex reaction transitions.
- `scripts/go-go-golems/04-wait-pr-ready.sh` — polls until the readiness check succeeds or times out.

Playbooks and snippets:

- `docs/go-go-golems/playbooks/logcopter-package-rollout-playbook.md` — detailed logcopter package-logger rollout guide.
- `docs/go-go-golems/playbooks/pr-readiness-check-scripts.md` — design notes for the readiness scripts.
- `examples/go-go-golems/Makefile.bump-go-go-golems.snippet.mk` — generic dependency bump target.
- `examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk` — dependency bump target that forces published-module resolution.

## Release train principle

Rollouts must follow dependency order. Do not bump and merge downstream repositories until the upstream repositories they need have been merged and published.

Typical order for core go-go-golems packages:

```text
logcopter -> glazed/clay -> geppetto -> pinocchio -> leaf applications
```

The exact order depends on the current `go.mod` graph. Inspect direct dependencies rather than relying on memory.

## Per-repository workflow

### 1. Land and publish the upstream package

In the upstream repository:

1. Ensure the PR is merged.
2. Ensure the release/tag/published module version exists.
3. Confirm what downstream should consume:

```bash
go list -m -versions github.com/go-go-golems/<upstream>
git fetch --tags
git describe origin/main --tags --always
```

If `origin/main` is ahead of the latest tag and downstream needs those commits, publish a new release before proceeding downstream.

### 2. Bump downstream go-go-golems dependencies

In the downstream repository, use the generic target:

```bash
make bump-go-go-golems
```

If there is any chance a local `go.work` can hide missing releases, use a `GOWORK=off` variant or run the equivalent commands manually:

```bash
deps="$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $1 }' go.mod | sort -u)"
for dep in $deps; do GOWORK=off go get "${dep}@latest"; done
GOWORK=off go mod tidy
```

Review the result:

```bash
git diff -- go.mod go.sum
go list -m github.com/go-go-golems/...
```

### 3. Validate without local workspace assumptions

Prefer `GOWORK=off` for smoke tests that prove the published dependency graph works:

```bash
GOWORK=off go test ./...
```

For logcopter-enabled repositories, run the non-mutating generated-file freshness check before any mutating generation command:

```bash
make logcopter-check
# only when intentionally refreshing generated files:
# go generate ./...
```

Run repo-specific checks too, for example:

```bash
make test
make lint
make ci
```

### 4. Commit and push

Commit only the intended dependency changes and related generated/check artifacts:

```bash
git status --short
git add go.mod go.sum
git commit -m "Bump go-go-golems dependencies"
git push <remote> <branch>
```

### 5. Trigger or wait for Codex review

If a PR needs a fresh Codex review after the push:

```bash
scripts/go-go-golems/02-trigger-codex-review.sh https://github.com/go-go-golems/<repo>/pull/<n>
```

Wait for readiness:

```bash
scripts/go-go-golems/04-wait-pr-ready.sh https://github.com/go-go-golems/<repo>/pull/<n> 30 1800
```

A PR is considered ready when:

- status checks exist;
- every status check is completed successfully, skipped, or neutral;
- a Codex signal exists;
- the latest Codex signal has a thumbs-up reaction or a satisfied body such as `Didn't find any major issues. :+1:`;
- the latest Codex signal has no `EYES` reaction;
- the latest Codex-authored body is empty, benign, or satisfied rather than substantive review feedback.

### 6. Merge only after readiness succeeds

After the wait script exits successfully, merge using the normal repository policy:

```bash
gh pr merge <n> --squash --delete-branch=false
```

If the PR touches `.github/workflows/*`, the GitHub CLI token needs `workflow` scope. If merge fails with a workflow-scope error, refresh auth:

```bash
gh auth refresh -h github.com -s workflow
```

Then retry the merge.

## Common gotchas

- Local `go.work` can hide missing published upstream symbols. Use `GOWORK=off` for downstream readiness checks.
- `go generate ./...` is mutating. For generated-file drift checks, run the non-mutating checker first.
- A merged upstream PR is not the same as a published upstream module version. Check tags/module versions before bumping downstream.
- Codex `EYES` reactions mean review may still be running; do not merge until the readiness checker accepts the latest signal.
- If Codex leaves substantive review text, treat the PR as not ready even when Actions are green.
