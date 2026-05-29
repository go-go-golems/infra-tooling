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

## Step 2: Implement the First Dependency-Aware Dashboard Pages

This step turned the dependency and issue-history tables into operator-facing views. The dashboard can now answer the first useful questions directly: what should be released in dependency order, what dependency bumps are available, and what happened to an individual repository during the rollout.

The implementation deliberately stayed inside the existing small Python tracker. It adds plain HTML routes and CLI queries without introducing a frontend build, a web framework, or background jobs. That keeps the dashboard close to the SQLite database and preserves the fast operator workflow that worked during INFRA-004.

### Prompt Context

**User prompt (verbatim):** "Add tasks to the ticket, then implement them task by task, committing at appropriate intervals, keeping a diary"

**Assistant interpretation:** Add concrete INFRA-005 implementation tasks, complete the highest-value dashboard tasks in order, commit code and docs at sensible boundaries, and update the diary as work progresses.

**Inferred user intent:** Move from design artifacts to working dashboard functionality while preserving reviewable commits and continuation-friendly ticket notes.

**Commit (code):** `0973b71b5807b87140be0719795e104fc4a01b00` — "Add dependency dashboard views and CLI queries"

### What I did

- Added implementation-specific tasks to INFRA-005 and checked off the tasks completed by this step.
- Extended `scripts/02-rollout-tracker.py` with dependency CLI commands:
  - `deps-scan`
  - `deps-modules`
  - `deps-edges`
  - `deps-release-order`
  - `deps-bumps`
- Added issue-history CLI commands:
  - `issue-refresh`
  - `issue-list`
- Added dashboard routes:
  - `/release-train`
  - `/bumps`
  - `/repo?repo=<name>`
- Linked repository names from the overview table to the new repository detail route.
- Rendered release layers from `release_order_layers`.
- Rendered grouped bump candidates from `dependency_bump_candidates`.
- Rendered repository detail pages from `repos`, `internal_dependency_edges`, `dependency_bump_candidates`, `repo_issue_log`, `repo_issue_steps`, `validations`, and `events`.
- Validated the new script with `python -m py_compile` and smoke-tested CLI/render functions against the populated INFRA-004 DB.

### Why

- The most valuable dashboard improvement was to turn the database into an operator decision tool: release order, bump queue, and per-repository history.
- The repo detail page needed grouped issue/fix history rather than only raw event rows.
- CLI commands make the same data available outside the browser and are easier to test in ticket workflows.

### What worked

- `deps-release-order` prints the dependency-ordered release layers from the SQLite DB.
- `deps-bumps --dependency parka` correctly shows downstream consumers such as `escuse-me` and `sqleton`.
- `issue-list --repo smailnail` shows grouped issue/fix categories including workflow YAML, Glazed lint, govulncheck, logcopter, and main verification.
- Direct calls to `html_release_train`, `html_bumps`, and `html_repo_detail(..., "smailnail")` all rendered non-empty HTML without exceptions.

### What didn't work

- N/A for the implementation itself. The main caveat is that `repo_issue_log` is heuristic because it is derived from existing event/validation text.

### What I learned

- The existing dashboard can be extended substantially without a framework if query/render code is kept simple and route-specific.
- The issue-history tables make repository pages far more useful than raw events alone, but classification quality depends on how consistently future events are written.
- Keeping dependency and issue refresh as explicit commands is safer than refreshing them on every browser request.

### What was tricky to build

- The tricky part was keeping the old overview route intact while adding richer pages. I avoided changing the existing data model for `repos` and instead made the new routes optional: if derived tables are missing, the page tells the operator what scanner to run.
- The repository page pulls from many tables. I kept it read-only and bounded raw events/validations to recent rows so the page remains useful even for noisy repositories.

### What warrants a second pair of eyes

- Review the HTML rendering for escaping and route behavior. The implementation uses `html.escape`, but this should be checked because dashboard rows include free-form event text.
- Review whether the new CLI command names should be kept as top-level commands or grouped under a future nested `deps`/`issues` subcommand.
- Review the generated bump command snippets; they intentionally do not try to infer every repository-specific validation command.

### What should be done in the future

- Add logcopter and Glazed health panels.
- Add blocked/skipped backlog and failure-classification views.
- Add tests around query/render helpers using a small fixture SQLite DB.

### Code review instructions

- Start with `scripts/02-rollout-tracker.py` around the new `html_release_train`, `html_bumps`, `html_repo_detail`, `deps_*_cmd`, and `issue_list_cmd` functions.
- Validate with:

```bash
python -m py_compile ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
python ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py deps-release-order
python ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py deps-bumps --dependency parka
python ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py issue-list --repo smailnail
```

### Technical details

- New routes are served by the existing `dashboard` command:
  - `http://127.0.0.1:8765/release-train`
  - `http://127.0.0.1:8765/bumps`
  - `http://127.0.0.1:8765/repo?repo=smailnail`
- The dependency refresh command invokes the INFRA-005 dependency scanner by default.
- The issue refresh command invokes the INFRA-005 issue-log scanner by default.

## Step 3: Add Logcopter and Glazed Health Panels

This step added a lightweight health scanner and dashboard views for the two rollout-specific policy areas that repeatedly caused issues: logcopter generation/checking and Glazed CLI linting. The purpose is not to replace CI. The purpose is to give the operator a fast, per-repository summary of likely configuration drift before running release or bump work.

The scanner reads local repository files and stores derived checks in SQLite. The dashboard can now render a global health page and include health checks on each repository detail page.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementing the dashboard tasks one by one, with commits and diary updates at sensible boundaries.

**Inferred user intent:** Keep turning the INFRA-005 design into working, reviewable dashboard functionality while preserving a clear implementation trail.

**Commit (code):** pending at diary write time — planned commit message: "Add rollout health dashboard panels"

### What I did

- Added `scripts/03-populate-repo-health.py`.
- Added `repo_health_checks` to the INFRA-004 SQLite DB.
- Populated 623 health-check rows.
- Added `health-refresh` and `health-list` CLI commands to `scripts/02-rollout-tracker.py`.
- Added dashboard route `/health` with filters for status and category.
- Added health-check rows to `/repo?repo=<name>` pages.
- Exported health snapshots under INFRA-005 `sources/`.

### Why

- Logcopter and Glazed lint issues were two of the most common rollout-specific failure classes.
- Repository detail pages should show not only what failed historically, but also whether current local configuration has obvious drift.
- A lightweight scanner provides fast signals without running expensive CI-equivalent commands for every repository.

### What worked

- `health-list --repo vault-envrc-generator --status warn` showed the expected warning about missing top-level `logcopter_generate.go`.
- The `/health` renderer and filtered health renderer produced non-empty HTML.
- The repository detail renderer for `vault-envrc-generator` includes health-check rows.
- The scanner summarized health state by category/status:
  - `glazed_lint`: 17 fail, 137 pass, 39 skip, 52 warn
  - `logcopter`: 22 fail, 287 pass, 69 warn

### What didn't work

- The first scanner version failed to detect Makefile targets with prerequisites, such as `glazed-lint: glazed-lint-build`. I fixed `extract_target_line` so it accepts both exact `target:` and `target: prerequisites` forms.

### What I learned

- Static health checks are useful but should be presented as inspection signals. A warning can mean “this repo uses a Makefile-only generation pattern” rather than “CI is broken”.
- Health checks need to be category-specific and file-linked so reviewers can quickly inspect the exact Makefile or `go.mod` source.

### What was tricky to build

- The scanner needed to avoid claiming too much. It does not run `make logcopter-check` or `make glazed-lint`; it records whether expected targets, versions, tool directives, package patterns, and allow-path comments exist. That keeps the scan cheap and safe.
- Makefile parsing is intentionally shallow. It is good enough for dashboard signals, but it should not be treated as a full Make parser.

### What warrants a second pair of eyes

- Review the fail/warn thresholds. Some skipped or blocked repositories may appear as health failures because they still have rollout flags in the tracker but intentionally lack generated targets.
- Review whether missing `logcopter_generate.go` should be a warning when Makefile generation targets exist.
- Review whether `GLAZED_LINT_ALLOW_PATHS` comment detection should be stricter or looser.

### What should be done in the future

- Add a blocked/skipped backlog page that explains intentional health failures separately from actionable failures.
- Add small fixture tests for `extract_target_line`, logcopter package comparison, and Glazed allow-path detection.
- Add filters on repo detail pages to hide `pass` checks when a repo has many rows.

### Code review instructions

- Start with `scripts/03-populate-repo-health.py` for scanner behavior.
- Then review `scripts/02-rollout-tracker.py` changes around `html_health`, `health_list_cmd`, `health-refresh`, and repo detail health rendering.
- Validate with:

```bash
python -m py_compile ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/03-populate-repo-health.py
python -m py_compile ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
python ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py health-list --repo vault-envrc-generator --status warn
```

### Technical details

- Health scan command:

```bash
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/03-populate-repo-health.py
```

- Dashboard routes:
  - `/health`
  - `/health?status=warn`
  - `/health?category=logcopter`
  - `/repo?repo=vault-envrc-generator`

## Step 4: Add Issue Classification and Blocked Backlog Pages

This step finished the first pass of the dashboard views promised by the INFRA-005 task list. The dashboard can now show issue classifications across repositories and a dedicated backlog for rows that are blocked, skipped, planned, or stuck in local validation.

These pages matter because they keep unresolved work visible without mixing it into the release train. Release pages should focus on what can move forward; backlog pages should explain what needs a separate decision.

### Prompt Context

**User prompt (verbatim):** (same as Step 2)

**Assistant interpretation:** Continue implementing the next dashboard task, committing and updating the diary after the step.

**Inferred user intent:** Complete the first useful set of dashboard pages before moving on to later refinements or tests.

**Commit (code):** pending at diary write time — planned commit message: "Add issue and backlog dashboard pages"

### What I did

- Added `/issues`, a dashboard page backed by `repo_issue_log`.
- Added filters for issue status and category.
- Added `/blocked`, a dashboard page for `blocked`, `skipped`, `planned`, and `local_validation` rows.
- Included issue counts and health-finding counts in the blocked/skipped backlog.
- Linked `/issues` and `/blocked` from dashboard navigation and the overview card.
- Smoke-tested `html_issues` and `html_blocked` render functions against the populated DB.

### Why

- The release train page intentionally hides blocked/skipped work so release candidates remain clear. That unresolved work still needs a dedicated page.
- Issue classifications provide a cross-repository view of recurring rollout problems, such as workflow YAML, Glazed lint, logcopter generation, govulncheck, and gosec.

### What worked

- `/issues` renders grouped issue/fix rows and supports query filters such as `?status=observed` or `?category=workflow_yaml`.
- `/blocked` renders the backlog with notes and suggested decision text.
- Both render functions produced non-empty HTML in local smoke tests.

### What didn't work

- N/A. This was a straightforward read-only dashboard extension.

### What I learned

- The dashboard needs separate pages for forward progress and unresolved backlog. Combining them would make the release train page harder to use.
- The derived issue and health tables compose well: blocked rows can show both unresolved issue-group counts and health-finding counts.

### What was tricky to build

- The main decision was what to include in the backlog. I included `blocked`, `skipped`, `planned`, and `local_validation`, because all four states represent work that should not appear as normal release candidates.

### What warrants a second pair of eyes

- Review the suggested decision text on `/blocked`; it is heuristic and may need ticket-specific wording.
- Review whether `/issues` should hide `fixed` by default or continue showing all categories by default.

### What should be done in the future

- Add fixture tests for the new render helpers.
- Consider adding counts from GitHub Actions or ggg snapshots once those are imported into SQLite.

### Code review instructions

- Review `scripts/02-rollout-tracker.py` functions `html_issues` and `html_blocked`.
- Validate with:

```bash
python -m py_compile ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
```

- Open:
  - `http://127.0.0.1:8765/issues`
  - `http://127.0.0.1:8765/issues?category=workflow_yaml`
  - `http://127.0.0.1:8765/blocked`

### Technical details

- `/issues` reads `repo_issue_log`.
- `/blocked` reads `repos`, plus correlated counts from `repo_issue_log` and `repo_health_checks`.
