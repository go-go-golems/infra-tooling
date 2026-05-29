# Changelog

## 2026-05-28

- Initial workspace created


## 2026-05-28

Created INFRA-004 ticket, initial rollout implementation guide, batch planner script, and generated batch artifacts from INFRA-003 inventory.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/analysis/01-rollout-analysis-and-implementation-guide.md — Initial rollout guide


## 2026-05-28

Added SQLite rollout tracker, CLI update verbs, auto-refresh dashboard, initialized current B2 PR progress, and started dashboard in tmux session infra004-dashboard.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py — Progress tracker implementation


## 2026-05-28

Advanced B2 logcopter-only wave: merged oak-git-db, go-go-agent-action, go-go-app-arc-agi, and salad with merge commits; verified or classified main actions; recorded voyage blocked and barbar skipped edge cases in SQLite tracker.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Detailed Step 3 rollout diary
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Current rollout progress state


## 2026-05-28

Completed B2 safe-mechanical triage: all remaining B2 repos are now blocked/skipped in tracker pending manual decisions about stdlib logging conflicts, archived/external modules, or placeholder module identity.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated B2 triage states


## 2026-05-28

Updated ggg action status behavior so no_runs is terminal/non-blocking under --watch, exits 0, reports ok=true, and no longer keeps batch action watches pending; installed updated ggg to ~/.local/bin.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/internal/cli/batch/actions.go — Batch action watch loop now only waits on pending
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/internal/cli/run/status.go — Single-repo watch loop now only waits on pending
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/pkg/actionstatus/actionstatus.go — State and exit-code semantics for no_runs

## 2026-05-29

Opened B3/B4/B5 logcopter baseline PRs for 5 repos (openai-mock-server #1, go-emrichen #39, cliopatra #17, escuse-me #83, jesus #7). Fixed govulncheck failures by upgrading go directive to 1.26.3 across 5 repos. Merged gitcommit #2. Discovered that toolchain directive is ignored by setup-go@v6 due to GOTOOLCHAIN=local.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Step 5: B3/B4/B5 PR wave

## 2026-05-29

Created scripts/03-fix-ci-workflows.py to systematically align all 15 repos' CI workflows with go-template canonical patterns: setup-go@v6 + go-version-file, golangci-lint-action@v9 + version-file, checkout@v6, added .golangci-lint-version. Bumped go directive to 1.26.3 in all repos.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/03-fix-ci-workflows.py — Systematic CI alignment script
- /home/manuel/code/wesen/go-go-golems/go-template/.github/workflows/lint.yml — Canonical lint workflow reference

## 2026-05-29

Fixed cascading CI failures: bumped golangci-lint to v2.12.2 (Go 1.26.3 compatible), fixed .golangci.yml v2 format for cliopatra/harkonnen, replaced gosec Docker action with go install, fixed pre-existing lint issues (QF1008, QF1012, S1009), upgraded golang.org/x/net to fix GO-2026-5026.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Steps 7-9: CI fix cascade and merge wave

## 2026-05-29

Merging wave: merged and tagged 7 repos (parka v0.6.2, go-go-app-inventory v0.0.2, markdown-quizz v0.0.1, openai-mock-server v0.0.2, sqleton v0.4.5, oak v0.5.2, cliopatra v0.6.4). Release workflows revealed pre-existing infra issues (goreleaser configs, homebrew tap auth, Docker registry auth). 8 PRs remain open with pre-existing failures.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated tracker with 22 released repos

## 2026-05-29

Opened logcopter baseline PRs for all 15 remaining non-xgoja planned repos (bobatea, go-go-os-backend, almanach, codex-sessions, font-util, form-generator, go-go-agent, go-go-mcp, js-analyzer, prescribe, prompto, sessionstream, tactician, uhoh, vault-envrc-generator, web-agent-example, zine-layout). Skipped mastoid (archived), logcopter (self-referential), geppetto (docsctl-only). All 25 PRs now in CI/Codex review.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Step 10: Batch PR opening
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated tracker with 25 open PRs

## 2026-05-29

Added glazed-lint Makefile targets to 31 repos and publish-docs release job to 10 repos. Pushed to 24 existing logcopter baseline PR branches. Opened 7 new PRs for already-released repos (sanitize, go-go-app-inventory, cliopatra, oak, openai-mock-server, parka, sqleton).

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Step 11: glazed-lint + publish-docs
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/04-add-glazed-lint-docsctl.py — Script for adding glazed-lint + docsctl

## 2026-05-29

Step 12: repaired P0 broken main push.yml workflows, fixed glazed-lint analyzer version issues, and verified latest main push.yml runs for almanach, form-generator, tactician, and web-agent-example.

### Related Files

- /home/manuel/code/wesen/go-go-golems/almanach/.github/workflows/push.yml — P0 main workflow repair
- /home/manuel/code/wesen/go-go-golems/form-generator/.github/workflows/push.yml — P0 main workflow repair
- /home/manuel/code/wesen/go-go-golems/tactician/.github/workflows/push.yml — P0 main workflow repair
- /home/manuel/code/wesen/go-go-golems/web-agent-example/.github/workflows/push.yml — P0 main workflow repair
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated dashboard source of truth after P0 verification


## 2026-05-29

Step 13: repaired P1 open PR branches, bumped stale glazed-lint analyzer pins, pushed branch fixes, and started CI triage for newly surfaced lint failures.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Recorded P1 branch repair and CI triage
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated PR branch head SHAs and P1 events


## 2026-05-29

Step 14: switched the active P1 loop to `ggg` readiness, merged fourteen `ggg`-ready PRs, fixed current-head Codex feedback and additional lint/glazed-lint blockers, and updated the SQLite dashboard tracker with merge SHAs, branch heads, Codex trigger events, and main workflow verification.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Step 14 ggg-driven readiness, Codex feedback, and repair diary
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated dashboard source of truth
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/06-open-prs.yaml — Tracker-derived open PR manifest for `ggg batch ready`

## 2026-05-29 16:06 UTC — Closed remaining tracked open PRs

- Repaired the remaining failed-check/Codex-feedback PRs using `ggg` readiness snapshots, current-head Codex comments, local validation, and targeted branch fixes.
- Merged the final tracked open PRs with merge commits: codex-sessions #2, docmgr #38, font-util #1, go-go-mcp #82, refactorio #1, smailnail #4, bobatea #97, oak #47, openai-mock-server #2, vault-envrc-generator #9, and jesus #7.
- Verified post-merge main rollout workflows for the merged repositories; unrelated secret scanning and image publish failures remain separate from Go/lint/security rollout verification.
- Directly repaired smailnail main after merge with commits 1016b63 and c756253, then verified the main golang-pipeline succeeded.
- Updated `sources/05-rollout-progress.sqlite`; `sources/06-open-prs.yaml` now contains no PR entries.

## 2026-05-29 16:30 UTC — Added status and release-order report

- Added `analysis/02-current-status-release-order-and-rollout-lessons.md`, an evidence-based report using the diary, SQLite tracker, ggg readiness logs, and changelog.
- Documented the current tracker status, the dependency-aware release/bump order, logcopter implications, and the major rollout issues encountered.
- Linked the report from `index.md` for ticket discoverability.

## 2026-05-29 18:10 UTC — Added normalized internal dependency tables

- Extended `sources/05-rollout-progress.sqlite` with normalized Go-Go-Golems dependency data for release-order and bump planning.
- Added tables populated from local `go.mod` files: `internal_modules`, `internal_dependency_edges`, `release_order_layers`, and `dependency_bump_candidates`.
- Created INFRA-005 to design dashboard improvements that expose these tables as release train, bump candidate, evidence, and health-check views.

## 2026-05-29 18:40 UTC — Added derived issue/fix log tables

- Extended `sources/05-rollout-progress.sqlite` with `repo_issue_log` and `repo_issue_steps` so repository detail pages can show grouped issues, fixes, validations, and source event references.
- Populated the derived issue log from existing tracker `events` and `validations`: 342 issue rows and 808 issue timeline steps.

## 2026-05-29 19:45 UTC — Added derived health-check table

- Extended `sources/05-rollout-progress.sqlite` with `repo_health_checks` for lightweight logcopter and Glazed lint dashboard health panels.
- Populated 623 health-check rows from local repository files.

## 2026-05-29 18:35 UTC — Released oak after bobatea/glazed API bump

- Migrated `oak` command loading/runtime code from old Glazed `layers`/`parameters` APIs to current `schema`/`fields`/`values` APIs.
- Updated `cmd/oak-repl` to the current bobatea REPL streaming/event-bus API.
- Pushed `oak` main commits `d7a45ae` and `fb1251d`, verified rollout-relevant CI gates, and released `oak v0.5.3`.

## 2026-05-29 18:48 UTC — Released refactorio after oak bump

- Bumped `refactorio` from `oak v0.5.2` to `oak v0.5.3`.
- Validated logcopter, glazed-lint, tests, and CI-pinned golangci-lint locally.
- Pushed `refactorio` commit `3e9142b`, verified rollout-relevant main checks, and released `refactorio v0.0.1`.

## 2026-05-29 18:58 UTC — Paused zine-layout release after dependency scanning failure

- Bumped `zine-layout` to `go-emrichen v0.0.11` and pushed commit `a8b2cba` after local logcopter, glazed-lint, tests, and CI-pinned golangci-lint passed.
- Did not create a release because GitHub dependency scanning failed in gosec with 39 legacy findings unrelated to the dependency bump.
- Recorded the blocker in the SQLite tracker as `dependency_scanning_failed_after_bump`.
