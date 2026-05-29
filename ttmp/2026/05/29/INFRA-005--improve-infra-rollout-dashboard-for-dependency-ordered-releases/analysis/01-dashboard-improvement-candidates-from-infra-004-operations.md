---
Title: Dashboard improvement candidates from INFRA-004 operations
Ticket: INFRA-005
Status: active
Topics:
    - automation
    - release
    - github
    - docsctl
    - logcopter
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
      Note: |-
        Existing dashboard and tracker implementation that should be evolved.
        Existing dashboard implementation analyzed for improvement candidates
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite
      Note: |-
        Existing SQLite rollout database, now extended with internal dependency tables.
        SQLite DB extended with normalized internal dependency tables
    - Path: ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
      Note: |-
        Script that populates normalized internal module/dependency/release-layer tables.
        Dependency scanner/populator introduced by this ticket
    - Path: ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/02-populate-repo-issue-log.py
      Note: Derived issue/fix tables added after identifying repo detail page needs
ExternalSources: []
Summary: Candidate dashboard improvements derived from operating the INFRA-004 rollout through repairs, ggg readiness, merges, and post-merge verification.
LastUpdated: 2026-05-29T13:45:00-04:00
WhatFor: Use this as the prioritized feature inventory before implementing the dependency-aware dashboard design.
WhenToUse: Before changing the dashboard UI, schema, or workflow automation for INFRA release trains.
---



# Dashboard Improvement Candidates from INFRA-004 Operations

The INFRA-004 dashboard started as a small live table over a SQLite database. That was the right first version. It made the rollout visible, it refreshed without a build step, and it gave the operator a place to record state transitions. After using it through a full cross-repository rollout, the missing pieces are clearer. The dashboard should not only show whether a repository is `pr_open`, `merged`, or `released`; it should help the operator decide what to do next.

The strongest lesson is that rollout work has two different orders. The PR order is driven by failure state: fix broken checks, address Codex feedback, wait for pending runs, merge when ready. The release order is driven by dependencies: tag upstreams first, bump those tags into downstreams, then release the downstreams. The current dashboard primarily represents the first order. The next version must represent both.

## 1. Current Dashboard Strengths

The existing dashboard has three properties worth preserving.

First, it reads SQLite directly on every request. This was useful during the rollout because command-line updates appeared without a frontend build or cache invalidation problem. The operator could run `02-rollout-tracker.py update-repo ...`, refresh the browser, and see the changed row.

Second, the schema is simple. The `repos` table contains a row per repository, and the `events` table records what happened. That simplicity made it easy to correct stale state, record merge SHAs, and leave a trail for the diary.

Third, the dashboard is operational rather than decorative. It displays the pieces the operator used: batch, repo, state, tracks, PR, merge SHA, tag, action status, and notes. The next version should keep this property. Every new panel should answer a real operator question.

## 2. Missing Views Discovered During INFRA-004

### 2.1 Dependency-order view

The most important missing view is a release train view ordered by first-party dependencies. During INFRA-004, we had to compute this manually from `repos.upstreams`, `go.mod`, and the report. That should be a first-class dashboard panel.

The view should answer:

- Which repositories can be released now because they have no unreleased internal dependencies?
- Which repositories are waiting on upstream tags?
- Which downstream repositories must be bumped after an upstream is tagged?
- Which dependency edges come from `go.mod`, and which come from tracker planning metadata?

This motivated the new normalized tables added to the SQLite database:

- `internal_modules`
- `internal_dependency_edges`
- `release_order_layers`
- `dependency_bump_candidates`

### 2.2 Bump-candidate view

Release work is not only tagging. It is also bumping downstream modules to the tags that were just created. The dashboard should show concrete bump candidates with current and available versions.

A useful row would read like this:

| Consumer | Dependency | Current | Available | Priority | Reason |
|---|---|---:|---:|---|---|
| `form-generator` | `sqleton` | `v0.3.4` | `v0.4.5` | high | `sqleton` has a newer available tag. |

This is more actionable than a graph alone. It tells the operator which `go get` command to run next.

### 2.3 Tracker-upstreams versus actual go.mod edges

The tracker’s `upstreams` field and `go.mod` direct dependencies do not always match. That is not always a bug. The tracker captures rollout intent, while `go.mod` captures current module reality. Both are useful.

The dashboard should show a diff:

- Edges present in both tracker and `go.mod` are strong dependencies.
- Edges present only in tracker may represent planned, transitive, or release-intent dependencies.
- Edges present only in `go.mod` may represent dependencies the batch planner missed.

This view would have helped during the `refactorio` and `vault-envrc-generator` work, where exact package/module structure mattered.

### 2.4 Next-action queue

`ggg batch ready` already classifies PRs into states such as `ready`, `failed_checks`, `codex_feedback`, `waiting_codex`, and `waiting_checks`. The dashboard should ingest or link to those snapshots and produce a queue ordered by action type.

A useful queue would group rows as:

1. **Fix now:** failed checks with known failed check kinds.
2. **Read Codex:** current-head Codex feedback exists.
3. **Wait:** pending checks or waiting for Codex.
4. **Merge candidates:** `ggg` says ready and per-PR evidence exists.
5. **Verify main:** merged but main workflows not yet verified.
6. **Release candidates:** main verified but not released.
7. **Bump candidates:** upstream tag exists but downstream still requires the old version.

The current dashboard shows state. It should also show queue position.

### 2.5 Main workflow verification panel

Main verification was a separate phase after merge. It should have a dedicated panel because PR checks are not enough. `smailnail` demonstrated why: a PR can merge and still expose main workflow problems.

The panel should show, per repository:

- Latest main SHA.
- `golang-pipeline` status.
- `golangci-lint` status.
- dependency scanning status.
- CodeQL status.
- Known unrelated failures such as secret scanning or image publishing.
- Last verification event from the `events` table.

This would separate rollout health from unrelated repository health.

### 2.6 Codex feedback panel

Codex feedback became a real readiness gate. The dashboard should show current-head Codex feedback distinctly from historical feedback.

The panel should answer:

- Does the PR have current-head actionable feedback?
- What files are mentioned?
- Was Codex retriggered after the last branch push?
- Is the current state `waiting_codex`, `codex_feedback`, or `ready`?

This would prevent the operator from merging a green PR that still has current-head review concerns.

### 2.7 Failure classifier and issue taxonomy

The rollout repeatedly encountered the same classes of failure:

- malformed `push.yml` test steps,
- missing `logcopter.go` files,
- `glazed-lint` package/version errors,
- `govulncheck` findings,
- `gosec` findings,
- unavailable Dependency Review,
- Nancy/OSS Index 401 behavior,
- CI/local tool version drift,
- generated artifacts accidentally committed.

The dashboard should let events or validations carry a `failure_class` field. A view grouped by failure class would show whether a problem is systemic or isolated.

### 2.8 Logcopter generation health

Logcopter created a new invariant: generation and check package lists must match, and generated files must be committed even if `.gitignore` hides them. The dashboard should have a logcopter-specific health view.

It should show:

- Whether `go.mod` requires `github.com/go-go-golems/logcopter`.
- Which version is required.
- Whether a `tool github.com/go-go-golems/logcopter/cmd/logcopter-gen` directive exists.
- Whether `logcopter_generate.go` exists.
- Whether `make logcopter-check` exists.
- Whether generate/check package lists appear to match.
- Whether `cmds/...` or other top-level command directories are omitted.

This view would have caught the `vault-envrc-generator` drift earlier.

### 2.9 Glazed lint policy health

The dashboard should also have a Glazed lint view:

- `GLAZED_VERSION` value.
- Whether the analyzer package is installable at that version.
- Whether `GLAZED_LINT_ALLOW_PATHS` exists.
- Whether allow paths are commented and scoped.
- Which package patterns are covered by `go vet -vettool`.
- Whether known command directories are skipped.

The dashboard should not decide whether an allow path is acceptable, but it should make broad exceptions visible.

### 2.10 Release evidence and missing-tag alerts

Many rows are `main_actions_verified` but not `released`. The dashboard should show release debt as a first-class status:

- verified but no tag,
- tag recorded but no release URL,
- release URL recorded but state not `released`,
- downstreams not bumped to the latest upstream tag.

This would have made the post-rollout question—“what should we release and bump?”—answerable from the dashboard itself.

### 2.11 Blocked/skipped backlog view

Blocked and skipped rows should not disappear into the main table. They need their own backlog view with explicit reasons and required decisions.

For example:

- `voyage` is archived/read-only and has pre-existing build failures.
- `bubble-table` has a non-Go-Go-Golems module path and no normal rollout targets.
- `barbar` generated no useful logcopter files mechanically.
- `geppetto` is skipped but important to B5 and downstream dependency convergence.

A blocked/skipped view would help decide whether to create follow-up tickets.

### 2.12 Evidence browser

The ticket accumulated many timestamped files under `sources/`: `ggg` readiness snapshots, Codex comment dumps, failed logs, and trigger outputs. The dashboard should link to evidence rather than forcing the operator to browse the filesystem.

A simple evidence browser could show:

- latest `ggg batch ready` snapshot,
- latest per-PR ready evidence,
- latest Codex comments,
- failed log files,
- generated manifests,
- report snapshots.

### 2.13 Command generator

The dashboard should provide copyable commands for the next action. These commands should be generated from database state.

Examples:

```bash
ggg pr ready https://github.com/go-go-golems/oak/pull/47 --findings --output json
```

```bash
GOWORK=off go get github.com/go-go-golems/parka@v0.6.2
GOWORK=off go mod tidy
make logcopter-check
make glazed-lint
GOWORK=off go test ./...
```

The command generator does not need to execute commands. Its value is consistency.

### 2.14 Dependency graph visualization

A small graph view would help, but it should not be the first implementation. Tables are more actionable for release work. The graph should come after normalized dependency tables and release layers exist.

The useful graph is not “all repositories”. That graph is too dense. The useful graph is scoped:

- selected repository and its upstreams/downstreams,
- current release train candidates,
- B5 planned graph,
- logcopter downstreams,
- skipped/blocked dependencies that block convergence.

### 2.15 State transition timeline

The `events` table became valuable during handoff. The dashboard should show repository timelines, not only the last note. A repo detail page should display:

- all state changes,
- validation records,
- Codex triggers,
- merge event,
- main verification event,
- release event.

This turns the dashboard into an operational audit log.

## 3. Prioritized Feature List

The following order is recommended because each feature enables the next one.

### P0: Make dependency-ordered releases possible

1. Add normalized internal dependency tables. **Done in this ticket via `01-populate-internal-dependencies.py`.**
2. Add release-layer queries to the dashboard.
3. Add bump-candidate queries to the dashboard.
4. Show tracker-upstream layers and actual `go.mod` layers side by side.

### P1: Make operator next actions visible

1. Add a next-action queue.
2. Add main workflow verification status.
3. Add current-head Codex status.
4. Add release debt status.

### P2: Make systemic failures visible

1. Add failure classification to events/validations.
2. Add logcopter generation health checks.
3. Add Glazed lint policy health checks.
4. Add blocked/skipped backlog views.

### P3: Improve navigation and evidence

1. Add repository detail pages.
2. Add evidence browser.
3. Add copyable command generator.
4. Add scoped dependency graph visualization.

## 4. Immediate Schema Improvement Already Applied

This ticket added normalized dependency information to the INFRA-004 SQLite database. The new tables are:

| Table | Purpose |
|---|---|
| `internal_modules` | Local Go-Go-Golems modules, whether they are tracked, tracker state, tags, release URL, and local git metadata. |
| `internal_dependency_edges` | Internal `go.mod` require edges with current version, direct/indirect marker, replace target, and dependency tracker state. |
| `release_order_layers` | Computed release layers for verified-but-unreleased repositories. Two trains are stored: actual `go.mod` direct edges and tracker upstream intent. |
| `dependency_bump_candidates` | Rows where a consumer requires an older internal module version than the available tag known to the scanner. |

The population command is:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
```

The scan populated:

- 76 internal modules,
- 212 internal `go.mod` dependency edges,
- release layers for `verified_unreleased_go_mod_direct`,
- release layers for `verified_unreleased_tracker_upstreams`,
- bump candidates for internal dependencies with newer known tags.

## 5. Queries That Should Become Dashboard Panels

### Release layers

```sql
SELECT layer, repo, depends_on_json, dependents_json
FROM release_order_layers
WHERE train_name = 'verified_unreleased_tracker_upstreams'
ORDER BY layer, repo;
```

### Bump candidates

```sql
SELECT repo, dependency_repo, current_required_version, available_tag, priority, reason
FROM dependency_bump_candidates
ORDER BY CASE priority WHEN 'high' THEN 0 ELSE 1 END, dependency_repo, repo;
```

### Internal dependency edges for one repo

```sql
SELECT dependency_repo, required_version, indirect, replace_target, dependency_state, dependency_tracker_tag, dependency_latest_local_tag
FROM internal_dependency_edges
WHERE repo = ?
ORDER BY indirect, dependency_repo;
```

### Downstream impact of releasing one repo

```sql
SELECT repo, required_version, indirect
FROM internal_dependency_edges
WHERE dependency_repo = ?
ORDER BY indirect, repo;
```

### Tracker intent versus go.mod reality

```sql
-- Tracker upstreams still live in repos.upstreams as JSON.
-- The dashboard can compare that JSON to direct internal_dependency_edges.
SELECT repo, dependency_repo
FROM internal_dependency_edges
WHERE indirect = 0
ORDER BY repo, dependency_repo;
```

## 6. Structured Issue/Fix History Added After Further Analysis

After the initial dependency tables, the next high-value data addition was a structured issue/fix log. The existing `events` and `validations` tables are chronological, but a repository detail page needs a grouped view: what issue was found, which category it belongs to, what fixed it, and which events or validations prove that outcome.

This ticket now adds two derived tables to the INFRA-004 database:

| Table | Purpose |
|---|---|
| `repo_issue_log` | One row per repository issue category, with title, severity, status, detected evidence, root cause, fix summary, commits, validation summary, and source references. |
| `repo_issue_steps` | Timeline steps for each issue, derived from `events` and `validations`, preserving source table/id, command, status, message, URL, and commit SHA when detected. |

The population command is:

```bash
cd /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/02-populate-repo-issue-log.py
```

The initial scan populated:

- 342 issue-log rows,
- 808 issue timeline steps.

This makes a repository detail page much more useful. Instead of showing only a latest note, it can show sections such as:

- Workflow YAML issue: detected from failed/malformed `push.yml`, fixed by repair commit, verified by main workflow success.
- Glazed lint issue: analyzer version, allow path, or package coverage problem, fixed by Makefile changes and `make glazed-lint` validation.
- Logcopter generation issue: missing generated files or mismatched package lists, fixed by generated files and `make logcopter-check`.
- Security issue: `govulncheck`, `gosec`, Dependency Review, or Nancy issue, with the fix and validation trail.
- Codex feedback: current-head review feedback, trigger events, and merge readiness evidence.

The issue log is derived, not hand-authored. That makes it cheap to refresh while preserving the original event stream as the audit source.

## 7. Decision Point

The next dashboard implementation should start with dependency release panels and repository detail pages, not visual polish. The operator now has enough evidence to need computed release order, bump candidates, upstream/downstream impact, and per-repo issue history. Those features directly reduce release risk. A graph can follow once the tables are doing the correct work.
