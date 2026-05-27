---
Title: GGG Docsctl Rollout Automation Implementation Guide
Ticket: INFRA-003
Status: active
Topics:
  - cli
  - automation
  - release
  - github
  - docsctl
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
  - /home/manuel/code/wesen/go-go-golems/infra-tooling/docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md
  - /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/batch/ready.go
  - /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/codex_comments.go
  - /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/rollout/root.go
Summary: Implementation guide for the next docsctl rollout ergonomics improvements: playbook fixes, batch readiness reports, grouped Codex feedback, and ggg rollout docsctl inventory/validate/plan.
LastUpdated: 2026-05-27T18:10:00-04:00
WhatFor: Guide implementation of ggg and playbook improvements discovered during the INFRA-003 docsctl rollout.
WhenToUse: Before changing ggg batch/Codex/rollout commands or the docsctl publishing playbook.
---

# GGG Docsctl Rollout Automation Implementation Guide

## Executive summary

The first INFRA-003 docsctl rollout required repeated manual work: patching seven workflows, correcting overly broad OIDC permissions after Codex feedback, opening PRs, summarizing readiness, grouping repeated Codex comments, and writing ad-hoc inventory scripts. This guide turns the first four improvement ideas into concrete implementation work.

The work has four parts:

1. **Patch the docsctl playbook** so it reflects the correct `docsctl validate --package --version --file` contract, recommends job-level OIDC permissions, forces a package identity checklist, and warns about exact `release.yaml` versus `release.yml` workflow refs.
2. **Improve `ggg batch ready` reporting** with `--summary-only` and `--markdown-report`, so operators can answer “where are we?” without reading verbose JSON.
3. **Add grouped batch Codex comments** with `ggg batch codex-comments --group-by-message`, so repeated Codex findings across rollout PRs can be fixed once across all repos.
4. **Add a docsctl rollout profile** with `ggg rollout docsctl inventory`, `validate`, and `plan`, so future docsctl publishing rollouts start from a repeatable candidate discovery and plan format instead of ticket-local scripts.

## Current-state map

### Playbook

The docsctl playbook lives at:

```text
docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md
```

It already explains the release-only docs publishing architecture, reusable workflow, Vault roles, and registry verification. The current improvement is to align its examples with what the code now requires and what Codex recommended during the rollout.

### Batch readiness

`ggg batch ready` lives at:

```text
internal/cli/batch/ready.go
```

It currently emits one row per PR plus a summary row. It already computes useful fields like `state`, `next_action`, `pending_checks`, `failed_checks`, and `merge_state_status`, but operators still need to mentally summarize the output. `--summary-only` should suppress per-PR detailed rows and emit grouped categories. `--markdown-report` should emit a copy/paste-ready Markdown report.

### Codex comments

Single-PR Codex comments live at:

```text
internal/cli/pr/codex_comments.go
```

The batch command namespace currently only has `ready`. A new `ggg batch codex-comments <prs.yaml>` command can reuse the same `ghclient.Client{}.Snapshot` and `prready.SortedSignals` primitives as the single-PR command, but add grouping across PRs.

### Rollout commands

Rollout commands are registered in:

```text
internal/cli/rollout/root.go
```

The existing rollout commands operate on a general rollout YAML. Docsctl needs a profile-specific subcommand group under `ggg rollout docsctl` that can run without editing repos:

```bash
ggg rollout docsctl inventory --workspace /path/to/workspace
ggg rollout docsctl validate --workspace /path/to/workspace
ggg rollout docsctl plan --workspace /path/to/workspace --output yaml
```

## Proposed UX

### Playbook changes

Validation examples should use:

```bash
docsctl validate \
  --file .docsctl/help.sqlite \
  --package <package> \
  --version v0.0.0-local
```

Workflow templates should scope OIDC permissions to the reusable workflow job:

```yaml
permissions:
  contents: write

jobs:
  publish-docs:
    name: Publish docs
    permissions:
      contents: read
      id-token: write
    uses: go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@main
```

Add a package identity checklist:

```markdown
- Public docs package name:
- Exporting binary / command directory:
- Export command:
- Release workflow path:
- Numeric GitHub repository ID:
- Vault role name:
- Should examples/demo CLIs be excluded?
- Does the package already appear in /api/packages?
```

### `ggg batch ready --summary-only`

Example:

```bash
ggg batch ready scripts/10-docsctl-publishing-prs.yaml --summary-only
```

Rows should be grouped by category:

```text
category          repository                       pr   state            next_action
summary           summary                              partial_ready
ready             go-go-golems/discord-bot         11   ready            merge_when_manual_review_allows
waiting_checks    go-go-golems/css-visual-diff     10   waiting_checks   wait_for_pending_checks
```

### `ggg batch ready --markdown-report`

Example:

```bash
ggg batch ready scripts/10-docsctl-publishing-prs.yaml --markdown-report > status.md
```

Output should be raw Markdown:

```markdown
## Batch readiness

- Ready: 1
- Waiting checks: 6
- Codex feedback: 0
- Failed checks: 0
- Merge conflicts: 0
- Errors: 0

### Ready
- discord-bot PR 11 — merge_when_manual_review_allows

### Waiting checks
- css-visual-diff PR 10 — Analyze, test, GoSec Security Scan
```

Implementation detail: `--markdown-report` is easiest as a direct Cobra command path or by letting the command print to stdout and return before using the Glazed processor. If keeping the Glazed processor, emit a `markdown` field, but raw stdout is the intended UX.

### `ggg batch codex-comments --group-by-message`

Example:

```bash
ggg batch codex-comments scripts/10-docsctl-publishing-prs.yaml --group-by-message
```

Rows:

```text
count  title                                      prs
2      Scope OIDC permission to docs publishing  go-go-goja#43,pinocchio#162
```

Grouping key:

1. Prefer the first bold title in the Codex comment body.
2. Fall back to the first non-empty line.
3. Normalize whitespace and strip Markdown badge noise where practical.

### `ggg rollout docsctl inventory/validate/plan`

`inventory` discovers command candidates:

```pseudo
for each repo under workspace:
  if repo/go.mod missing: skip
  for each cmd/*/main.go:
    record repo, cmd_dir, package_name default repo basename
```

`validate` runs export + docsctl validate:

```pseudo
for each candidate:
  sqlite = tempdir/repo/cmd/help.sqlite
  run GOWORK=off go run ./cmd/<name> help export --format sqlite --output-path sqlite
  if sqlite missing or empty: status=export_failed
  else run docsctl validate --file sqlite --package package --version v0.0.0-local
  status = validate_ok or validate_failed
```

`plan` emits a YAML plan with selected validated candidates:

```yaml
profile: docsctl
workspace: /home/manuel/workspaces/2026-05-24/add-js-providers
repositories:
  - name: css-visual-diff
    path: /...
    package_name: css-visual-diff
    workflow: .github/workflows/release.yaml
    export_command: GOWORK=off go run ./cmd/css-visual-diff help export --format sqlite --output-path .docsctl/help.sqlite
    sqlite_path: .docsctl/help.sqlite
    vault_role: docsctl-css-visual-diff-publisher
    status: validate_ok
```

## Implementation phases

### Phase 1: Docs and task setup

- Add this guide.
- Add tasks to INFRA-003.
- Record diary Step 5.

### Phase 2: Playbook patch

- Update validation examples.
- Update reusable job template to job-level OIDC.
- Add package identity checklist.
- Add release workflow filename warning.
- Commit as a docs-only commit.

### Phase 3: Batch reporting

- Add `summary-only` and `markdown-report` flags to `internal/cli/batch/ready.go`.
- Refactor readiness collection so both normal rows and reports share the same data.
- Validate with the live INFRA-003 PR YAML.
- Commit.

### Phase 4: Batch Codex grouping

- Add `internal/cli/batch/codex_comments.go`.
- Register it in `internal/cli/batch/root.go`.
- Reuse PR list loading and Codex signal parsing.
- Validate against the current INFRA-003 PR YAML.
- Commit.

### Phase 5: Docsctl rollout profile

- Add `internal/cli/rollout/docsctl.go`.
- Register it in `internal/cli/rollout/root.go`.
- Implement inventory, validate, and plan subcommands.
- Validate against `/home/manuel/workspaces/2026-05-24/add-js-providers`.
- Commit.

### Phase 6: Diary, changelog, and delivery

- Update the diary after each substantial step.
- Update changelog and doc relations.
- Run `go test ./...`.
- Run `docmgr doctor --ticket INFRA-003 --stale-after 30`.
- Upload refreshed bundle to reMarkable if requested or if the guide changes materially.

## Testing strategy

Minimum validation:

```bash
go test ./...
ggg batch ready <prs.yaml> --summary-only
ggg batch ready <prs.yaml> --markdown-report
ggg batch codex-comments <prs.yaml> --group-by-message
ggg rollout docsctl inventory --workspace /home/manuel/workspaces/2026-05-24/add-js-providers
ggg rollout docsctl validate --workspace /home/manuel/workspaces/2026-05-24/add-js-providers --include css-visual-diff
ggg rollout docsctl plan --workspace /home/manuel/workspaces/2026-05-24/add-js-providers --include css-visual-diff
```

## Risks

- `--markdown-report` must not break existing Glazed output behavior for normal `batch ready`.
- `docsctl validate` can be slow because it runs `go run`; support `--include` to test small subsets.
- Multi-CLI repos still need human package identity decisions; automation should expose ambiguity rather than silently picking the wrong command.
- Workflow patching is intentionally not part of this implementation batch; first build inventory/reporting primitives.
