---
Title: Glazed linting rollout diary
Ticket: INFRA-002
Status: active
Topics:
  - cli
  - automation
  - release
  - github
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological diary for rolling out Glazed CLI policy linting across go-go-golems repositories.
LastUpdated: 2026-05-27T11:20:00-04:00
WhatFor: Preserve exact steps, commands, failures, validations, PRs, and release actions for the Glazed linting rollout.
WhenToUse: Read before resuming INFRA-002 or reviewing the rollout.
---

# Diary

## Goal

This diary records the rollout of Glazed CLI policy linting across go-go-golems repositories. It captures the repository inventory, generated scripts, Makefile/CI changes, diagnostics, fixes, validations, pull requests, readiness checks, merges, and releases.

## Step 1: Initialize the rollout ticket

I created a dedicated ticket before editing repositories because this work will touch multiple repos and may produce PRs, releases, and helper scripts. The ticket gives us one place to store inventory scripts, PR YAML files, validation notes, and the chronological diary.

The rollout will follow the infra-tooling Glazed linting playbook and use the newly installed `ggg` binary for Codex triggering, readiness checks, and release operations.

### Prompt Context

**User prompt (verbatim):** "ok, go ahead. keep a detailed diary as you work, so that we can trace all your steps, and keep all scripts you write in the scripts/ folder of the ticket."

**Assistant interpretation:** Start the Glazed lint rollout, create traceable ticket bookkeeping, and ensure all helper scripts are saved under the ticket scripts directory.

**Inferred user intent:** The user wants a repeatable multi-repository rollout rather than ad-hoc edits, with enough diary detail to audit or resume the work later.

**Commit (code):** N/A — ticket setup only.

### What I did
- Created docmgr ticket `INFRA-002 — Roll out Glazed CLI policy linting across go-go-golems repositories`.
- Created `design-doc/01-glazed-linting-rollout-plan.md`.
- Created `reference/01-glazed-linting-rollout-diary.md`.
- Replaced the default task list with a phased rollout checklist.
- Created the ticket `scripts/` directory.

### Why
- Multi-repository lint rollouts need a central record of exact commands and repository decisions.
- Helper scripts should be versioned in the ticket so later operators can rerun or improve them.

### What worked
- `docmgr ticket create-ticket` created the workspace.
- `docmgr doc add` created rollout plan and diary documents.

### What didn't work
- N/A.

### What I learned
- INFRA-001 established a useful pattern for this work: ticket scripts, YAML PR lists, `ggg` readiness, and detailed diary entries.

### What was tricky to build
- The target repository set is not yet fixed. I will first inventory local go-go-golems repositories that depend on Glazed, then record inclusions and exclusions before changing code.

### What warrants a second pair of eyes
- Whether the target set should include every local repo that depends on Glazed, or only the recently active repos in the xgoja workspace.

### What should be done in the future
- Add and run a repository inventory script under this ticket's `scripts/` directory.
- Use the inventory to decide rollout order and exclusions.

### Code review instructions
- Review `tasks.md` for rollout scope.
- Review upcoming inventory script before using it to drive repository edits.

### Technical details

Ticket path:

```text
/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories
```

## Step 2: Inventory local Glazed-dependent repositories

The second step added and ran the repository inventory script. The inventory is deliberately broad: it scans the recent xgoja workspace and the canonical `/home/manuel/code/wesen/go-go-golems` tree, then reports repositories whose `go.mod` mentions `github.com/go-go-golems/glazed`.

The first result shows that this rollout is larger than the recent xgoja set. Many local go-go-golems repositories depend on Glazed, and some already have `glazed-lint` targets. The rollout therefore needs exclusions and prioritization rather than editing every discovered path blindly.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue by creating traceable scripts under the ticket and using them to drive the rollout.

**Inferred user intent:** The user wants every automation step preserved in the ticket, including discovery scripts.

**Commit (code):** pending — ticket setup and inventory artifacts ready to commit.

### What I did
- Added `scripts/01-inventory-glazed-repos.sh`.
- Ran the script from the infra-tooling repo.
- Captured output to `sources/01-glazed-repo-inventory.tsv`.
- Updated the rollout plan with first inventory observations.
- Checked off the inventory-script and inventory-run tasks.

### Why
- The rollout needs a concrete repository list before any Makefile or CI edits.
- The inventory needs to record dirty repos and duplicate checkouts so we do not accidentally edit backup or workspace copies.

### What worked
- The script produced TSV with repo path, module path, Glazed version, Makefile/lint target presence, workflow presence, lefthook presence, package layout, and git cleanliness.
- It identified repositories that already have `glazed-lint` wiring, including `glazed`, `clay`, `geppetto`, and `pinocchio`.

### What didn't work
- The broad scan surfaced backup, nested, dirty, and placeholder-module repositories. This is useful evidence but not a final target set.

### What I learned
- The canonical `/home/manuel/code/wesen/go-go-golems` tree contains many more Glazed-dependent repositories than the xgoja workspace.
- We need an explicit exclusion list for backup/experimental/dirty/placeholder repos.

### What was tricky to build
- The same module can appear in both the canonical tree and the xgoja workspace. The script records real paths and git status, but target selection still needs human policy.

### What warrants a second pair of eyes
- Whether to include every clean canonical Glazed-dependent repository in this rollout or split the work into smaller batches.

### What should be done in the future
- Add a second script that applies the Makefile/CI wiring only to an explicit target list.
- Exclude duplicate workspace checkouts unless intentionally chosen.

### Code review instructions
- Review `scripts/01-inventory-glazed-repos.sh`.
- Review `sources/01-glazed-repo-inventory.tsv` before deciding target repositories.

### Technical details

Inventory command:

```bash
ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/01-inventory-glazed-repos.sh \
  | tee ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/sources/01-glazed-repo-inventory.tsv
```
