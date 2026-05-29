---
Title: Diary
Ticket: INFRA-005
Status: active
Topics:
    - automation
    - release
    - github
    - docsctl
    - logcopter
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
      Note: Dependency scanner created in Step 1.
    - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/analysis/01-dashboard-improvement-candidates-from-infra-004-operations.md
      Note: Dashboard improvement candidate analysis created in Step 1.
    - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/design-doc/01-dependency-aware-rollout-dashboard-design-guide.md
      Note: Dashboard design guide created in Step 1.
    - Path: /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite
      Note: INFRA-004 SQLite database extended with normalized dependency tables in Step 1.
ExternalSources: []
Summary: Chronological diary for INFRA-005 dashboard improvement work.
LastUpdated: 2026-05-29T14:10:00-04:00
WhatFor: Track implementation and design steps for dependency-aware rollout dashboard improvements.
WhenToUse: Before resuming INFRA-005 implementation or reviewing why the dependency schema/design was introduced.
---

# Diary

## Goal

This diary records the implementation and design work for INFRA-005: improving the INFRA rollout dashboard so it can support dependency-ordered releases, dependency bumps, rollout evidence review, and health checks.

## Step 1: Add Dependency Tables and Write the Dashboard Improvement Design

The first step turned the post-INFRA-004 lessons into concrete data and design artifacts. The dashboard had already proven useful as a live status table, but the release-order question exposed a gap: the SQLite database needed normalized internal dependency facts, and the dashboard needed a design for rendering those facts as operator actions.

I added a dependency population script, ran it against the INFRA-004 SQLite database, exported evidence snapshots, and wrote two ticket documents. The first document lists dashboard improvements discovered during the rollout. The second document is a design guide for implementing dependency-aware release and bump views.

### Prompt Context

**User prompt (verbatim):** "add go-go-golems internal dependency information to the sqlite database, so that it is easier to do dependency ordered releasing and bumping of dependencies. Create a new ticket for improvements to the dashboard stuff, and add an analysis + design guide. Before writing the guide, but as you do your analysis , make a list of things that it would make sense to add to the dashboard now that we have been using it for a significant amount of time."

**Assistant interpretation:** Extend the existing INFRA-004 tracker database with queryable internal dependency information, then create a new dashboard-improvement ticket containing both a feature-candidate analysis and a design guide.

**Inferred user intent:** Make the dashboard capable of guiding the next release train, not just showing PR status, and preserve the design rationale in ticket docs.

**Commit (code):** N/A — changes are currently uncommitted in the `infra-tooling` worktree.

### What I did

- Created ticket `INFRA-005 — Improve INFRA rollout dashboard for dependency ordered releases`.
- Added `scripts/01-populate-internal-dependencies.py` to scan local Go-Go-Golems `go.mod` files and populate normalized dependency tables in the INFRA-004 SQLite DB.
- Ran the script against `sources/05-rollout-progress.sqlite`.
- Added new SQLite tables: `internal_modules`, `internal_dependency_edges`, `release_order_layers`, and `dependency_bump_candidates`.
- Added `scripts/02-populate-repo-issue-log.py` to derive grouped issue/fix history from tracker `events` and `validations`.
- Added new SQLite tables: `repo_issue_log` and `repo_issue_steps`.
- Exported CSV/text snapshots under `INFRA-005/sources/` for review.
- Wrote `analysis/01-dashboard-improvement-candidates-from-infra-004-operations.md` before writing the design guide, then updated it with the issue-log addition.
- Wrote `design-doc/01-dependency-aware-rollout-dashboard-design-guide.md`, then updated it to include issue-history tables and repository detail page behavior.
- Linked the new docs and scripts from the INFRA-005 index.
- Updated the INFRA-004 changelog to record that the existing tracker DB now contains normalized dependency and issue/fix tables.

### Why

- The INFRA-004 dashboard could show repository state, but it could not directly answer release-order or bump-order questions.
- The existing `repos.upstreams` JSON field was useful but hard to query and did not include exact `go.mod` versions.
- Release trains need both planning dependencies and actual module dependencies, because they answer different questions.
- Writing the candidate list before the design guide kept the design grounded in real operational pain from the rollout.

### What worked

- The dependency scan populated 76 internal modules and 212 internal `go.mod` dependency edges.
- The script produced two release-order trains: one based on actual direct `go.mod` edges, and one based on tracker upstreams.
- The script produced 120 dependency bump candidate rows.
- The issue-log scan populated 342 grouped issue rows and 808 issue timeline steps.
- Frontmatter validation passed for both new docs and the diary.
- The analysis and design guide now provide an implementation plan for dashboard routes, CLI commands, query functions, health panels, and per-repo issue history.

### What didn't work

- My first `docmgr validate frontmatter` command used paths prefixed with `ttmp/` while running from the repository root; `docmgr` resolved those relative to its docs root and looked for `ttmp/ttmp/...`. I reran validation with absolute paths, and both docs validated successfully.

### What I learned

- The existing tracker already had dependency intent in `repos.upstreams`, but release planning needs a normalized table plus exact required versions from `go.mod`.
- Tracker-upstream order and direct-`go.mod` order differ in useful ways; the dashboard should expose both rather than choosing one silently.
- Local tags can produce useful bump candidates, but the design should consider adding remote tag verification later so `latest_local_tag` does not become stale.

### What was tricky to build

- Parsing `go.mod` directly is simple enough for this use case, but it requires preserving direct versus indirect dependencies and replacement targets. The script records `indirect` and `replace_target` so the dashboard can decide whether a bump is mandatory or advisory.
- Some internal modules are not INFRA-004 tracker rows, such as important foundations outside the rollout scope. The schema therefore includes `internal_modules.in_tracker` and dependency state fields rather than assuming every internal module has a `repos` row.
- Release-order computation has two valid sources: actual module edges and tracker planning edges. I stored both as separate `train_name` values in `release_order_layers`.

### What warrants a second pair of eyes

- Review whether `dependency_bump_candidates.available_tag` should use local tags, remote GitHub tags, tracker tags, or a ranked combination.
- Review whether indirect internal dependencies should be shown by default or hidden behind a filter.
- Review the dashboard design’s recommendation to keep command execution outside the browser.
- Review whether the new dependency tables should be migrated into `02-rollout-tracker.py` schema directly or kept as a scanner-owned extension first.

### What should be done in the future

- Add dashboard routes for release layers and bump candidates.
- Add CLI commands to inspect dependency edges, release order, and bump candidates.
- Add repository detail pages that show dependencies, dependents, events, validations, and evidence links.
- Add logcopter and Glazed lint health panels.

### Code review instructions

- Start with `scripts/01-populate-internal-dependencies.py` and verify the schema/table names.
- Inspect the generated DB tables with:

```bash
sqlite3 sources/05-rollout-progress.sqlite '.tables'
sqlite3 -header -column sources/05-rollout-progress.sqlite \
  "select train_name, layer, count(*) from release_order_layers group by train_name, layer"
```

- Read the analysis document before the design guide; the design guide intentionally builds on the candidate list.
- Validate docs with:

```bash
docmgr validate frontmatter --doc /absolute/path/to/analysis.md --suggest-fixes
docmgr validate frontmatter --doc /absolute/path/to/design.md --suggest-fixes
```

### Technical details

- Dependency scan command:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
```

- Issue-log scan command:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/02-populate-repo-issue-log.py
```

- Scan result:
  - `modules=76`
  - `internal_edges=212`
  - `dependency_bump_candidates=120`
  - `repo_issue_log=342`
  - `repo_issue_steps=808`
- Release order layer counts:
  - `verified_unreleased_go_mod_direct`: layers 1/2/3 = 25/10/1
  - `verified_unreleased_tracker_upstreams`: layers 1/2/3 = 25/9/2
