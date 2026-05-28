---
title: INFRA-004 Intern Release Train Playbook
status: active
type: reference
created: 2026-05-28
tags:
  - infra
  - release-train
  - logcopter
  - github-actions
  - codex
  - xgoja
---

# INFRA-004 Intern Release Train Playbook

This playbook is for continuing the INFRA-004 rollout across Go-Go-Golems repositories. It assumes you are working in the Go-Go-Golems monorepo workspace layout and using the INFRA-004 tracker as the source of truth.

The goal is to keep the release train moving without losing track of repository state. Work is organized as small PRs, one PR per repository, with local validation, Codex review, GitHub checks, merge-commit merges, main-branch verification, and optional release tagging.

## Current ticket and tracker

Ticket directory:

```bash
T=/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos
```

Dashboard:

```text
tmux session: infra004-dashboard
URL: http://127.0.0.1:8765/
```

Tracker commands:

```bash
$T/scripts/02-rollout-tracker.py summary
$T/scripts/02-rollout-tracker.py list --batch B5
$T/scripts/02-rollout-tracker.py list --state pr_open
$T/scripts/02-rollout-tracker.py list --state blocked
```

The tracker is the operational source of truth. Update it whenever you open a PR, record a validation command, find a blocker, merge a PR, verify main actions, or push a release tag.

## Important rules

1. Never push directly to `main`.
2. One focused PR per repository.
3. Merge only with merge commits:

```bash
gh pr merge <PR_URL> --merge --delete-branch
```

4. Do not squash.
5. Do not merge until `ggg pr ready` reports `state: ready`.
6. Treat `no_runs` as OK for main verification when the repository has no relevant workflow runs.
7. Do not implement xgoja/API provider bindings without confirming the intended API surface.
8. If a change starts to require design decisions, stop and ask for help.

## Current high-level status

B5 logcopter prerequisites have been merged and main-verified for several repositories:

```text
cozodb-goja
openai-app-server
go-go-gepa
goja-github-actions
vm-system
scraper
```

Open or risky work at the time this playbook was written:

```text
smailnail PR #4: open, Codex satisfied, but Go Vulnerability Check was failing.
```

Skipped by explicit user request:

```text
common-sense
plunger
biberon
bucheron
ecrivain
```

Blocked/needs decision:

```text
voyage: archived/read-only and has pre-existing compile failures.
bubble-table, raza, terraform-provider-stytch-b2b: ownership/external-module intent needs confirmation.
barbar: root-only package main; mechanical logcopter rollout needs manual decision.
```

## The work lanes

Use three lanes in parallel.

### Lane A: Release/tag safe merged repositories

Tag repositories that are already:

1. merged,
2. `main_actions_verified`,
3. release preflight clean or warning explicitly accepted,
4. useful as dependencies for the next work.

Recommended first release order:

```text
sanitize
go-sqlite-regexp
cozodb-goja
openai-app-server
go-go-gepa
goja-github-actions
```

Hold these until their risk is resolved:

```text
scraper: release preflight warnings about frontend/pnpm setup.
smailnail: do not release until PR #4 is merged and main-verified.
vm-system: no GoReleaser config; tag only if a module consumer needs it.
```

### Lane B: Fix open or blocked PRs

Work on PRs that are not `ready`. Typical commands:

```bash
ggg pr ready <PR_URL> --findings --output json

gh pr checks <PR_NUMBER> --watch=false

gh run list -R go-go-golems/<repo> --branch <branch> --limit 10 --json databaseId,name,status,conclusion,url,headSha
```

Fix only the current blocker. Do not expand the PR into unrelated cleanup.

### Lane C: Start B5 xgoja or B5 prerequisites

After logcopter prerequisites are merged, B5 work can move toward xgoja/API work. Do not guess the API. Start with analysis and confirm intent before implementing provider bindings.

B5 repositories:

```text
cozodb-goja
go-go-gepa
go-go-goja
go-minitrace
goja-github-actions
openai-app-server
pinocchio
scraper
smailnail
vm-system
workspace-manager
```

The xgoja-only repositories are good candidates for analysis work once dependency releases are in progress:

```text
go-go-goja
go-minitrace
workspace-manager
```

`pinocchio` depends on `sanitize`, so prefer to tag `sanitize` before dependency-bump work there.

## How to check whether a PR can be merged

Run:

```bash
ggg pr ready <PR_URL> --findings --output json
```

You may merge only if the output includes:

```json
"state": "ready",
"terminal": true
```

and every finding that matters is `ok: true`.

Ready means:

- merge state is clean,
- checks are successful or no checks are configured,
- Codex is satisfied/benign.

If Codex comments are stale for an older head commit, trigger a fresh review:

```bash
ggg pr codex-trigger <PR_URL> --wait-for-auto 30s
```

If Codex appears stuck, wait a few minutes and rerun readiness. If still stuck and no actual review comments exist, ask for help before merging.

## How to merge and verify main

When a PR is ready:

```bash
REPO=<repo>
PR=<pr-url>

$T/scripts/02-rollout-tracker.py update-repo "$REPO" \
  --state ready \
  --event 'PR readiness green; about to merge.'

gh pr merge "$PR" --merge --delete-branch

MERGE=$(gh pr view "$PR" --json mergeCommit --jq .mergeCommit.oid)

$T/scripts/02-rollout-tracker.py merge "$REPO" \
  --sha "$MERGE" \
  --url "$PR"

ggg run status \
  --repo go-go-golems/$REPO \
  --branch main \
  --sha "$MERGE" \
  --ignore-workflow "Secret Scanning" \
  --watch \
  --output json

$T/scripts/02-rollout-tracker.py update-repo "$REPO" \
  --state main_actions_verified \
  --merge-sha "$MERGE" \
  --action-status checked \
  --event 'Main branch actions watched after merge.'
```

If `ggg run status` returns `no_runs` with `ok: true`, that is acceptable.

## How to open batched logcopter prerequisite PRs

This is the standard pattern for B5 logcopter prerequisite repositories.

```bash
REPO=<repo>
cd /home/manuel/code/wesen/go-go-golems/$REPO

git fetch origin main
git switch main
git branch --set-upstream-to=origin/main main || true
git pull --ff-only

git switch -c infra/b5-logcopter-baseline

GOWORK=off go get github.com/go-go-golems/logcopter@latest
GOWORK=off go get -tool github.com/go-go-golems/logcopter/cmd/logcopter-gen@latest
```

Create `logcopter_generate.go`. Use `zlog` when the repository already imports stdlib/global `log` heavily and you want to avoid name collisions:

```go
package REPO_PACKAGE_NAME

//go:generate go tool logcopter-gen -include-main -var zlog -area-prefix go-go-golems.REPO -strip-prefix github.com/go-go-golems/REPO ./cmd/... ./pkg/...
```

Run generator:

```bash
GOWORK=off go tool logcopter-gen \
  -include-main \
  -var zlog \
  -area-prefix go-go-golems.$REPO \
  -strip-prefix github.com/go-go-golems/$REPO \
  ./cmd/... ./pkg/...
```

Add Makefile targets:

```make
.PHONY: logcopter-generate
logcopter-generate:
	GOWORK=off go tool logcopter-gen -include-main -var zlog -area-prefix go-go-golems.REPO -strip-prefix github.com/go-go-golems/REPO ./cmd/... ./pkg/...

.PHONY: logcopter-check
logcopter-check:
	GOWORK=off go tool logcopter-gen -include-main -var zlog -area-prefix go-go-golems.REPO -strip-prefix github.com/go-go-golems/REPO -check ./cmd/... ./pkg/...
```

Then:

```bash
gofmt -w logcopter_generate.go $(find cmd pkg -name logcopter.go 2>/dev/null || true)
GOWORK=off go mod tidy
make logcopter-check
GOWORK=off go test ./...
ggg release preflight --repo . --output json || true
```

Open PR:

```bash
git add .
git commit -m "Add B5 logcopter baseline"
git push -u origin infra/b5-logcopter-baseline

PR=$(gh pr create \
  --title "Add B5 logcopter baseline" \
  --body "## Summary
- add generated logcopter package loggers as a B5 prerequisite
- register logcopter-gen as a Go tool
- add logcopter generation/check targets

## Validation
- make logcopter-check
- GOWORK=off go test ./...
- ggg release preflight --repo . --output json

INFRA-004 B5 predependency work before xgoja/API changes." \
  --base main \
  --head infra/b5-logcopter-baseline)
```

Update tracker:

```bash
$T/scripts/02-rollout-tracker.py update-repo "$REPO" \
  --state pr_open \
  --branch infra/b5-logcopter-baseline \
  --pr-url "$PR" \
  --head-sha $(git rev-parse --short HEAD) \
  --event 'Opened batched B5 logcopter prerequisite PR before xgoja/API work.'

$T/scripts/02-rollout-tracker.py validation "$REPO" \
  --command 'make logcopter-check' \
  --status pass

$T/scripts/02-rollout-tracker.py validation "$REPO" \
  --command 'GOWORK=off go test ./...' \
  --status pass
```

Trigger Codex but do not wait on each PR before moving to the next repository:

```bash
ggg pr codex-trigger "$PR" --wait-for-auto 5s || true
```

## How to do batched watching

After opening several PRs, collect their URLs and run one watch pass:

```bash
for pr in \
  https://github.com/go-go-golems/repo-a/pull/N \
  https://github.com/go-go-golems/repo-b/pull/N \
  https://github.com/go-go-golems/repo-c/pull/N
 do
  echo "=== $pr ==="
  ggg pr ready "$pr" --findings --output json || true
done
```

Merge only those that return `ready`. Leave the others open and fix in the next batch.

## Release/tagging playbook

Only tag after:

- PR merged,
- `main_actions_verified`,
- release preflight clean or warnings accepted.

Recommended safe release order:

```text
sanitize
go-sqlite-regexp
cozodb-goja
openai-app-server
go-go-gepa
goja-github-actions
```

Per repository:

```bash
REPO=<repo>
cd /home/manuel/code/wesen/go-go-golems/$REPO

git switch main
git pull --ff-only

ggg release preflight --repo . --output json

TAG=$(svu patch)
git tag "$TAG"
git push origin "$TAG"

$T/scripts/02-rollout-tracker.py release "$REPO" \
  --tag "$TAG" \
  --release-url "https://github.com/go-go-golems/$REPO/releases/tag/$TAG"
```

Then watch release actions in parallel:

```bash
for repo in sanitize go-sqlite-regexp cozodb-goja openai-app-server go-go-gepa goja-github-actions; do
  echo "=== $repo ==="
  gh run list -R go-go-golems/$repo --limit 5
done
```

If a release fails, do not delete tags unless explicitly instructed. Record the failure in the tracker and ask for help if the fix is not obvious.

## Edge cases and what to do

### No status checks found

If `ggg pr ready` says no status checks found and all other gates pass, that is OK.

If main verification returns:

```json
"state": "no_runs",
"ok": true
```

record `main_actions_verified`.

### Codex has stale feedback

If the finding says Codex feedback is stale for an older head commit:

```bash
ggg pr codex-trigger <PR_URL> --wait-for-auto 30s
```

Wait and rerun readiness.

### Codex has current review comments

Fix the comment if it is mechanical and obviously correct. Examples:

- generated target accidentally runs `go generate ./...` and triggers frontend generation,
- logger area is changed unintentionally,
- missing logcopter configuration after swapping global log imports.

Ask for help if the comment asks for design/API decisions or changes the intended behavior.

### Go Vulnerability Check fails on Go standard library vulnerabilities

This often means govulncheck is running with an unpatched Go toolchain. Prefer fixing the workflow, not forcing every developer through a patch-level `toolchain` directive in `go.mod`.

Example fix:

```yaml
- name: Set up Go
  uses: actions/setup-go@v6
  with:
    go-version: '1.25.10'
```

Use this only for the govulncheck job unless the whole workflow truly needs the patch toolchain.

### Repository requires build tags

Some repositories intentionally fail without build tags. `smailnail` requires `sqlite_fts5` or `fts5` for mirror code.

Use:

```bash
make test
```

instead of raw:

```bash
go test ./...
```

because `make test` supplies the required tags.

For logcopter generation in such repositories, set `GOFLAGS` in the Makefile target:

```make
logcopter-generate:
	GOWORK=off GOFLAGS="-tags=$(SQLITE_TAGS)" go tool logcopter-gen ...

logcopter-check:
	GOWORK=off GOFLAGS="-tags=$(SQLITE_TAGS)" go tool logcopter-gen ... -check ...
```

### Existing `log` imports conflict with generated logger

Use generated logger variable `zlog`:

```bash
go tool logcopter-gen -var zlog ...
```

This avoids changing existing stdlib/global `log` behavior in large repositories. Only replace existing `log` calls if the rollout requires it and the change is straightforward.

### The repository has `package main`

By default, ensure generation includes main packages:

```bash
logcopter-gen -include-main ...
```

If a root-only `package main` repository does not produce meaningful package loggers, ask for help before inventing a structure.

### Release preflight has warnings

Warnings can be accepted if they are known and unrelated to the current baseline. Errors should usually block tagging.

Record warnings explicitly:

```bash
$T/scripts/02-rollout-tracker.py validation <repo> \
  --command 'ggg release preflight --repo . --output json' \
  --status warn \
  --note 'pre-existing release warning; not introduced by logcopter baseline'
```

Do not mark warnings as `pass` unless the command actually returned `ok: true`.

### Release preflight has errors

Do not tag unless instructed. Either fix the release configuration in a separate focused PR or ask for help.

### Publish-image workflow fails on pull requests

If a publish-image workflow is incorrectly running on PRs and failing before jobs start, consider restricting it to `push`/`workflow_dispatch` if the repo does not need image builds on PRs.

Do not make this change silently for production-critical repositories. Ask if unsure.

### Repository is archived or push is rejected

Stop and mark blocked. Do not try to work around repository permissions.

```bash
$T/scripts/02-rollout-tracker.py update-repo <repo> \
  --state blocked \
  --event 'Repository archived/read-only; push rejected.'
```

### xgoja/API work

Do not implement provider bindings until API intent is confirmed. For xgoja tasks, start with a short analysis PR or design note if needed. Ask for help before changing public API shape.

## When to ask for help

Ask for help when any of these happen:

- a Codex comment requires an API/design decision,
- a release preflight error involves GoReleaser, cgo, cross-compilation, signing, or publishing,
- a workflow failure is unrelated to your PR and not clearly pre-existing,
- a repo has external ownership or non-Go-Go-Golems module path concerns,
- a repository has large generated/frontend artifacts and `go generate` changes many files,
- tests pass locally but fail in CI for reasons you cannot reproduce within 15 minutes,
- you are tempted to push directly to `main`.

## Final checklist before handing back

At the end of your work session, run:

```bash
$T/scripts/02-rollout-tracker.py summary
$T/scripts/02-rollout-tracker.py list --state pr_open
$T/scripts/02-rollout-tracker.py list --state blocked
```

Then report:

- PRs opened,
- PRs merged,
- main merge SHAs verified,
- release tags pushed,
- blockers and exact failing checks,
- what should be done next.
