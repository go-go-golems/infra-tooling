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
