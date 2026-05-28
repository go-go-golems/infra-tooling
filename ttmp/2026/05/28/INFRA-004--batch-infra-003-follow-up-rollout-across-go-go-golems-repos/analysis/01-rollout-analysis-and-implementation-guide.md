---
Title: Rollout analysis and implementation guide
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
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/analysis/01-repository-follow-up-inventory-for-logcopter-docsctl-glazed-linting-and-xgoja.md
      Note: Source analysis that defines the follow-up tracks.
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/41-repository-follow-up-inventory.json
      Note: |-
        Source machine-readable inventory from INFRA-003.
        Source machine-readable inventory for INFRA-004 batching
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/01-plan-rollout-batches.py
      Note: |-
        Script that regenerates the INFRA-004 batch plan from INFRA-003 artifacts.
        Batch planner script
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/01-rollout-batches.json
      Note: |-
        Machine-readable batch plan derived from the INFRA-003 inventory and local go.mod dependency inspection.
        Generated machine-readable rollout batches
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/02-rollout-batches.tsv
      Note: Tabular batch plan for quick filtering and PR manifest construction.
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/03-rollout-batches.md
      Note: Human-readable batch plan grouped by dependency and rollout risk.
ExternalSources: []
Summary: Implementation guide for batching INFRA-003 follow-up work into dependency-aware PR waves with parallel review/CI and sequential dependency releases.
LastUpdated: 2026-05-28T00:00:00-04:00
WhatFor: Use this document to drive INFRA-004 rollout PR creation, validation, merge, main-action verification, release tagging, and diary/changelog capture.
WhenToUse: Before opening or reviewing any INFRA-004 follow-up PR across Go-Go-Golems repositories.
---


# Rollout analysis and implementation guide

## Goal

INFRA-004 turns the INFRA-003 follow-up inventory into an executable rollout plan. The work is intentionally batched: create one focused PR per repository, but move related repositories through branch creation, validation, PR readiness, Codex feedback, and action watching in parallel. Dependency releases remain sequential.

The four tracks are:

1. **logcopter baseline** — add generated package loggers plus `make logcopter-check`.
2. **docsctl publishing** — add/enable release docs publishing only after local help export and validation work.
3. **Glazed linting** — add `make glazed-lint` and hook CI where the repository imports Glazed.
4. **xgoja providers** — do not mechanically add bindings; first confirm API intent for each provider candidate.

## Source inventory

Primary inputs:

- INFRA-003 analysis: `../INFRA-003.../analysis/01-repository-follow-up-inventory-for-logcopter-docsctl-glazed-linting-and-xgoja.md`.
- Machine-readable inventory: `sources/41-repository-follow-up-inventory.json`.
- Compact inventory: `sources/42-repository-follow-up-inventory.tsv`.
- Docsctl rollout guide: `design-doc/01-docsctl-publishing-rollout-analysis-and-implementation-guide.md`.
- GitHub Actions status helper guide: `design-doc/03-ggg-github-actions-status-helper-implementation-guide.md`.
- Playbooks: package release train, logcopter rollout, and docsctl publishing rollout.

INFRA-003 scanned 77 Go repositories and found 70 with at least one follow-up flag:

| Track | Candidate count |
|---|---:|
| logcopter addition | 65 |
| docsctl CI/CD + publish | 39 |
| Glazed linting | 49 |
| xgoja provider bindings | 11 |

## Batch generation

`INFRA-004/scripts/01-plan-rollout-batches.py` reads the INFRA-003 JSON inventory, inspects local `go.mod` files under `/home/manuel/code/wesen/go-go-golems`, and writes:

- `sources/01-rollout-batches.json`
- `sources/02-rollout-batches.tsv`
- `sources/03-rollout-batches.md`

Regenerate with:

```bash
python3 infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/01-plan-rollout-batches.py
```

## Operational batches

The generated batches are risk/dependency buckets, not a mandate to merge every repo in a bucket before starting the next. Within a bucket, keep one branch and PR per repository.

### B1 — foundation and upstream libraries

Run first where needed as dependencies. Do not batch tightly coupled downstream consumers until upstream release tags are available.

Representative repos: `logcopter`, `common-sense`, `dmeta`, `esper`, `go-sqlite-regexp`, `infra-tooling`, `sanitize`, plus upstream-like repos detected by first-party downstream usage.

### B2 — leaf logcopter-only repositories

Lowest rollout risk. These should be the first practical branch/PR wave for mechanical logcopter adoption because they do not require docsctl/Vault decisions and mostly do not gate downstream dependency releases.

Representative repos: `ai-in-action-app`, `barbar`, `go-go-agent-action`, `go-go-app-sqlite`, `oak-git-db`, `salad`, `voyage`.

### B3 — Glazed linting without docsctl

Add `glazed-lint` and logcopter baseline where safe. No docs publishing workflow should be added in this batch.

Representative repos: `markdown-quizz`, `plunger`, `refactorio`, `go-go-app-inventory`.

### B4 — docsctl + Glazed CLI leaf packages

Add docsctl publishing only after confirming `help export --format sqlite` and `docsctl validate --package <package> --version <version> --file <file>`. This batch may need Terraform/Vault roles outside individual repos.

Representative repos: `almanach`, `biberon`, `bucheron`, `devctl`, `docmgr`, `plz-confirm`, `remarquee`, `zine-layout`, and the other docsctl candidates in `sources/03-rollout-batches.md`.

### B5 — xgoja provider/API-intent candidates

Do not mechanically generate provider bindings. Confirm the JavaScript API shape first. Repos in this bucket may still receive safe logcopter/glazed baseline work, but provider work must remain separate or explicitly scoped.

Representative repos: `go-go-goja`, `go-minitrace`, `pinocchio`, `workspace-manager`, `cozodb-goja`, `goja-github-actions`, `scraper`, `smailnail`, `vm-system`.

## Per-repository PR shape

Use a consistent branch name unless a repo already has a conflicting branch:

```text
infra/baseline-rollout
```

One PR per repo. Combine all safe baseline work for that repo:

- generated logcopter logger files and `make logcopter-check`;
- `make glazed-lint` and CI hook if it imports Glazed;
- docsctl export/publish workflow only when local export is proven and Vault role intent is clear;
- xgoja provider bindings only after explicit API confirmation.

Never push directly to `main`. Merge using:

```bash
gh pr merge <n> --merge --delete-branch
```

No squash merges.

## Validation loop

Run locally per repo:

```bash
make logcopter-check || true
make glazed-lint || true
GOWORK=off go test ./...
ggg release preflight --repo . --output json
```

The first two commands are allowed to be absent during early branches; missing targets must be interpreted in context. After a repo receives the target, absence is a blocker.

## PR readiness loop

Build a PR manifest after opening PRs and watch the batch:

```bash
ggg batch ready prs.yaml --watch --until actionable
```

For individual PRs:

```bash
ggg pr ready <pr> --findings --output json
ggg pr watch <pr> --interval-seconds 30 --timeout-seconds 1800 --output json
```

Wait 20–30 seconds for automatic Codex before manual trigger:

```bash
ggg pr codex-trigger <pr> --wait-for-auto 30s
```

Do not retrigger Codex if the current head already has a satisfied Codex signal.

## Main-action and release loop

After merge, verify main branch Actions in parallel with a manifest:

```bash
ggg batch actions actions.yaml \
  --ignore-workflow "Secret Scanning" \
  --watch \
  --output json
```

Release dependency roots sequentially. Once independent repos are tagged, watch workflows in parallel. Do not retag failed releases; fix and create a new patch/minor tag.

## Recording requirements

Keep the INFRA-004 diary current and also update INFRA-003 changelog/diary when the original rollout ticket needs continuity. Each completed repo should record:

- PR URL;
- merge SHA;
- main action workflow URL/status;
- release tag;
- release workflow URL;
- docs URL if docsctl was enabled;
- failures and fixes.

Run hygiene checks before considering the work complete:

```bash
docmgr doctor --ticket INFRA-004 --stale-after 30
docmgr doctor --ticket INFRA-003 --stale-after 30
```
