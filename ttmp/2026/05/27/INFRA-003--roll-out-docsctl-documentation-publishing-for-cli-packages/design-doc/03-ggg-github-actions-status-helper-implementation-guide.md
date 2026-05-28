---
Title: ggg GitHub Actions Status Helper Implementation Guide
Slug: ggg-github-actions-status-helper-implementation-guide
DocType: design-doc
Topics:
  - docsctl
  - release-train
  - automation
  - github-actions
---

# ggg GitHub Actions Status Helper Implementation Guide

## Goal

Add two small `ggg` helpers for release-train operators who need to answer one question quickly: “did the post-merge or post-tag GitHub Actions runs finish, and are the only failures the failures we deliberately tolerate?”

The helpers should replace ad-hoc `gh run list | jq ...` loops while staying thin wrappers around GitHub CLI data.

## Commands

### `ggg run status`

Single repository / single commit status checker.

Example:

```bash
ggg run status \
  --repo go-go-golems/go-go-goja \
  --branch main \
  --sha 2a5b79c \
  --ignore-workflow "Secret Scanning" \
  --output table
```

Responsibilities:

- call `gh run list` for one repository;
- filter runs by branch and optional SHA prefix;
- emit one row per workflow run;
- classify each row as `ok`, `ignored_failure`, `failed`, `pending`, or `other`;
- return non-zero when any non-ignored failure exists, and a distinct non-zero code when runs are still pending.

### `ggg batch actions`

Batch status checker from a YAML manifest.

Example manifest:

```yaml
repos:
  - repo: go-go-golems/css-visual-diff
    branch: main
    sha: 8559422
  - repo: go-go-golems/go-go-goja
    branch: main
    sha: 2a5b79c
```

Example command:

```bash
ggg batch actions /tmp/logcopter-actions.yaml \
  --ignore-workflow "Secret Scanning" \
  --ignore-workflow "Dependency Graph" \
  --output json
```

Responsibilities:

- load a YAML list of repositories;
- call the same collection/classification path used by `ggg run status`;
- emit per-run rows and a final summary row;
- support `--summary-only` for operator dashboards;
- support `--watch` with interval/timeout so release trains can wait until no pending runs remain.

## Exit codes

Use the existing `exitcode.Request` pattern.

- `0`: all matching runs are successful, skipped, neutral, or explicitly ignored.
- `1`: at least one matching run failed, was cancelled, reported an unknown terminal conclusion, and was not ignored.
- `2`: at least one matching run is still queued/in-progress, or no matching runs have appeared yet, and there are no non-ignored failures yet.

This mirrors the release-train workflow: failures require action; pending checks require waiting; ignored failures are recorded but do not block.

## Ignore model

Start with simple workflow-name ignores. That covers the current release-train need: Secret Scanning failures are expected noise for several repositories.

Flags:

```bash
--ignore-workflow "Secret Scanning"
--ignore-workflow "Dependency Graph"
```

A future extension can add `--ignore-conclusion` or per-repo ignore rules if needed, but workflow-name ignores are enough for the immediate logcopter rollout.

## Output schema

Per-run rows:

- `repo`
- `branch`
- `sha`
- `workflow`
- `status`
- `conclusion`
- `classification`
- `ignored`
- `url`
- `created_at`

Summary row:

- `repo: summary`
- `ok`
- `failed`
- `ignored_failures`
- `pending`
- `success`
- `total`
- `state`

## Implementation plan

1. Add a small package for action-run collection/classification.
2. Add `internal/cli/run` with `ggg run status`.
3. Add `ggg batch actions` under the existing batch command group.
4. Add tests for classification and manifest parsing.
5. Update release-train playbooks to use the new helpers before tagging/bumping downstream.
6. Validate against the saved logcopter main-action status set.

## Release-train use

After a generated-logcopter PR is merged, operators should run:

```bash
ggg run status --repo go-go-golems/<repo> --branch main --sha <merge-sha> \
  --ignore-workflow "Secret Scanning"
```

For a multi-repo rollout, keep a manifest next to the ticket evidence and run:

```bash
ggg batch actions ttmp/.../scripts/<rollout-actions>.yaml \
  --ignore-workflow "Secret Scanning" \
  --watch
```

Only proceed to release tags and downstream dependency bumps after non-ignored failures are gone.
