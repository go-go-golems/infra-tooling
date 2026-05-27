---
Title: ggg rollout implementation phases and tasks
Ticket: INFRA-002
Status: active
Topics:
    - cli
    - automation
    - release
    - github
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: internal/cli/rollout
      Note: |-
        New Glazed/Cobra command group for ggg rollout
        New Glazed/Cobra rollout command group
    - Path: internal/cli/root.go
      Note: Registers the rollout command group
    - Path: pkg/rollout
      Note: |-
        New rollout package with config, inventory, validation, branch, PR, status, and report primitives
        New rollout package with config
    - Path: ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/12-ggg-rollout.yaml
      Note: |-
        INFRA-002 rollout configuration used to validate the new commands
        Rollout config used to validate the new commands
ExternalSources: []
Summary: Concrete phase/task breakdown and implementation status for the first ggg rollout slice.
LastUpdated: 2026-05-27T14:05:00-04:00
WhatFor: Track what was planned, what was implemented, what was validated, and what remains for future ggg rollout work.
WhenToUse: Use before continuing rollout automation implementation or reviewing the first implementation slice.
---


# ggg rollout implementation phases and tasks

## Executive summary

This document turns the `ggg rollout` design into a build plan and records the first implementation slice. The first slice is intentionally operational and useful without attempting to solve all rollout patching. It adds repository inventory, rollout YAML creation, validation execution, branch checks, PR push/open plumbing, combined status, and Markdown reporting.

The current implementation establishes the package boundaries and command patterns needed for later profile-specific patching. It does not yet implement `rollout plan` or `rollout apply` for Glazed-lint Makefile edits. Those remain future work because they need a more careful patch model and idempotence tests.

## Phase 0: Stabilize the rollout scope

Status: complete.

Tasks:

- [x] Use INFRA-002 as the source workflow because it contains real rollout scripts and real PRs.
- [x] Keep the first implementation generic rather than hard-coding Glazed linting into all commands.
- [x] Preserve existing `ggg pr` and `ggg batch` behavior instead of duplicating readiness logic.
- [x] Keep mutating operations behind explicit `--yes` or `--dry-run` behavior.
- [x] Keep rollout commands row-oriented through Glazed output.

Outcome:

- New command group: `ggg rollout`.
- New package: `pkg/rollout`.

## Phase 1: Rollout data model and YAML configuration

Status: complete.

Tasks:

- [x] Add `rollout.Config`.
- [x] Add selection model:
  - explicit include list;
  - module-requirement filters;
  - exclude list with reason.
- [x] Add validation command model.
- [x] Add pull-request output model.
- [x] Add release/readiness placeholders for future orchestration.
- [x] Add `LoadConfig` and `SaveConfig` helpers.
- [x] Add target resolution from explicit include list or inventory filtering.
- [x] Create an INFRA-002 rollout YAML fixture:
  - `scripts/12-ggg-rollout.yaml`.

Implemented files:

- `pkg/rollout/config.go`
- `internal/cli/rollout/init.go`

Validation:

```bash
go run ./cmd/ggg rollout init --help
go test ./pkg/rollout -count=1
```

## Phase 2: Repository inventory

Status: complete.

Tasks:

- [x] Walk a workspace for `go.mod` files.
- [x] Skip noisy directories:
  - `.git`
  - `node_modules`
  - `vendor`
  - `.cache`
  - `.bin`
  - `dist`
  - `build`
- [x] Parse module path and require versions from `go.mod`.
- [x] Filter by required module path, e.g. `github.com/go-go-golems/glazed`.
- [x] Detect Makefile presence.
- [x] Detect lint-related Makefile targets:
  - `lint`
  - `lintmax`
  - `glazed-lint`
  - `glazed-lint-build`
- [x] Detect GitHub workflow and lefthook presence.
- [x] Detect Go package directories.
- [x] Inspect git branch, ahead count, and dirty state.
- [x] Expose as `ggg rollout inventory`.

Implemented files:

- `pkg/rollout/inventory.go`
- `internal/cli/rollout/inventory.go`
- `pkg/rollout/inventory_test.go`

Validation:

```bash
go run ./cmd/ggg rollout inventory \
  --root /home/manuel/workspaces/2026-05-24/add-js-providers \
  --require-module github.com/go-go-golems/glazed \
  --output json
```

## Phase 3: Cross-repository validation runner

Status: complete.

Tasks:

- [x] Load validation commands from rollout YAML.
- [x] Run each command in each target repository.
- [x] Continue after failures when configured.
- [x] Write one log per repository and command.
- [x] Emit one row per validation result.
- [x] Return exit code `4` when validation rows fail.
- [x] Add dry-run mode.
- [x] Validate INFRA-002 with `make glazed-lint` in all ten target repositories.

Implemented files:

- `pkg/rollout/validate.go`
- `internal/cli/rollout/validate.go`
- `pkg/rollout/validate_test.go`

Validation:

```bash
ggg rollout validate \
  ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/12-ggg-rollout.yaml \
  --output json
```

Result:

- All ten INFRA-002 target repositories passed `make glazed-lint`.
- Captured output:
  - `sources/15-ggg-rollout-validate.json`
- Captured per-repo logs:
  - `sources/ggg-rollout-logs/`

## Phase 4: Branch hygiene checks

Status: complete.

Tasks:

- [x] Inspect each target branch against configured branch and base.
- [x] Report current branch.
- [x] Report ahead count from base.
- [x] Report dirty tracked/untracked state.
- [x] Add `ggg rollout branch` status mode.
- [x] Add guarded `ggg rollout branch --commit --yes` helper for future use.
- [x] Fix branch commit staging to ignore missing workflow directories.

Implemented files:

- `pkg/rollout/branch.go`
- `internal/cli/rollout/branch.go`

Validation:

```bash
ggg rollout branch scripts/12-ggg-rollout.yaml --output json
```

Result:

- After squashing Codex feedback fix commits, all INFRA-002 target branches are one commit ahead of `origin/main` again.

## Phase 5: PR push/open plumbing

Status: implemented but not used for INFRA-002 PR creation.

Tasks:

- [x] Add `ggg rollout push-prs`.
- [x] Require `--yes` unless `--dry-run` is set.
- [x] Verify branch name before pushing.
- [x] Refuse branches with more than one ahead commit.
- [x] Support `--no-verify-push` with required reason for real pushes.
- [x] Use `gh pr create` for PR creation.
- [x] Write a `ggg batch ready` compatible PR YAML file.

Implemented files:

- `pkg/rollout/pushprs.go`
- `internal/cli/rollout/push_prs.go`

Validation:

```bash
go test ./pkg/rollout -count=1
ggg rollout push-prs scripts/12-ggg-rollout.yaml --dry-run --output json
```

Note:

- INFRA-002 PRs already existed before this command was implemented, so the command was not used to create those PRs.

## Phase 6: Combined rollout status

Status: complete.

Tasks:

- [x] Combine local branch state with remote PR readiness.
- [x] Reuse the existing PR list YAML.
- [x] Reuse `ghclient.Readiness` and `prready.Report`.
- [x] Emit one status row per target repository.
- [x] Return non-zero when any branch or PR is not OK.
- [x] Use the command to detect Codex feedback on six INFRA-002 PRs.
- [x] Use the command again after fixes and squashing.

Implemented files:

- `pkg/rollout/status.go`
- `internal/cli/rollout/status.go`

Validation:

```bash
ggg rollout status scripts/12-ggg-rollout.yaml --output json
```

Captured outputs:

- `sources/16-ggg-rollout-status.json`
- `sources/18-ggg-rollout-status-after-fixes.json`
- `sources/20-ggg-rollout-status-after-squash.json`

Latest observed status:

- Ready:
  - `css-visual-diff`
  - `discord-bot`
  - `glazed`
  - `workspace-manager`
- Waiting for new checks/Codex after fix push:
  - `geppetto`
  - `go-go-goja`
  - `goja-git`
  - `go-minitrace`
  - `loupedeck`
  - `pinocchio`

## Phase 7: Rollout reporting

Status: complete.

Tasks:

- [x] Generate a Markdown report from rollout config.
- [x] Include target repository table.
- [x] Include branch check summary.
- [x] Include validation commands and log directory.
- [x] Include PR YAML when present.
- [x] Expose as `ggg rollout report`.

Implemented files:

- `pkg/rollout/report.go`
- `internal/cli/rollout/report.go`

Validation:

```bash
ggg rollout report scripts/12-ggg-rollout.yaml --write-to sources/rollout-report.md
```

## Phase 8: INFRA-002 PR feedback repair discovered by rollout status

Status: complete for pushed fixes; remote checks still pending.

Tasks:

- [x] Run `ggg rollout status` and discover six PRs had current-head Codex feedback.
- [x] Inspect feedback with `ggg pr codex-comments`.
- [x] Fix `GLAZED_LINT_DIRS` not being honored in:
  - `geppetto`
  - `pinocchio`
- [x] Pin `glazed-lint` tool installation to `v1.3.4` in older Glazed-dependent repos:
  - `go-go-goja`
  - `goja-git`
  - `go-minitrace`
  - `loupedeck`
- [x] Validate `make glazed-lint` in all six fixed repositories.
- [x] Push fixes to existing PR branches.
- [x] Squash each fixed branch back to one commit ahead of `origin/main`.
- [x] Force-push with lease to keep branches focused.
- [x] Retrigger Codex after the final squashed heads.
- [x] Capture status after squashing.

Latest branch heads after the final amendment pass:

| Repo | Branch head |
| --- | --- |
| `css-visual-diff` | `fb0f6ee Run Glazed CLI policy linting` |
| `discord-bot` | `4cb3eeb Run Glazed CLI policy linting` |
| `geppetto` | `bf477e63 Run Glazed CLI policy linting` |
| `glazed` | `f1e9091 Run Glazed CLI policy linting` |
| `go-go-goja` | `dfdddfd Run Glazed CLI policy linting` |
| `goja-git` | `587aeb3 Run Glazed CLI policy linting` |
| `go-minitrace` | `e434a48 Run Glazed CLI policy linting` |
| `loupedeck` | `85b10f6 Run Glazed CLI policy linting` |
| `pinocchio` | `b5442c7 Run Glazed CLI policy linting` |
| `workspace-manager` | `e0106ef Run Glazed CLI policy linting` |

## Phase 9: Installed binary

Status: complete.

Tasks:

- [x] Run all infra-tooling tests.
- [x] Build and install the updated binary to `~/.local/bin/ggg`.
- [x] Verify `ggg rollout --help` shows the new command group.

Validation:

```bash
go test ./...
go build -o ~/.local/bin/ggg ./cmd/ggg
~/.local/bin/ggg rollout --help
```

## Phase 10: Final Makefile hardening after second Codex round

Status: complete locally and pushed; remote checks are pending.

Tasks:

- [x] Replace `@latest` fallback installs with explicit `GLAZED_LINT_TOOL_VERSION ?= v1.3.4`.
- [x] Prefix Glazed vettool invocations with `GOWORK=off` so standalone `make glazed-lint` does not use ambient parent workspaces.
- [x] Apply the hardening consistently across all ten rollout PR branches.
- [x] Validate all ten repositories with `ggg rollout validate`.
- [x] Amend each rollout branch back to one focused commit ahead of `origin/main`.
- [x] Force-push with lease.
- [x] Retrigger Codex on the final amended heads.
- [x] Capture final rollout status.

Artifacts:

- `sources/23-ggg-rollout-validate-after-codex-round2.json`
- `sources/26-codex-retrigger-final-amended-heads.json`
- `sources/27-ggg-rollout-status-final-amended-heads.json`

Latest observed status:

- All ten local branches are clean, on `infra-002/glazed-lint`, and exactly one commit ahead of `origin/main`.
- All ten PRs are waiting for checks/Codex on the final amended heads.
- No PR was merged.

## Remaining future work

These tasks remain intentionally out of scope for the first implementation slice:

- [ ] Add `ggg rollout plan` for profile-specific patch planning.
- [ ] Add `ggg rollout apply --profile glazed-lint` with idempotent Makefile/workflow patch operations.
- [ ] Add typed patch-operation diff output.
- [ ] Add diagnostic parser for `glazed-lint` logs and allow-path suggestions.
- [ ] Add fake GitHub client tests for `push-prs`.
- [ ] Add temp-git-repo tests for `branch --commit`.
- [ ] Add raw GraphQL pagination in `ghclient` so readiness is not limited to the first page.
- [ ] Add command-output fixture tests for the new rollout commands.
- [ ] Add a merge command only if a future ticket explicitly authorizes merge automation.

## Implemented command summary

```text
ggg rollout inventory
ggg rollout init
ggg rollout validate
ggg rollout branch
ggg rollout push-prs
ggg rollout status
ggg rollout report
```

## Review checklist

- Start with `pkg/rollout/config.go` to understand the YAML model.
- Read `pkg/rollout/inventory.go` for repository discovery and git facts.
- Read `pkg/rollout/validate.go` for cross-repo command execution.
- Read `pkg/rollout/status.go` for the connection to PR readiness.
- Read `internal/cli/rollout/*.go` to verify each operation is exposed as a Glazed command.
- Run `go test ./...`.
- Run `ggg rollout inventory --root /home/manuel/workspaces/2026-05-24/add-js-providers --require-module github.com/go-go-golems/glazed`.
- Run `ggg rollout branch scripts/12-ggg-rollout.yaml`.
