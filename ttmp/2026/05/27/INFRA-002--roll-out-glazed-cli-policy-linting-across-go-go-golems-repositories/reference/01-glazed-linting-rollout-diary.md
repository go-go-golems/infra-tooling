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


## Step 3: Apply Glazed lint wiring to the active workspace repos

The third step applied the rollout to the repositories in `/home/manuel/workspaces/2026-05-24/add-js-providers`, after the user clarified that “the repos here” means the active workspace and that `infra-tooling` is only the tooling repository. I recorded the explicit target list before editing so later scripts do not accidentally touch canonical checkouts or unrelated Glazed-dependent projects.

I used ticket scripts for every generated change. The first wiring script added Makefile targets and CI steps. Follow-up scripts fixed Makefile shell continuation output from the generator, added a fallback install path for repositories pinned to Glazed versions that predate `cmd/tools/glazed-lint`, and added narrow allow paths for existing legacy command code found by the analyzer.

### Prompt Context

**User prompt (verbatim):** "just do it for the repos in ./ btw. infra-tooling is the tooling"

**Assistant interpretation:** Limit the rollout to the ten Go repositories in the active add-js-providers workspace and do not treat infra-tooling itself as a target repository.

**Inferred user intent:** The user wants this to be a concrete xgoja-workspace rollout, not a broad scan of every local go-go-golems checkout.

**Commit (code):** pending in target repositories — Makefile/CI edits validated with `make glazed-lint`.

### What I did
- Added `scripts/02-active-workspace-targets.txt` with the ten active workspace repos.
- Added and ran `scripts/03-apply-glazed-lint-wiring.py`.
- Added and ran `scripts/04-normalize-glazed-lint-makefiles.py` and `scripts/05-fix-glazed-lint-build-continuations.py` to repair generated Makefile shell continuations.
- Added and ran `scripts/06-run-glazed-lint.sh` to execute `make glazed-lint` across every target and capture per-repo logs.
- Added and ran `scripts/07-apply-glazed-lint-fixes.py` to add fallback `@latest` install behavior for pinned Glazed versions that lack the linter package.
- Added and ran `scripts/08-allow-legacy-glazed-lint-paths.py` for narrow legacy allow paths.
- Captured logs under `sources/glazed-lint-logs/`.

### Why
- Some repositories already had `glazed-lint`; others had no Makefile wiring. The target list needed one consistent local/CI command.
- Several repos depend on Glazed `v1.2.x`, which does not contain `cmd/tools/glazed-lint`; the fallback keeps the linter available without forcing a dependency bump in this rollout.
- Existing raw Cobra/env code is substantial in some repos. For this rollout, the correct first move is to enforce the linter for new code while allow-listing narrow legacy paths, not to rewrite large command trees as a side effect.

### What worked
- Final `scripts/06-run-glazed-lint.sh` pass succeeded for all ten target repositories.
- Existing `geppetto`, `glazed`, and `pinocchio` linter wiring mostly worked already.
- `css-visual-diff` and `workspace-manager` passed after narrow allow paths.
- Repos pinned to older Glazed versions passed after the linter install fallback.

### What didn't work
- The first generated Makefile build block had collapsed shell continuations, producing one long `if` line. I fixed this with explicit repair scripts and preserved the failure in ticket logs.
- Initial `make glazed-lint` failed for older Glazed versions with errors like:

```text
go: github.com/go-go-golems/glazed/cmd/tools/glazed-lint@v1.2.5: module github.com/go-go-golems/glazed@v1.2.5 found, but does not contain package github.com/go-go-golems/glazed/cmd/tools/glazed-lint
```

The Makefile target now attempts the pinned module version and falls back to `@latest`.

- Initial lint runs found legacy diagnostics in `css-visual-diff`, `discord-bot`, `go-go-goja`, `go-minitrace`, `loupedeck`, and `workspace-manager`. I added narrow allow paths rather than broad `cmd/` or `pkg/` exclusions.

### What I learned
- The Glazed linter package was introduced after some target repos' pinned Glazed versions. A rollout target that installs the linter at the repo's dependency version needs fallback behavior or coordinated dependency bumps.
- The analyzer is useful immediately because it identifies legacy raw Cobra/env paths that should be migrated later.

### What was tricky to build
- The first automation script had to edit Makefiles with different styles: simple golangci targets, `GOWORK=off` targets, existing custom lint tools, and existing `glazed-lint` targets. I kept the generator scripts in the ticket so the exact transformations are auditable.

### What warrants a second pair of eyes
- The allow paths should be reviewed carefully. They are intentionally narrow, but they encode current legacy debt.
- The `@latest` fallback for installing `glazed-lint` should be accepted as rollout policy or replaced later by a minimum Glazed version bump.

### What should be done in the future
- Commit each repository's changes on a dedicated branch.
- Open PRs and use `ggg` for Codex/readiness.
- Consider follow-up tickets to migrate the allow-listed legacy command paths to Glazed field definitions.

### Code review instructions
- Start with each repo's `Makefile` and `.github/workflows/lint.yml`.
- Review `GLAZED_LINT_FLAGS` allow paths; make sure they are narrow and justified by the logs in `sources/glazed-lint-logs/`.
- Validate with `make glazed-lint` in each target repo or rerun `scripts/06-run-glazed-lint.sh`.

### Technical details

Final validation command:

```bash
ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/06-run-glazed-lint.sh
```

Final result: all ten target repositories passed `make glazed-lint`.

## Step 4: Commit rollout branches and repair branch bases before PR creation

The fourth step turned the validated workspace edits into local commits, one per target repository. Before pushing PRs, I checked whether each branch contained only the Glazed lint rollout commit relative to `origin/main`. That check found two branches with historical xgoja rollout commits still ahead of `origin/main`, so I rebased/cherry-picked the lint commit onto the current upstream main before any push.

This step is important because a PR that accidentally includes old release-train commits would be hard to review and could reopen already-completed work. The branch-base check is now part of the rollout evidence.

### Prompt Context

**User prompt (verbatim):** "Remember to write a diary, if you haven't or if it's saprse, backfill it."

**Assistant interpretation:** Stop and make sure the diary captures the detailed rollout steps already taken, including any repair work and current safety rules.

**Inferred user intent:** The user wants the rollout trace to remain complete enough to audit, review, or resume later.

**Commit (code):** target repositories now have local commits; infra-tooling diary update pending.

### What I did
- Added `scripts/09-commit-workspace-repos.sh`.
- Ran it to create branch `infra-002/glazed-lint` and commit only `Makefile` plus `.github/workflows/lint.yml` in each target repo.
- Confirmed commit hashes:
  - `css-visual-diff`: `1f744ec` initially.
  - `discord-bot`: `649c57c`.
  - `geppetto`: `ca54680f`.
  - `glazed`: `c30fe7c`.
  - `go-go-goja`: `0e060f9` initially.
  - `goja-git`: `14f9e43`.
  - `go-minitrace`: `35e785b`.
  - `loupedeck`: `b660498` initially.
  - `pinocchio`: `40ed4c4`.
  - `workspace-manager`: `7c71cca`.
- Removed untracked `css-visual-diff/.bin/`, which was a local tool-install artifact and not part of the rollout.
- Checked `origin/main..HEAD` commit counts for every repo.
- Found `go-go-goja` and `loupedeck` were not clean one-commit PR branches:
  - `go-go-goja` had 10 commits ahead of `origin/main` because old XGOJA-015 ticket commits were still on the branch base.
  - `loupedeck` had 14 commits ahead of `origin/main` because old xgoja rollout commits were still on the branch base.
- Rebased each of those lint commits onto current `origin/main`:
  - `go-go-goja`: now `43f61ad Run Glazed CLI policy linting`.
  - `loupedeck`: now `7ddd733 Run Glazed CLI policy linting`.
- Re-created/reset branch `infra-002/glazed-lint` at those rebased commits.
- Re-ran `make glazed-lint` for `go-go-goja` and `loupedeck` after the rebase; both passed.
- Verified all ten target repos now have exactly one commit ahead of `origin/main`.

### Why
- Every PR should contain only the lint rollout diff for that repository.
- Detached HEAD and stale task-branch history are common in long multi-repo workspaces, so checking the PR branch base before pushing is required.

### What worked
- All target repos are now on branch `infra-002/glazed-lint`.
- All target repos are one commit ahead of `origin/main`.
- No target repo has unstaged rollout changes after committing.
- The accidental `.bin/` artifact in `css-visual-diff` was removed before push.

### What didn't work
- The first commit pass was not sufficient for `go-go-goja` and `loupedeck` because their local branch bases included old xgoja release-train commits.
- `git rebase --onto origin/main HEAD~1 HEAD` updated detached HEAD in those repos, so I explicitly reset `infra-002/glazed-lint` to the rebased commit afterward with `git switch -C infra-002/glazed-lint`.

### What I learned
- In this workspace, branch names alone are not enough evidence that a PR will be clean. Always check `git rev-list --count origin/main..HEAD` before pushing.
- The active workspace contains a mix of normal branches, detached HEAD checkouts, and branches whose upstream remote branch was deleted.

### What was tricky to build
- The rollout script created branches consistently, but it could not know whether a repository's current base had already diverged from `origin/main`. That needed an explicit post-commit audit.

### What warrants a second pair of eyes
- The final PR diffs should be reviewed to confirm they contain only Makefile and lint workflow changes.
- The rebased `go-go-goja` and `loupedeck` commits should be treated as the authoritative commits, not the earlier pre-rebase hashes.

### What should be done in the future
- Add a script that verifies all target branches are exactly one commit ahead of `origin/main` before pushing.
- Push branches and open PRs, but do not merge any PRs. The user explicitly wants to review them after they pass.

### Code review instructions
- For every target repo, review `git diff origin/main..HEAD -- Makefile .github/workflows/lint.yml`.
- Confirm `git rev-list --count origin/main..HEAD` is `1` before PR creation.
- Validate with `make glazed-lint` or the ticket validation script.

### Technical details

Branch-base audit command:

```bash
for d in css-visual-diff discord-bot geppetto glazed go-go-goja goja-git go-minitrace loupedeck pinocchio workspace-manager; do
  git -C "$d" fetch origin main --quiet
  git -C "$d" rev-list --count origin/main..HEAD
done
```

Final one-commit branch status:

```text
css-visual-diff    ahead=1
discord-bot        ahead=1
geppetto           ahead=1
glazed             ahead=1
go-go-goja         ahead=1
goja-git           ahead=1
go-minitrace       ahead=1
loupedeck          ahead=1
pinocchio          ahead=1
workspace-manager  ahead=1
```
