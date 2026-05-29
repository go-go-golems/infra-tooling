---
Title: Release Train Handoff, 2026-05-29
Ticket: INFRA-004
Status: active
Topics:
  - automation
  - release
  - github
  - logcopter
  - docsctl
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite
    Note: Source-of-truth rollout and release-train tracker database.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md
    Note: Chronological diary through release-train Steps 16-20.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/analysis/02-current-status-release-order-and-rollout-lessons.md
    Note: Long-form status report and release-order rationale.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
    Note: SQLite tracker CLI and dashboard server.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/release-train-20260529-layer2/releases.tsv
    Note: Release evidence for Oak, Refactorio, Jesus, and Smailnail.
ExternalSources: []
Summary: End-of-day handoff for resuming the INFRA-004 dependency-ordered release train after Layer 1a and selected Layer 2/3 releases.
LastUpdated: 2026-05-29T19:25:00-04:00
WhatFor: Use this as the restart checklist for next week’s release-train continuation.
WhenToUse: Before creating any more tags, dependency-bump commits, or tracker updates for INFRA-004.
---

# Release Train Handoff, 2026-05-29

This handoff captures where the INFRA-004 release train stopped at end of day and how to resume next week without replaying the whole conversation. It assumes the reader has repository access under `/home/manuel/code/wesen/go-go-golems` and is working from the infra tracker repository at `/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling`.

## 1. High-Level State

The PR rollout phase is complete: no tracked INFRA-004 PRs remain open. The current work is the dependency-ordered release train: tag upstream modules first, bump those tags into downstream modules, validate locally and in GitHub Actions, then create downstream releases and update the SQLite tracker.

The tracker DB remains the source of truth:

```text
/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite
```

Dashboard server, if still running:

```text
http://127.0.0.1:8765/
http://127.0.0.1:8765/release-train
http://127.0.0.1:8765/bumps
http://127.0.0.1:8765/health
http://127.0.0.1:8765/issues
http://127.0.0.1:8765/blocked
```

Restart dashboard if needed:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers
tmux kill-session -t infra004-dashboard 2>/dev/null || true
tmux new-session -d -s infra004-dashboard \
  "cd /home/manuel/workspaces/2026-05-24/add-js-providers && python3 infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py dashboard --port 8765"
```

## 2. What Was Released Today

### Layer 1a upstreams released

These were tagged first because they unblock downstream Layer 2 repositories:

| Repository | Release | URL |
|---|---:|---|
| `bobatea` | `v0.1.6` | https://github.com/go-go-golems/bobatea/releases/tag/v0.1.6 |
| `go-emrichen` | `v0.0.11` | https://github.com/go-go-golems/go-emrichen/releases/tag/v0.0.11 |
| `go-go-mcp` | `v0.0.19` | https://github.com/go-go-golems/go-go-mcp/releases/tag/v0.0.19 |
| `go-go-os-backend` | `v0.0.6` | https://github.com/go-go-golems/go-go-os-backend/releases/tag/v0.0.6 |
| `parka` | `v0.6.3` | https://github.com/go-go-golems/parka/releases/tag/v0.6.3 |
| `plz-confirm` | `v0.0.5` | https://github.com/go-go-golems/plz-confirm/releases/tag/v0.0.5 |
| `sessionstream` | `v0.0.6` | https://github.com/go-go-golems/sessionstream/releases/tag/v0.0.6 |
| `uhoh` | `v0.0.9` | https://github.com/go-go-golems/uhoh/releases/tag/v0.0.9 |

### Layer 2 / Layer 3 released after downstream bumps

| Repository | Bump / change | Release | Notes |
|---|---|---:|---|
| `oak` | `bobatea v0.1.6`; migrated old Glazed `layers`/`parameters` API to current `schema`/`fields`/`values`; updated bobatea REPL API | `v0.5.3` | Release URL: https://github.com/go-go-golems/oak/releases/tag/v0.5.3 |
| `refactorio` | `oak v0.5.3` | `v0.0.1` | Release URL: https://github.com/go-go-golems/refactorio/releases/tag/v0.0.1 |
| `jesus` | `go-go-mcp v0.0.19`; transitive `geppetto v0.10.17` and OpenTelemetry `v1.42.0` | `v0.0.1` | Release URL: https://github.com/go-go-golems/jesus/releases/tag/v0.0.1 |
| `smailnail` | `go-go-mcp v0.0.19` | `v0.0.1` | Release URL: https://github.com/go-go-golems/smailnail/releases/tag/v0.0.1 |

Tracker evidence file:

```text
ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/release-train-20260529-layer2/releases.tsv
```

## 3. Important Blocker: `zine-layout`

`zine-layout` was bumped from `go-emrichen v0.0.10` to `v0.0.11` and pushed to `main`:

```text
/home/manuel/code/wesen/go-go-golems/zine-layout
a8b2cbaca50e55b92644cfea8f374e0fa28f79b0 — Bump go-emrichen to v0.0.11
```

Local validation passed:

```bash
GOWORK=off go get github.com/go-go-golems/go-emrichen@v0.0.11
GOWORK=off go mod tidy
make logcopter-check
make glazed-lint
GOWORK=off go test ./...
GOWORK=off go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --timeout=5m
```

GitHub Actions result:

- `golang-pipeline`: success
- `golangci-lint`: success
- `Secret Scanning`: failure, recurring unrelated gate
- `Dependency Scanning`: **failure**

Dependency scanning run:

```text
https://github.com/go-go-golems/zine-layout/actions/runs/26655226837
```

Do **not** release `zine-layout` until this is triaged. The failure is a broad gosec backlog, not a simple patch-bump problem. Representative findings:

- `G115` integer conversion findings in `pkg/pagelayout/renderer/renderer.go` and `pkg/zinelayout/image.go`.
- `G404` weak random generator in `pkg/repo/sqlite/sqlite.go`.
- `G703` path traversal taint findings in serve/project/image paths.
- `G120` multipart parsing limit issue.
- `G112` missing `ReadHeaderTimeout`.
- `G107` variable URL `http.Get` findings in generated/CLI API commands.
- `G302` output file permission in `pkg/presets/presets.go`.
- `G706` log injection taint findings.

Tracker state was updated to:

```text
zine-layout action_status = dependency_scanning_failed_after_bump
```

Recommended next-week decision for `zine-layout`:

1. either fix high-confidence findings properly,
2. or add narrow, justified suppressions where findings are acceptable,
3. or mark the release intentionally blocked and continue the train without it.

Do not broad-disable gosec for the repo as release-train busywork.

## 4. Known Recurring Non-Blocking Failures

Several repos still show unrelated `Secret Scanning` failures. `smailnail` also has a `publish-image` startup failure. Earlier INFRA-004 policy treated these as outside the Go/lint/dependency baseline for this release train.

Do not silently ignore new failures. The distinction is:

- **Release-train relevant:** `golang-pipeline`, `golangci-lint`, `Dependency Scanning`, CodeQL where present, plus local `make logcopter-check`, `make glazed-lint`, and tests.
- **Previously treated as unrelated:** secret scanning and image publishing failures that pre-exist or are not caused by the bump.

## 5. How to Resume

Start here:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling
T=ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos
DB=$T/sources/05-rollout-progress.sqlite

git status --short
python $T/scripts/02-rollout-tracker.py deps-release-order
python $T/scripts/02-rollout-tracker.py deps-bumps
sqlite3 "$DB" "select repo,state,tag,action_status from repos order by repo;"
```

Check the diary before doing more work:

```bash
less $T/reference/01-diary.md
```

Most relevant diary steps:

- Step 16: Layer 1a release start.
- Step 17: Oak Glazed/Bobatea migration and release.
- Step 18: Refactorio release.
- Step 19: Zine-layout blocked by dependency scanning.
- Step 20: Jesus and Smailnail releases.

## 6. Mechanical Release Loop

For a candidate repo `R` with upstream tag `U@vX.Y.Z`:

```bash
cd /home/manuel/code/wesen/go-go-golems/R
git status --short
git checkout main
git pull --ff-only origin main

GOWORK=off go get github.com/go-go-golems/U@vX.Y.Z
GOWORK=off go mod tidy

make logcopter-check
make glazed-lint
GOWORK=off go test ./...
GOWORK=off go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --timeout=5m

git diff -- go.mod go.sum
git add go.mod go.sum
git commit -m "Bump U to vX.Y.Z"
git push origin main
```

Then poll Actions:

```bash
gh run list -R go-go-golems/R --branch main --commit <sha> --limit 10 \
  --json databaseId,name,status,conclusion,url \
  --jq '.[] | [.databaseId,.name,.status,(.conclusion//""),.url] | @tsv'
```

If rollout-relevant gates pass, create the next patch release from the current `main` SHA:

```bash
latest=$(gh api repos/go-go-golems/R/tags --jq '.[].name' | python -c 'import sys,re; vs=[]
for t in sys.stdin.read().splitlines():
 m=re.fullmatch(r"v(\d+)\.(\d+)\.(\d+)", t)
 if m: vs.append((tuple(map(int,m.groups())), t))
print(max(vs)[1] if vs else "v0.0.0")')

next=$(python - <<PY
import re
m=re.fullmatch(r'v(\d+)\.(\d+)\.(\d+)', '$latest')
print(f'v{m.group(1)}.{m.group(2)}.{int(m.group(3))+1}')
PY
)

main=$(gh api repos/go-go-golems/R/git/ref/heads/main --jq '.object.sha')
gh release create "$next" -R "go-go-golems/R" --target "$main" --title "$next" --generate-notes
```

Update tracker:

```bash
sqlite3 "$DB" "
update repos
set state='released',
    tag='$next',
    release_url='https://github.com/go-go-golems/R/releases/tag/$next',
    head_sha='$main',
    action_status='release_created',
    updated_at=datetime('now')
where repo='R';
insert into events(repo,kind,message,url,created_at)
values('R','release_created','Created $next at $main after dependency bump','https://github.com/go-go-golems/R/releases/tag/$next',datetime('now'));
"
```

Append to evidence TSV if it belongs to the current layer file:

```bash
printf 'R\t%s\t%s\thttps://github.com/go-go-golems/R/releases/tag/%s\n' "$next" "$main" "$next" \
  >> "$T/sources/release-train-20260529-layer2/releases.tsv"
```

Then update diary/changelog and commit infra docs/tracker changes.

## 7. Repo-Specific Validation Notes

### `smailnail`

Use sqlite tags for tests and CI-version golangci-lint:

```bash
GOWORK=off go test -tags sqlite_fts5 ./...
GOWORK=off GOFLAGS='-tags=sqlite_fts5' go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --timeout=5m
```

Running golangci-lint without the tag produced a sentinel typecheck failure in `pkg/mirror/require_fts5_build_tag.go`.

### `zine-layout`

Has a pushed bump but no release. See Section 3.

### `oak`

Was not just a bump. It required a focused migration from old Glazed APIs to new schema/fields/values APIs and a bobatea REPL API update. It is now released as `v0.5.3`; next work should not need to revisit Oak unless downstream repos expose new issues.

## 8. Suggested Next Candidates

Skip `zine-layout` for now unless the next session is specifically about its gosec backlog.

Reasonable mechanical candidates to try next:

| Candidate | Upstreams to bump | Notes |
|---|---|---|
| `scraper` | `sessionstream v0.0.6` | Should be a simple Layer 2 path if validation stays clean. |
| `sqleton` | `parka v0.6.3` | After release, it unblocks `form-generator`. Note current tracker already has old `sqleton` tag `v0.4.5`; verify newest tag before choosing next version. |
| `escuse-me` | `go-emrichen v0.0.11`, `parka v0.6.3` | May pull more transitive deps; validate carefully. |
| `go-go-agent` | `bobatea v0.1.6`, `go-emrichen v0.0.11` | Also depends on `geppetto`/`pinocchio`; watch for broader graph movement. |
| `go-go-app-inventory` | `go-go-os-backend v0.0.6`, `plz-confirm v0.0.5` | Also has `go-go-goja`/`pinocchio`; validate published-dep mode with `GOWORK=off`. |

Candidates that are just independent Layer 1 modules can also be tagged without downstream bump work, but the most valuable remaining work is releasing/bumping the downstream dependents.

## 9. Current Tracker Snapshot

As of this handoff, the state counts are:

```text
blocked             4
local_validation    1
main_actions_verified 25
planned             4
released            27
skipped             9
```

Selected rows:

```text
jesus       released              v0.0.1   release_created
oak         released              v0.5.3   release_created
refactorio  released              v0.0.1   release_created
smailnail   released              v0.0.1   release_created
zine-layout main_actions_verified          dependency_scanning_failed_after_bump
```

## 10. Git / Commit State

Important infra commits from today:

```text
465e332 INFRA-004: record release train layer 1a
7947366 INFRA-004: record oak release train bump
78830af INFRA-004: record refactorio release train bump
2f79912 INFRA-004: record zine-layout release blocker
4b31ed8 INFRA-004: record jesus and smailnail releases
```

Important repository commits from today:

```text
oak        d7a45ae Bump bobatea and update Glazed APIs
oak        fb1251d Suppress deferred Clay init migration warning
refactorio 3e9142b Bump oak to v0.5.3
zine-layout a8b2cba Bump go-emrichen to v0.0.11   # pushed, not released
jesus      64d32b1 Bump go-go-mcp to v0.0.19
smailnail  a747e77 Bump go-go-mcp to v0.0.19
```

Before resuming, confirm all involved worktrees are clean:

```bash
for r in oak refactorio zine-layout jesus smailnail; do
  echo "== $r =="
  git -C /home/manuel/code/wesen/go-go-golems/$r status --short
done

git -C /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling status --short
```

## 11. Stop Conditions

Continue autonomously for mechanical work: bump, tidy, local validation, push, poll, create release, update tracker and diary.

Stop and ask for help if any of these happen:

1. API migration is needed and is larger than the focused Oak-style migration.
2. Tests fail for behavioral reasons, not just missing build tags.
3. Dependency scanning fails with a broad backlog, like `zine-layout`.
4. A dependency bump drags in `go-go-goja`, `pinocchio`, or `geppetto` changes that alter APIs or runtime behavior.
5. The fix would require broad suppressions or disabling security gates.
6. GitHub Actions disagree with local validation in a way that is not just a known tag/GOFLAGS mismatch.

## 12. End-of-Day Recommendation

Next week, start with `scraper` or `sqleton`, not `zine-layout`.

- `scraper` should test the `sessionstream v0.0.6` downstream path.
- `sqleton` should test the `parka v0.6.3` downstream path and then potentially unblock `form-generator`.
- Keep `zine-layout` visible as a blocked security-cleanup item rather than trying to finish it as part of routine release mechanics.
