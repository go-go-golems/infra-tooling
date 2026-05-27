---
Title: PR readiness with ggg
Ticket: PR-REVIEW-READY-001
Status: active
Topics:
    - automation
    - github
    - cicd
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../cmd/ggg/main.go
      Note: Installed CLI entry point for PR readiness operations
    - Path: ../../../internal/cli/pr/ready.go
      Note: `ggg pr ready` command implementation
    - Path: ../../../internal/cli/pr/codex_trigger.go
      Note: `ggg pr codex-trigger` command implementation
    - Path: ../../../internal/cli/pr/codex_comments.go
      Note: `ggg pr codex-comments` command implementation
    - Path: ../../../internal/cli/batch/ready.go
      Note: `ggg batch ready` command implementation
    - Path: ../../../pkg/prready/prready.go
      Note: Readiness state machine shared by single-PR and batch commands
    - Path: ../../../pkg/prready/testdata
      Note: Snapshot fixtures for ready, failed-check, and Codex-feedback states
ExternalSources:
    - https://github.com/go-go-golems/pinocchio/pull/158
Summary: Usage notes for the installed `ggg` CLI commands that decide whether a PR is ready based on completed checks and Codex review signals.
LastUpdated: 2026-05-27T03:45:00-04:00
WhatFor: Use this when batching PR readiness checks across many repositories.
WhenToUse: Before merging rollout PRs that require both green CI and a satisfied Codex review.
---

# PR readiness with `ggg`

## Executive summary

Use the installed `ggg` CLI for go-go-golems pull-request readiness. The old shell/Python scripts in `scripts/go-go-golems/` are historical references; new playbooks and operator workflows should call `ggg` directly.

`ggg` checks the same merge gate we use in release trains:

1. GitHub mergeability is clean (no conflicts / blocked merge state);
2. status checks exist;
3. every check/status is successful, skipped, or neutral;
4. a Codex signal exists;
5. Codex is not still running (`EYES` reaction);
6. Codex did not leave current-head substantive review comments;
7. stale Codex feedback from older commits does not block the current head.

The commands emit concise table output by default and row-oriented structured output with Glazed flags such as `--output json`, `--output yaml`, or `--output csv`.

## Command overview

```text
ggg pr ready <pr>                  # classify one PR
ggg pr ready <pr> --findings       # include finding rows for debugging
ggg pr codex-trigger <pr> --wait-for-auto 30s  # wait for automatic Codex before manual trigger
ggg pr codex-trigger --file prs.yaml --wait-for-auto 30s
ggg pr codex-comments <pr>         # list Codex review bodies and inline comments
ggg pr watch <pr>                  # poll one PR until ready or actionable
ggg batch ready prs.yaml           # classify many PRs
ggg batch ready prs.yaml --watch --until actionable  # poll until there is operator work
```

The CLI should already be installed on the operator PATH. If you are testing from a checkout before installation, use `go run ./cmd/ggg ...` from the infra-tooling repository.

## PR list format

Batch commands use YAML rather than ad-hoc newline files:

```yaml
prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2
  - ref: go-go-golems/go-minitrace#11
```

Keep release-train PR lists in the active ticket's `scripts/` directory so they are reviewable and reusable.

## Single-PR workflow

Check readiness:

```bash
ggg pr ready https://github.com/go-go-golems/<repo>/pull/<number>
```

Get machine-readable output:

```bash
ggg pr ready https://github.com/go-go-golems/<repo>/pull/<number> --output json
```

Show detailed findings when a PR is not ready:

```bash
ggg pr ready https://github.com/go-go-golems/<repo>/pull/<number> --findings
```

After opening a PR, let `ggg` wait briefly before manually triggering Codex. GitHub/Codex often starts an automatic review within 20-30 seconds, and an immediate manual trigger can create duplicate review runs.

Trigger Codex review only if no review appears after that short delay, no review is running, no satisfied signal already exists, and current-head feedback is not already present:

```bash
ggg pr codex-trigger https://github.com/go-go-golems/<repo>/pull/<number> --wait-for-auto 30s
```

Safety behavior:

- If the latest signal has an `EYES` reaction, `ggg pr codex-trigger` skips by default.
- If a human recently posted `@codex review`, it skips by default to avoid duplicate trigger spam.
- If current-head Codex feedback already exists, it skips by default.
- If Codex is already satisfied (thumbs-up reaction or satisfied Codex body) for the current head, it skips by default.
- Use `--force` only when you intentionally want a new Codex pass despite those guards.
- Use `--dry-run --output json` to inspect what would happen without posting a comment.

Inspect Codex feedback directly:

```bash
ggg pr codex-comments https://github.com/go-go-golems/<repo>/pull/<number>
```

This command emits Codex-authored review bodies and inline comments with reviewed commit, current/stale status, path, line, body, and URL.

## Batch workflow

Trigger Codex for many PRs:

```bash
ggg pr codex-trigger --file /path/to/prs.yaml --wait-for-auto 30s
```

Check many PRs once:

```bash
ggg batch ready /path/to/prs.yaml
```

Watch one PR until it is ready or actionable:

```bash
ggg pr watch https://github.com/go-go-golems/<repo>/pull/<number> --interval-seconds 30 --timeout-seconds 1800
```

Watch a batch until there is something for the operator to do:

```bash
ggg batch ready /path/to/prs.yaml --watch --until actionable --interval-seconds 30 --timeout-seconds 1800
```

Watch mode defaults to `--until actionable`, which stops when:

- every PR is ready;
- any PR reaches terminal Codex feedback;
- any PR has failed checks;
- any PR has merge conflicts or blocked mergeability;
- any PR reports an API/tool error;
- or some PR becomes ready while others are still waiting.

The last case exits with code `5` so release-train operators can merge/release ready repositories in dependency order instead of sleeping through actionable progress.

Alternative stop modes:

- `--until all-ready`: keep polling through partial readiness; stop only when all PRs are ready or a terminal blocker appears.
- `--until terminal`: stop on all-ready or terminal blockers, not on partial readiness.
- `--until first-ready`: stop as soon as at least one PR is ready or a terminal blocker appears.

## Exit codes

Use a built/installed `ggg` binary when exact process status matters. `go run` wraps program exits.

| Code | Meaning |
| --- | --- |
| `0` | Ready / all ready. |
| `1` | Not ready yet, usually waiting for checks or Codex. |
| `2` | Tool/API error. |
| `3` | Current-head Codex feedback requires operator action. |
| `4` | Failed checks require operator action. |
| `5` | Batch partial readiness: at least one PR is ready while others still wait. |
| `6` | Merge conflicts or blocked mergeability require operator action. |

## Readiness states

Known `state` values include:

- `ready`
- `waiting_checks`
- `waiting_codex`
- `no_codex`
- `failed_checks`
- `merge_conflict`
- `codex_feedback`
- `not_ready`
- `error`

`terminal=true` means waiting alone will not make the PR mergeable; a human/code change is needed.

## Implementation notes

`ggg` uses GitHub GraphQL through `gh` and decodes:

- `statusCheckRollup.contexts.nodes` for check runs and legacy status contexts;
- PR reviews and PR comments for Codex signals;
- review inline comments for actual code-review feedback;
- `reactionGroups` for `THUMBS_UP` and `EYES` reactions;
- reviewed commit markers in Codex bodies so stale feedback does not block the current head.

The readiness state machine lives in `pkg/prready` and has snapshot fixtures under `pkg/prready/testdata` for ready, failed checks, current-head Codex feedback, running Codex, stale feedback, and truncated feedback cases.

## Historical scripts

The old scripts remain in `scripts/go-go-golems/` as historical references and for environments where `ggg` has not been installed yet:

- `00-pr-ready-check.sh`
- `01-pr-ready-check.py`
- `02-trigger-codex-review.sh`
- `03-watch-codex-reactions.py`
- `04-wait-pr-ready.sh`
- `05-batch-pr-ready.sh`
- `06-batch-trigger-codex-review.sh`

Do not add new playbook examples that call these scripts. Prefer the installed `ggg` commands above.
