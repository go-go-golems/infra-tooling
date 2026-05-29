---
Title: Current Status, Release Order, and Rollout Lessons
Ticket: INFRA-004
Status: active
Topics:
  - automation
  - cli
  - release
  - docsctl
  - logcopter
  - github
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite
    Note: SQLite tracker used as the final status source of truth for this report.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md
    Note: Chronological implementation diary used to reconstruct the rollout sequence and issue taxonomy.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/changelog.md
    Note: Ticket changelog containing major rollout milestones.
  - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/06-open-prs.yaml
    Note: Tracker-derived ggg PR manifest; final state contains no open PR entries.
ExternalSources:
  - https://github.com/go-go-golems/logcopter/releases/tag/v0.1.0
Summary: Final INFRA-004 status report with dependency-aware release order, logcopter implications, and the main issues encountered during the rollout.
LastUpdated: 2026-05-29T16:30:00-04:00
WhatFor: Use this report to decide the next release train, dependency bump order, and follow-up cleanup tickets after the INFRA-004 rollout.
WhenToUse: Before tagging unreleased repositories, bumping first-party dependencies, or resuming B5/xgoja work.
---

# Current Status, Release Order, and Rollout Lessons

This report explains where INFRA-004 stands after the large cross-repository rollout, how to release and bump the remaining repositories in dependency order, and what went wrong along the way. It is intentionally written as a teaching document rather than a terse handoff. The goal is that the next operator can read this once, understand the system of constraints, and then run the next release train without rediscovering the same failure modes.

The core idea of the rollout is simple: add or verify a standard infrastructure baseline across many Go-Go-Golems repositories. The implementation was not simple because each repository sits in a graph of first-party dependencies, GitHub workflow variants, generated files, security scans, Codex review state, and historical local conventions. A green check mark was not always sufficient. A merge was safe only when the branch was mergeable, the relevant checks passed, current-head Codex feedback was satisfied, and the SQLite tracker reflected the same state the dashboard displayed.

## 1. Executive Summary

At the end of the last rollout pass, the tracked open PR queue is empty. The final tracker-derived manifest, `sources/06-open-prs.yaml`, contains only the top-level `prs:` key and no PR URLs. This is the practical meaning of “the open PR cleanup is done”: every PR that the tracker still considered open was repaired, rechecked through `ggg`, merged with a merge commit, and then verified on `main`.

The SQLite tracker currently reports the following state distribution:

| State | Count | Meaning |
|---|---:|---|
| `main_actions_verified` | 37 | PRs or main repairs landed, and the rollout-relevant main workflows were verified. |
| `released` | 15 | Release evidence was already recorded in the tracker. |
| `skipped` | 9 | The repository was intentionally skipped or deferred by earlier triage. |
| `blocked` | 4 | The repository cannot safely be rolled out mechanically without a separate decision. |
| `planned` | 4 | B5/xgoja or API-intent work remains planned but not implemented. |
| `local_validation` | 1 | Local validation exists, but the row is not complete in the release tracker. |

The most important operational conclusion is this:

> **The rollout PR phase is complete, but the release train is not complete.**

There are 36 repositories in `main_actions_verified` without a release URL recorded. Those repositories are ready for the next release train, subject to the dependency order in this report. The release train should not be run alphabetically. It should be run in layers, because downstream modules such as `form-generator`, `refactorio`, `smailnail`, and `jesus` should consume tagged versions of their first-party dependencies if the intent is to publish a coherent set of module releases.

The second important conclusion concerns `logcopter`:

> **`logcopter` itself does not appear to need a new release before the next train. The rollout repositories already require `github.com/go-go-golems/logcopter v0.1.0`, and local inspection found the expected `tool github.com/go-go-golems/logcopter/cmd/logcopter-gen` directive in the generated repositories.**

That does not mean `logcopter` is irrelevant. It means it is a foundation already pinned to `v0.1.0`. If `logcopter` changes again, it must be released before every downstream bump. If it does not change, the release train can start with the verified repositories that depend on it.

## 2. The Evidence Model

A rollout of this size needs a source of truth. In this ticket, the source of truth is the SQLite tracker at:

`/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite`

The tracker has three important tables:

| Table | Purpose |
|---|---|
| `repos` | One row per repository, with batch, tracks, upstreams, state, branch, PR URL, merge SHA, release URL, and action status. |
| `events` | Append-only event log for repairs, merges, Codex triggers, and verification. |
| `validations` | Command-level validation records. |

The workflow that produced this report used four evidence classes:

1. **Tracker state.** The `repos` table tells us which repositories are released, verified, blocked, skipped, or still planned.
2. **Diary state.** `reference/01-diary.md` explains why state changed, what failed, and what was done to repair it.
3. **Readiness evidence.** Timestamped `ggg` JSON files under `sources/` show how PRs moved through `failed_checks`, `codex_feedback`, `waiting_codex`, `waiting_checks`, and `ready`.
4. **GitHub Actions evidence.** Post-merge `gh run list` checks verified `golang-pipeline`, `golangci-lint`, dependency scanning, CodeQL, or equivalent rollout workflows on `main`.

The important distinction is between **PR readiness** and **release readiness**. PR readiness asks: can this branch be merged now? Release readiness asks: should this repository be tagged now, and should its downstream consumers be bumped to that tag? The rollout has completed the first question for tracked open PRs. The rest of this report answers the second question.

## 3. Current Status by Batch

The original batch plan separated repositories by risk and dependency position. B1 contained foundations and upstream libraries. B2 contained leaf logcopter-only repositories. B3 added Glazed linting without docsctl. B4 added docsctl and Glazed CLI policy to leaf packages and moderately connected packages. B5 was reserved for xgoja/provider/API-intent candidates.

| Batch | State | Count |
|---|---:|---:|
| B1 | `local_validation` | 1 |
| B1 | `main_actions_verified` | 3 |
| B1 | `released` | 3 |
| B1 | `skipped` | 2 |
| B2 | `blocked` | 4 |
| B2 | `released` | 7 |
| B2 | `skipped` | 1 |
| B3 | `main_actions_verified` | 2 |
| B3 | `released` | 1 |
| B3 | `skipped` | 1 |
| B4 | `main_actions_verified` | 29 |
| B4 | `skipped` | 5 |
| B5 | `main_actions_verified` | 3 |
| B5 | `planned` | 4 |
| B5 | `released` | 4 |

This table has one subtle point: `main_actions_verified` is a strong CI state, not a release state. It means the merged code was observed to work on `main` for the rollout-relevant workflows. It does not mean a Git tag and GitHub release were created.

## 4. What Is Finished

The tracked open PR phase is finished. The final open-PR manifest contains no PR URLs. The last wave of merged PRs included:

| Repository | PR | Merge commit |
|---|---:|---|
| `codex-sessions` | #2 | `5b3c3000f30ba120aca58962fda449d470ef47df` |
| `docmgr` | #38 | `fcce3b05109ce55986550489b61d26c6b62bd246` |
| `font-util` | #1 | `3ef7ebdd973196fb432d979833c5f46aec35148d` |
| `go-go-mcp` | #82 | `6fee74138607915071f82c8f11172b4e97011824` |
| `refactorio` | #1 | `3921c65656f779dc932d84cadbeabe2ff3b1ef65` |
| `smailnail` | #4 | `858202456989816978817009bea599db7700e26b` |
| `bobatea` | #97 | `1746358ea61ca46d4370fb6f8a92fed53a28f9d2` |
| `oak` | #47 | `c89acf357980358ccf9967b7a3fb891def9c92bc` |
| `openai-mock-server` | #2 | `6fc0902a768906ca797bbef362b6f2e6ea190371` |
| `vault-envrc-generator` | #9 | `a7df61047f4536c726122f0412391a6444328c78` |
| `jesus` | #7 | `6ed5e5ef06b6080815293a2af78571cb698d1729` |

Earlier in the same workstream, additional PRs had already been merged and verified. The combined effect is that the dashboard no longer has a queue of tracked `pr_open` rows.

The post-merge main verification also completed for the newly merged repositories. In several repositories, secret scanning still failed. Those failures are not ignored; they are simply not the same as the Go/lint/security rollout gate. The rollout gate was concerned with workflows such as `golang-pipeline`, `golangci-lint`, dependency scanning, CodeQL, `govulncheck`, and `gosec`. Secret scanning and image publishing require their own ownership path.

## 5. What Is Not Finished

The unfinished work is not a hidden PR queue. It is a set of release, blocked, skipped, and planned rows.

| Repository | Batch | State | Reason |
|---|---|---|---|
| `infra-tooling` | B1 | `local_validation` | Local validation exists, but the row is not complete in the tracker. |
| `bubble-table` | B2 | `blocked` | Module path is `github.com/evertras/bubble-table` and there are no normal `cmd/pkg/internal` rollout targets; ownership/module intent must be confirmed. |
| `raza` | B2 | `blocked` | Module path is `github.com/wesen/raza` and logging style is mixed; needs manual logging plan. |
| `terraform-provider-stytch-b2b` | B2 | `blocked` | Module path is `github.com/mento/terraform-provider-stytch-b2b`; ownership/module intent must be confirmed. |
| `voyage` | B2 | `blocked` | Repository is archived/read-only, and local tests had pre-existing build failures. |
| `barbar` | B2 | `skipped` | Mechanical generation produced no useful `logcopter.go` for the root-only package main repo. |
| `common-sense`, `plunger`, `biberon`, `bucheron`, `ecrivain`, `geppetto`, `mastoid` | B1/B3/B4 | `skipped` | Skipped earlier; each requires a manual decision before resuming. |
| `go-go-goja`, `go-minitrace`, `pinocchio`, `workspace-manager` | B5 | `planned` | API-intent/xgoja work remains planned and was not implemented in this PR cleanup pass. |

The most consequential skipped row is `geppetto`. Many downstream repositories depend on `geppetto`. If the next release train wants every downstream module to consume the newest first-party infrastructure baseline, `geppetto` must be revisited. If the next release train only wants to tag repositories whose own rollout PRs landed, then the train can proceed without waiting for `geppetto`, but downstream `go.mod` files will not necessarily converge on a fully updated graph.

## 6. Logcopter: The Constraint That Shapes the Train

`logcopter` matters because generated code and check targets are now present across many repositories. The rollout pattern is:

```text
logcopter_generate.go
  -> go:generate go tool logcopter-gen ... ./cmd/... ./pkg/...
  -> generated logcopter.go files
  -> make logcopter-check verifies the generated files are current
  -> CI runs generation/checking before tests
```

This pattern makes generated files part of the repository contract. If a repository changes package structure, adds command packages, or changes the `go:generate` package list, the generated files and `make logcopter-check` must stay aligned.

The current status is favorable for the release train:

- `logcopter` has a recorded released version, `v0.1.0`.
- The rollout repositories inspected locally require `github.com/go-go-golems/logcopter v0.1.0`.
- Most generated repositories include the Go tool directive `tool github.com/go-go-golems/logcopter/cmd/logcopter-gen`.
- The final issue wave was not caused by a missing `logcopter` release. It was caused by repository-specific generation/check target mismatches and ignored generated files.

The practical rule is:

> If `logcopter` remains at `v0.1.0`, do not start the release train by retagging `logcopter`. Start with downstream repositories in topological order. If `logcopter` changes, release `logcopter` first, then bump every repository that runs `go tool logcopter-gen`, regenerate files, run `make logcopter-check`, and only then tag downstream modules.

The repositories most directly shaped by the logcopter constraint are the B5 planned repositories:

| Repository | Why logcopter matters |
|---|---|
| `go-go-goja` | It is planned B5 work and depends on `logcopter`; many xgoja consumers depend on it. |
| `go-minitrace` | It depends on both `go-go-goja` and `logcopter`; it should not be released before `go-go-goja` if the goal is a coherent graph. |
| `pinocchio` | It depends on `bobatea`, `go-go-goja`, `logcopter`, `sanitize`, `sessionstream`, and `uhoh`; it is a downstream convergence point. |
| `workspace-manager` | It depends on `go-go-goja` and `logcopter`; release it after `go-go-goja`. |

## 7. Recommended Release and Bump Order

The next release train should be run in layers. A layer can be executed in parallel if the repositories are independent. Between layers, downstream repositories should bump first-party dependencies to the tags produced by earlier layers, then run validation before tagging.

There are two variants of the plan. The first is the **practical release train** for the repositories already verified on `main`. The second is the **B5/xgoja continuation train**, which depends on planned work that has not been implemented yet.

### 7.1 Practical Release Train for Verified Repositories

The following layers are computed from the tracker’s `upstreams` field for repositories in `main_actions_verified` without release URLs. Dependencies outside this candidate set are treated as already external to this train, skipped, or separate prerequisites.

#### Layer 1: release the independent verified repositories

Release these first. They have no dependency on another unreleased verified repository in this train.

| Repository group | Repositories |
|---|---|
| B1/B3 foundations and independent libraries | `bobatea`, `go-go-os-backend` |
| B4 mostly leaf CLI/library packages | `almanach`, `cliopatra`, `codex-sessions`, `devctl`, `docmgr`, `font-util`, `gitcommit`, `go-emrichen`, `go-go-mcp`, `harkonnen`, `js-analyzer`, `openai-mock-server`, `parka`, `plz-confirm`, `prescribe`, `prompto`, `remarquee`, `sessionstream`, `tactician`, `uhoh`, `vault-envrc-generator`, `vm-system`, `web-agent-example` |

For operator efficiency, prioritize the repositories that other unreleased repositories depend on:

1. `bobatea`
2. `go-emrichen`
3. `parka`
4. `go-go-mcp`
5. `go-go-os-backend`
6. `plz-confirm`
7. `sessionstream`
8. `uhoh`

After those are tagged, the rest of Layer 1 can be tagged in any convenient order.

#### Layer 2: bump Layer 1 tags into their dependents, then release

These repositories depend on one or more Layer 1 repositories. Before tagging them, bump their first-party dependencies to the Layer 1 tags where applicable, run validation, and then tag.

| Repository | Wait for | Why |
|---|---|---|
| `oak` | `bobatea` | `oak` depends on `bobatea`. |
| `sqleton` | `parka` | `sqleton` depends on `parka`. |
| `escuse-me` | `go-emrichen`, `parka` | Both are upstreams in the tracker. |
| `go-go-agent` | `bobatea`, `go-emrichen` | Both are upstreams, plus external/skipped dependencies such as `geppetto` and `pinocchio`. |
| `go-go-app-inventory` | `go-go-os-backend`, `plz-confirm` | Both are unreleased verified upstreams. |
| `jesus` | `go-go-mcp` | `jesus` depends on `go-go-mcp`. |
| `smailnail` | `go-go-mcp` | `smailnail` depends on `go-go-mcp`. |
| `scraper` | `sessionstream` | `scraper` depends on `sessionstream`. |
| `zine-layout` | `go-emrichen` | `zine-layout` depends on `go-emrichen`. |

Layer 2 is where release hygiene matters most. If a repository has already been merged and verified on `main`, it is tempting to tag it immediately. That works for a local release, but it does not produce a dependency-converged release train. A dependency-converged train tags upstreams first, then updates downstream `go.mod` files to consume those tags.

#### Layer 3: release final dependents

These repositories depend on Layer 2 outputs.

| Repository | Wait for | Why |
|---|---|---|
| `form-generator` | `sqleton`, `uhoh` | `form-generator` depends on both. `uhoh` is Layer 1; `sqleton` is Layer 2. |
| `refactorio` | `oak` | `refactorio` depends on `oak`. |

`refactorio` deserves special attention because the rollout removed a local replacement for `oak` and selected released `oak v0.5.2` during PR repair. Before tagging `refactorio`, check whether the newly released `oak` tag should replace the currently selected version.

### 7.2 B5 and xgoja Continuation Train

The B5 train is different. It contains planned work, not just release tagging. The order should be:

1. Confirm whether `logcopter v0.1.0` remains the intended generator/runtime version. If not, release `logcopter` first and bump every downstream generator user.
2. Resolve the `geppetto` decision. `go-go-goja` and several important applications depend on it, but the tracker currently marks `geppetto` as skipped.
3. Implement and verify `go-go-goja` B5 work.
4. Release `go-go-goja`.
5. Bump `go-go-goja` into direct xgoja consumers: `go-minitrace`, `workspace-manager`, `pinocchio`, and any already-verified B5 consumers that should converge on the new tag.
6. Release `go-minitrace` and `workspace-manager` after their `go-go-goja` bump and validation.
7. Release `pinocchio` after `bobatea`, `go-go-goja`, `sanitize`, `sessionstream`, and `uhoh` are all at intended tags.
8. Revisit higher-level applications that depend on `pinocchio`, especially `go-go-agent`, `jesus`, `web-agent-example`, and `go-go-app-inventory`, if the release train’s goal includes dependency convergence rather than only tagging already-merged code.

The key point is that `logcopter` is not the current blocker for B5. The current blockers are API-intent work and upstream graph decisions around `go-go-goja`, `pinocchio`, and `geppetto`.

## 8. Suggested Bump Procedure

For each repository in the release train, use the same mechanical loop. The point of the loop is to keep repository state, generated files, and the tracker consistent.

```bash
# 1. Start from clean main.
git checkout main
git pull --ff-only origin main

# 2. Bump first-party dependencies released in earlier layers.
GOWORK=off go get github.com/go-go-golems/<upstream>@<tag>
GOWORK=off go mod tidy

# 3. Regenerate and check generated files.
GOWORK=off go generate ./...
make logcopter-check
make glazed-lint

# 4. Run tests with the same tags used in CI.
GOWORK=off go test ./...
# or repository-specific tags, for example:
GOWORK=off go test -tags sqlite_fts5 ./...

# 5. Commit the bump.
git add go.mod go.sum '**/logcopter.go' Makefile .github/workflows
git commit -m "Bump rollout dependencies"
git push

# 6. Verify CI, tag, release, and record in the tracker.
```

The exact test command is repository-specific. `smailnail` needs sqlite tags. Some repositories have docsctl export checks. Some have web generation. The safe rule is to read the repository’s CI workflow and run the same commands locally before tagging.

## 9. Issues Encountered

This section is long because the failure modes are the most useful part of the ticket. The implementation itself is mechanical once the failure modes are known.

### 9.1 The tracker was stale and had to be reconciled

The first operational issue was trust. The previous handoff was useful but not authoritative. The dashboard rendered from `sources/05-rollout-progress.sqlite`, so the first priority was to reconcile that database with live GitHub state. Without that step, every later readiness decision would have been ambiguous.

The tracker reconciliation established three habits that should remain standard:

- Regenerate `sources/06-open-prs.yaml` from SQLite before each `ggg batch ready` run.
- Record branch head SHAs, merge SHAs, Codex triggers, and main verification events in SQLite.
- Treat diary text as narrative evidence, not as the canonical state machine.

### 9.2 Several workflows had malformed YAML step structure

The most damaging rollout bug was a malformed `push.yml` pattern:

```yaml
      -
        name: run unit tests
      - name: Verify Glazed CLI policy
        run: make glazed-lint

        run: go test ./...
```

This is not a harmless formatting issue. It changes the executable structure of the workflow. The intended unit-test step loses its `run`, and the final `run: go test ./...` is attached incorrectly. Some PRs looked greener than they should have because the workflow did not execute the intended command in the intended step.

Codex repeatedly caught this class of issue even when raw GitHub checks were not enough. The fix was to restore separate executable steps:

```yaml
      - name: Verify Glazed CLI policy
        run: make glazed-lint
      - name: Run unit tests
        run: go test ./...
```

The same pattern appeared in multiple repositories and even survived into `smailnail` main after merge. That required direct main repair.

### 9.3 The Glazed analyzer version was not always installable

The rollout added a `make glazed-lint` target that installs:

```text
github.com/go-go-golems/glazed/cmd/tools/glazed-lint@<version>
```

Older Glazed tags did not contain that package. The observed failure looked like this:

```text
go: github.com/go-go-golems/glazed/cmd/tools/glazed-lint@v1.2.7: module github.com/go-go-golems/glazed@v1.2.7 found, but does not contain package github.com/go-go-golems/glazed/cmd/tools/glazed-lint
```

The broad repair was to use a Glazed version that contains the analyzer. Most repositories converged on `v1.3.6`; some Codex feedback argued for `v1.3.4` where it believed that was the highest published tag. The local tag check showed `v1.3.6` exists, but current-head Codex feedback still had to be satisfied for merge readiness. The lesson is to verify both the actual module and the current-head review state. `ggg` treats current-head Codex state as part of readiness.

### 9.4 Generated files were sometimes hidden by `.gitignore`

`jesus` generated command-package logcopter files under paths ignored by `.gitignore`. A normal `git status` did not show the files, but `go generate` and `make logcopter-check` made clear that generated files were missing from the committed tree. The repair was to force-add the generated files:

```bash
git add -f cmd/jesus/logcopter.go cmd/jesus/cmd/logcopter.go
```

This is a general logcopter lesson. Generated files are part of the rollout contract. If `.gitignore` hides them, the operator must either force-add them or change the generation pattern. Leaving them untracked will make CI fail at `git diff --exit-code` or `make logcopter-check`.

### 9.5 Logcopter package lists had to stay aligned

Several repositories had mismatched package lists between generation, checking, and linting. `vault-envrc-generator` is the clearest example. The generator scanned `./cmd/... ./cmds/... ./pkg/...`, but the check or glazed-lint target scanned only `./cmd/... ./pkg/...`. Codex correctly pointed out that the real command implementations live under `cmds/`, so the policy check was incomplete.

The repair was to make the package lists identical where they represented the same contract:

```text
./cmd/... ./cmds/... ./pkg/...
```

The deeper rule is that package-list drift is a correctness bug. A generator target, a check target, and a policy-lint target are three views of the same package set. If they disagree, CI may pass while the repository is only partially covered.

### 9.6 Security tooling failures had different root causes

The security failures were not one problem. They were several problems with similar GitHub check names.

| Failure class | Example repositories | Resolution |
|---|---|---|
| Vulnerable `go-jose` versions | `go-go-mcp`, `jesus`, `vault-envrc-generator`, `smailnail` | Bumped `go-jose` to fixed versions and reran `govulncheck`. |
| Gosec rule changes or taint findings | `codex-sessions`, `docmgr`, `font-util`, `go-go-mcp`, `jesus`, `oak`, `openai-mock-server` | Added scoped excludes or repaired code where appropriate. |
| Dependency Review unavailable | `font-util`, `openai-mock-server` | For `font-util`, made unsupported Dependency Review non-blocking. For `openai-mock-server`, removed the unavailable job with an explanatory comment and kept `govulncheck`/`gosec` blocking. |
| Nancy/OSS Index 401 or unavailable behavior | `oak`, `openai-mock-server` | Removed Nancy jobs rather than weakening them with `continue-on-error`; kept `govulncheck` and `gosec` blocking. |
| Secret scanning failures | Many main branches | Recorded as unrelated to this rollout’s Go/lint/security verification. |

The important decision was to distinguish a scan that is unavailable from a scan that found a real vulnerability. Initially, some workflows used `continue-on-error` for Nancy or Dependency Review. Codex correctly objected that this weakens the policy. The final approach removed unavailable integrations when they could not function in the repository and left available Go-native security gates blocking.

### 9.7 Local tools did not always match CI tools

`prescribe` exposed a common pitfall. The local `golangci-lint` binary reported no issues, but CI used `golangci-lint v2.12.2` and still found `QF1012` findings. The correct local reproduction was:

```bash
go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run --timeout=5m
```

After switching to the CI-pinned version, the remaining findings became visible and were fixed. This should be the default method whenever local lint disagrees with CI.

### 9.8 Some dependency graphs contained local-only replacements

`refactorio` depended on `oak` through a local replacement:

```text
replace github.com/go-go-golems/oak => ../oak
```

That is acceptable for local development but not for a publishable module release. The repair removed the local replacement, selected a released `oak` version, bumped Glazed, and then fixed the lint findings that appeared under CI-version lint. Before `refactorio` is tagged in the final release train, it should be checked again against the newly released `oak` tag produced by this train.

### 9.9 Some disabled docsctl templates were better removed than kept

`go-go-app-inventory` and `sanitize` had disabled docsctl publishing templates even though the tracker marked `needs_docsctl=0` and local commands did not expose working Glazed help export. Keeping disabled, broken release templates is a future footgun. The repair removed those disabled templates and documented why.

The lesson is that a disabled workflow block is still part of the repository’s operational story. If it points at a command that does not exist, it will become a problem when someone enables it later.

### 9.10 One validation command produced an accidental artifact

Running a `go-go-app-inventory` validation command created `data/inventory.db`. That file was accidentally committed, then immediately removed, and `data/` was added to `.gitignore`. This was caught and corrected during the PR repair loop.

The general rule is that validation commands can write state. Before committing, always run:

```bash
git status --short --untracked-files=all
```

and inspect every generated artifact.

### 9.11 Direct main repairs were exceptional but necessary

Direct pushes to `main` are not the default workflow. They were used only when broken workflow code had already landed on `main` and prevented verification. The most recent example was `smailnail`:

| Commit | Reason |
|---|---|
| `1016b6329309ee72bad17bd842539289fe50f34a` | Fixed the malformed main `push.yml` unit-test step. |
| `c756253204ca3f6a689f64d294273e28351599cd` | Fixed the main `glazed-lint` gate by adding sqlite tags and scoped legacy allow paths. |

After the second repair, `smailnail` main `golang-pipeline`, `golangci-lint`, dependency scanning, and CodeQL succeeded on `c756253`.

## 10. The Operating Model That Worked

The reliable loop was:

```text
1. Regenerate the PR manifest from SQLite.
2. Run ggg batch ready.
3. For each non-ready PR, inspect either failed checks or current-head Codex feedback.
4. Reproduce locally with the CI-equivalent command.
5. Push a narrow fix.
6. Update the SQLite tracker.
7. Trigger Codex if the PR is waiting for current-head review.
8. Re-run ggg.
9. Merge only when ggg says ready and per-PR Codex comments are satisfied/current.
10. Verify main workflows after merge.
11. Record merge SHA, main verification, diary, and changelog.
```

This worked because each step has a single responsibility. SQLite tracks state. `ggg` decides readiness. Local commands reproduce failure classes. GitHub Actions confirms remote behavior. The diary explains why the operator made each decision.

## 11. What To Do Next

The next operator should not look for more tracked open PRs. There are none. The next operator should decide which of these three paths they are taking.

### Path A: Run the practical release train

Use the Layer 1 -> Layer 2 -> Layer 3 plan in Section 7.1. Tag upstreams first, bump downstream `go.mod` files, run local validation, and tag dependents. This is the recommended next step if the goal is to publish the merged rollout work.

### Path B: Resume B5/xgoja work

Start with the planned rows: `go-go-goja`, `go-minitrace`, `pinocchio`, and `workspace-manager`. Do not treat these as simple release tasks. They require API-intent confirmation and likely design work.

### Path C: Clean up skipped/blocked repositories

Open a follow-up ticket for skipped and blocked repositories. The blocked B2 repositories are not good candidates for mechanical rollout. They require ownership decisions, module path decisions, or archived-repository decisions.

## 12. Short Answer to the Release-Order Question

If the question is “what should we release and bump next, especially because of logcopter?”, the short answer is:

1. Do **not** release `logcopter` again unless it changed after `v0.1.0`. The rollout graph is already pinned to `logcopter v0.1.0`.
2. Release the upstream-ish verified repositories first: `bobatea`, `go-emrichen`, `parka`, `go-go-mcp`, `go-go-os-backend`, `plz-confirm`, `sessionstream`, and `uhoh`.
3. Release the remaining Layer 1 leaves: `almanach`, `cliopatra`, `codex-sessions`, `devctl`, `docmgr`, `font-util`, `gitcommit`, `harkonnen`, `js-analyzer`, `openai-mock-server`, `prescribe`, `prompto`, `remarquee`, `tactician`, `vault-envrc-generator`, `vm-system`, and `web-agent-example`.
4. Bump those tags into Layer 2 dependents and release: `oak`, `sqleton`, `escuse-me`, `go-go-agent`, `go-go-app-inventory`, `jesus`, `smailnail`, `scraper`, and `zine-layout`.
5. Finish with Layer 3: `form-generator` after `sqleton` and `uhoh`, and `refactorio` after `oak`.
6. Treat B5 separately: resolve `go-go-goja` and `geppetto` before attempting a dependency-converged xgoja release train.

## 13. Appendix: Tracker-Derived Tables

### Repositories with release evidence already recorded

| Repository | State | Tag | Release URL |
|---|---|---|---|
| `dmeta` | `released` | `v0.0.1` | https://github.com/go-go-golems/dmeta/releases/tag/v0.0.1 |
| `esper` | `released` | `v0.0.1` | https://github.com/go-go-golems/esper/releases/tag/v0.0.1 |
| `go-sqlite-regexp` | `released` | `v0.0.2` | https://github.com/go-go-golems/go-sqlite-regexp/releases/tag/v0.0.2 |
| `sanitize` | `main_actions_verified` | `v0.0.2` | https://github.com/go-go-golems/sanitize/releases/tag/v0.0.2 |
| `ai-in-action-app` | `released` | `v0.0.1` | https://github.com/go-go-golems/ai-in-action-app/releases/tag/v0.0.1 |
| `go-go-agent-action` | `released` | `v1.0.2` | https://github.com/go-go-golems/go-go-agent-action/releases/tag/v1.0.2 |
| `go-go-app-arc-agi` | `released` | `v0.0.2` | https://github.com/go-go-golems/go-go-app-arc-agi/releases/tag/v0.0.2 |
| `go-go-app-sqlite` | `released` | `v0.0.2` | https://github.com/go-go-golems/go-go-app-sqlite/releases/tag/v0.0.2 |
| `go-go-host` | `released` | `v0.0.1` | https://github.com/go-go-golems/go-go-host/releases/tag/v0.0.1 |
| `oak-git-db` | `released` | `v0.0.1` | https://github.com/go-go-golems/oak-git-db/releases/tag/v0.0.1 |
| `salad` | `released` | `v0.0.1` | https://github.com/go-go-golems/salad/releases/tag/v0.0.1 |
| `cozodb-goja` | `released` | `v0.0.2` | https://github.com/go-go-golems/cozodb-goja/releases/tag/v0.0.2 |
| `go-go-gepa` | `released` | `v0.0.2` | https://github.com/go-go-golems/go-go-gepa/releases/tag/v0.0.2 |
| `goja-github-actions` | `released` | `v0.0.1` | https://github.com/go-go-golems/goja-github-actions/releases/tag/v0.0.1 |
| `openai-app-server` | `released` | `v0.0.1` | https://github.com/go-go-golems/openai-app-server/releases/tag/v0.0.1 |

### Release layers for verified repositories

| Layer | Repositories | Internal reason |
|---:|---|---|
| 1 | `almanach`, `bobatea`, `cliopatra`, `codex-sessions`, `devctl`, `docmgr`, `font-util`, `gitcommit`, `go-emrichen`, `go-go-mcp`, `go-go-os-backend`, `harkonnen`, `js-analyzer`, `openai-mock-server`, `parka`, `plz-confirm`, `prescribe`, `prompto`, `remarquee`, `sessionstream`, `tactician`, `uhoh`, `vault-envrc-generator`, `vm-system`, `web-agent-example` | No dependency on another unreleased verified repository. |
| 2 | `escuse-me`, `go-go-agent`, `go-go-app-inventory`, `jesus`, `oak`, `scraper`, `smailnail`, `sqleton`, `zine-layout` | These depend on Layer 1 outputs such as `bobatea`, `go-emrichen`, `go-go-mcp`, `go-go-os-backend`, `parka`, `plz-confirm`, `sessionstream`, or `uhoh`. |
| 3 | `form-generator`, `refactorio` | `form-generator` waits for `sqleton`; `refactorio` waits for `oak`. |

### Final references

- SQLite tracker: `sources/05-rollout-progress.sqlite`
- Final open PR manifest: `sources/06-open-prs.yaml`
- Final diary entry: `reference/01-diary.md`, Step 15
- Final changelog entry: `changelog.md`, “Closed remaining tracked open PRs”
- Final readiness evidence: timestamped `sources/07-ggg-batch-ready-*.json` and `sources/20-*` through `sources/30-*`
