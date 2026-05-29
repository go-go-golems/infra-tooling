---
Title: Dependency-aware rollout dashboard design guide
Ticket: INFRA-005
Status: active
Topics:
    - automation
    - release
    - github
    - docsctl
    - logcopter
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py
      Note: |-
        Existing Python tracker/dashboard to extend.
        Target implementation for dashboard routes and CLI extensions
    - Path: ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/analysis/01-dashboard-improvement-candidates-from-infra-004-operations.md
      Note: |-
        Feature inventory derived from operating INFRA-004.
        Feature inventory that informed the design
    - Path: ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
      Note: |-
        Dependency scanner and schema extension implemented before this design guide.
        New dependency schema and population logic
    - Path: ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/02-populate-repo-issue-log.py
      Note: Issue/fix history populator for repository detail pages
ExternalSources: []
Summary: Design guide for evolving the INFRA rollout dashboard from a status table into a dependency-aware release and bump cockpit.
LastUpdated: 2026-05-29T14:00:00-04:00
WhatFor: Use this guide to implement dashboard changes that support dependency ordered releases, bump planning, ggg readiness, and rollout evidence review.
WhenToUse: Before modifying the rollout tracker schema, CLI, or HTML dashboard.
---



# Dependency-Aware Rollout Dashboard Design Guide

The rollout dashboard has reached the point where a single status table is no longer enough. It solved the first problem: show what state each repository is in while a batch rollout is active. The next problem is harder: decide which repository to release next, which downstream repositories must be bumped afterward, and which visible failures are relevant to the current phase of work.

This guide describes the next version of the dashboard. The design keeps the original strengths—SQLite, a small Python server, no build step, command-line-friendly updates—and adds a normalized dependency model, release train views, bump candidate views, and evidence-oriented operator pages.

## 1. Executive Summary

The next dashboard should become a release cockpit for Go-Go-Golems infrastructure rollouts. It should still display repository state, but its central unit of work should be the **next safe action**.

A safe action is different in each phase:

- During PR rollout, the next safe action may be “fix failed checks”, “address current-head Codex feedback”, “wait for pending checks”, or “merge this ready PR”.
- During main verification, the next safe action may be “inspect the failed main pipeline” or “mark main actions verified”.
- During release, the next safe action may be “tag this Layer 1 repository” or “bump this newly tagged upstream into its downstream consumers”.
- During cleanup, the next safe action may be “resolve blocked module ownership” or “open a follow-up ticket for skipped repositories”.

The dashboard already contains the base facts in `repos`, `events`, and `validations`. INFRA-005 adds a second layer of facts: normalized internal modules, internal dependency edges, release layers, and dependency bump candidates. The UI should make these facts actionable.

## 2. Problem Statement

The existing dashboard answers “what state is this repository in?” That was enough when the operator was creating branches and watching PRs. It became insufficient after the rollout reached the release phase.

The release phase asks questions the current dashboard cannot answer directly:

1. Which verified repositories can be tagged now without waiting for another unreleased first-party dependency?
2. Which repositories depend on the tag I am about to create?
3. Which `go.mod` files currently require an older internal module version?
4. Does the dependency order come from actual `go.mod` edges, tracker planning metadata, or both?
5. Which failures are rollout blockers, and which are unrelated repository health issues?
6. What evidence supports a row’s current state?

During INFRA-004, these answers came from manual SQLite queries, diary reconstruction, `ggg` JSON snapshots, local `go.mod` inspection, GitHub Actions, and generated reports. The next dashboard should turn that manual reconstruction into normal UI behavior.

## 3. Conceptual Model

The dashboard should model four related graphs.

### 3.1 Repository state graph

The repository state graph is the original tracker model. Each repository moves through states such as `planned`, `pr_open`, `ready`, `merged`, `main_actions_verified`, and `released`.

```text
planned
  -> branch_created
  -> local_validation
  -> pr_open
  -> ready
  -> merged
  -> main_actions_verified
  -> released
```

The real workflow has side states such as `blocked`, `skipped`, `codex_feedback`, and `codex_waiting`. These are not errors in the model. They are operational states that say what kind of intervention is needed.

### 3.2 Dependency graph

The dependency graph is a directed graph from consumer to dependency:

```text
form-generator -> sqleton
form-generator -> uhoh
refactorio     -> oak
oak            -> bobatea
```

For release ordering, the edge means “do not release the consumer as part of a converged train until the dependency has been released and the consumer has been bumped if necessary.”

### 3.3 Evidence graph

The evidence graph connects a repository state to supporting records:

```text
repo
  -> events
  -> validations
  -> ggg readiness JSON
  -> Codex comments JSON
  -> GitHub Actions run URLs
  -> merge SHA
  -> release URL
```

A dashboard row without evidence is just a claim. The next version should make evidence easy to inspect.

### 3.4 Action graph

The action graph is derived. It says what the operator should do next. It is computed from repository state, dependency edges, check state, and evidence freshness.

Examples:

- A repository in `main_actions_verified` with no release URL and no unreleased dependencies becomes a release candidate.
- A repository that depends on a just-released upstream and still requires an older version becomes a bump candidate.
- A PR with current-head Codex feedback becomes a review-fix candidate.
- A merged repository with no successful main workflow event becomes a main-verification candidate.

The action graph is what makes the dashboard useful during a long rollout.

## 4. Data Model

The existing schema should remain compatible. Do not replace `repos`; extend it.

### 4.1 Existing tables

| Table | Responsibility |
|---|---|
| `repos` | Repository identity, rollout tracks, state, PR URL, merge SHA, tag, release URL, action status, and notes. |
| `events` | Human-readable event stream for state changes, repairs, merges, Codex triggers, and verification. |
| `validations` | Local command outcomes. |

### 4.2 New dependency tables added by INFRA-005

The script `scripts/01-populate-internal-dependencies.py` adds four dependency-oriented tables to the INFRA-004 SQLite database.

#### `internal_modules`

This table normalizes local Go-Go-Golems modules.

| Column | Meaning |
|---|---|
| `module` | Go module path, primary key. |
| `repo` | Repository directory name. |
| `path` | Local checkout path. |
| `in_tracker` | Whether this module has a `repos` row. |
| `tracker_state` | Current tracker state if tracked. |
| `tracker_batch` | Batch ID if tracked. |
| `tracker_tag` | Tag recorded in the tracker. |
| `tracker_release_url` | Release URL recorded in the tracker. |
| `latest_local_tag` | Latest local git tag found in the checkout. |
| `head_sha` | Local checkout HEAD SHA at scan time. |
| `scanned_at` | Scan timestamp. |

This table matters because not every internal module is in the rollout tracker. `clay` and `glazed`, for example, appear as important upstreams even though they are not normal INFRA-004 rows.

#### `internal_dependency_edges`

This table records internal Go module dependencies discovered from `go.mod`.

| Column | Meaning |
|---|---|
| `repo` | Consumer repository. |
| `module` | Consumer module path. |
| `dependency_repo` | Internal dependency repository. |
| `dependency_module` | Internal dependency module path. |
| `required_version` | Version required by the consumer. |
| `indirect` | Whether the `go.mod` line is marked `// indirect`. |
| `replace_target` | Replacement target if the module is replaced. |
| `dependency_state` | Tracker state of the dependency if tracked. |
| `dependency_tracker_tag` | Dependency tag from tracker if present. |
| `dependency_latest_local_tag` | Latest local git tag for the dependency. |

This table is the basis for actual dependency-order release planning.

#### `release_order_layers`

This table stores computed release layers. There are currently two train names:

| Train | Meaning |
|---|---|
| `verified_unreleased_go_mod_direct` | Layers computed from direct internal `go.mod` edges. |
| `verified_unreleased_tracker_upstreams` | Layers computed from `repos.upstreams`, preserving rollout planning intent. |

The dashboard should show both. When they disagree, the disagreement is useful information.

#### `dependency_bump_candidates`

This table shows consumers whose required internal version differs from an available tag known to the scanner.

| Column | Meaning |
|---|---|
| `repo` | Consumer repository. |
| `dependency_repo` | Internal dependency repository. |
| `current_required_version` | Version currently required. |
| `available_tag` | Newer or different known tag. |
| `priority` | `high` when the dependency is released or main-verified. |
| `reason` | Human-readable explanation. |

This is the table that turns dependency data into bump work.

### 4.3 New issue-history tables added by INFRA-005

The script `scripts/02-populate-repo-issue-log.py` adds two issue-history tables. These tables are derived from `events` and `validations`, so they do not replace the original audit trail. They reorganize it for repository detail pages.

#### `repo_issue_log`

This table stores one grouped issue per repository and category.

| Column | Meaning |
|---|---|
| `repo` | Repository affected by the issue. |
| `issue_key` | Stable key, currently the issue category. |
| `category` | Machine-readable category such as `workflow_yaml`, `glazed_lint`, `logcopter_generation`, `govulncheck`, or `codex_feedback`. |
| `severity` | Initial severity used for dashboard sorting. |
| `title` | Human-readable issue title. |
| `status` | Derived status such as `fixed`, `observed`, `warning`, or `blocked`. |
| `detected_at` | Earliest detected/warning step timestamp, or first step timestamp. |
| `evidence_summary` | First useful detection message. |
| `root_cause` | Category-level root cause explanation. |
| `fix_summary` | Last useful fix or validation message. |
| `fixed_at` | Timestamp of the latest fix/validation step when present. |
| `fix_commits_json` | Commit SHAs found in source messages. |
| `validation_summary` | Recent validations associated with this issue. |
| `source_refs_json` | References back to `events:<id>` and `validations:<id>`. |

#### `repo_issue_steps`

This table stores the issue timeline.

| Column | Meaning |
|---|---|
| `issue_id` | Parent `repo_issue_log` row. |
| `step_time` | Time of the source event or validation. |
| `step_kind` | Derived kind: `detected`, `warning`, `fix_progress`, `fix_or_validation`, or `note`. |
| `source_table` / `source_id` | Original audit source. |
| `command` | Validation command when applicable. |
| `status` | Event kind or validation status. |
| `message` | Original message or validation summary. |
| `url` | Source URL when present. |
| `commit_sha` | First commit SHA found in the message, when present. |

These tables are what make a repository detail page useful. A page for `smailnail`, for example, can show separate cards for workflow YAML repair, Glazed lint repair, govulncheck repair, logcopter validation, and main verification.

## 5. Dashboard Pages

The dashboard can remain a single Python HTML renderer, but it should gain routes. A route can be simple: inspect `urlparse(self.path).path` and render a different HTML page. No frontend framework is required.

### 5.1 Overview page

The overview page should retain the current table but add summary cards:

- PR queue count.
- Main-verification queue count.
- Release candidate count.
- Bump candidate count.
- Blocked/skipped count.
- Latest dependency scan timestamp.

The overview should answer: “Where is the rollout right now?”

### 5.2 Release train page

Route: `/release-train?train=verified_unreleased_tracker_upstreams`

This page should show release layers.

| Layer | Repository | Depends on | Dependents | State | Tag | Release URL | Suggested action |
|---:|---|---|---|---|---|---|---|
| 1 | `bobatea` | — | `oak`, `go-go-agent` | `main_actions_verified` | — | — | Tag release. |
| 2 | `oak` | `bobatea` | `refactorio` | `main_actions_verified` | — | — | Wait for/bump `bobatea`, then tag. |

The page should include a toggle between:

- tracker-upstream order,
- direct `go.mod` order.

The operator should be able to see why `form-generator` is later than `sqleton`, and why `refactorio` is later than `oak`.

### 5.3 Bump candidates page

Route: `/bumps`

This page should show `dependency_bump_candidates` grouped by dependency. Grouping by dependency is important because the operator usually releases one upstream and then wants to bump every downstream.

Example grouping:

```text
sqleton -> form-generator
parka   -> escuse-me, sqleton
go-mcp  -> jesus, smailnail
```

Each row should include copyable commands:

```bash
GOWORK=off go get github.com/go-go-golems/sqleton@v0.4.5
GOWORK=off go mod tidy
make logcopter-check
make glazed-lint
GOWORK=off go test ./...
```

The command block should be conservative. It should not assume every repository has the same tags or test command. Where the dashboard cannot know the test command, it should say “check `.github/workflows/push.yml`”.

### 5.4 Repository detail page

Route: `/repo/<repo>`

The repository detail page should combine everything known about one repository:

- tracker row,
- direct internal dependencies,
- internal dependents,
- bump candidates,
- grouped issue log rows from `repo_issue_log`,
- issue timelines from `repo_issue_steps`,
- validations,
- raw events,
- PR URL,
- merge SHA,
- release URL,
- generated command snippets.

The issue section should come before the raw event stream. It should answer the human question first: what went wrong, how did we fix it, and what evidence proves the fix? The raw event stream should remain available below it as the audit trail.

This page is the replacement for repeatedly running ad hoc `sqlite3` queries.

### 5.5 Evidence page

Route: `/evidence`

The evidence page should list files under the ticket’s `sources/` directory by category:

- `07-ggg-batch-ready-*.json`,
- Codex trigger directories,
- Codex feedback directories,
- failed logs,
- report snapshots,
- generated PR manifests.

The first implementation can link to local paths as text. A later implementation can render JSON summaries.

### 5.6 Blocked and skipped page

Route: `/blocked`

This page should show blocked and skipped rows with notes and a recommended decision type:

- ownership decision,
- archived repository decision,
- module path decision,
- manual rollout needed,
- remove from rollout scope.

This keeps unresolved work visible without mixing it into release candidates.

## 6. Query Layer

Do not embed complex SQL directly in the HTML string. The next version of `02-rollout-tracker.py` should separate query functions from render functions.

Suggested structure:

```python
def query_overview(con): ...
def query_release_layers(con, train_name): ...
def query_bump_candidates(con): ...
def query_repo_detail(con, repo): ...
def render_overview(data): ...
def render_release_train(data): ...
def render_bumps(data): ...
def render_repo_detail(data): ...
```

This is not abstraction for its own sake. It makes the dashboard testable. A unit test can call `query_release_layers` against a small fixture database and verify that layer computation is displayed correctly.

## 7. Dependency Scanner Integration

The dependency scanner currently runs as a separate script:

```bash
ttmp/2026/05/29/INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases/scripts/01-populate-internal-dependencies.py
```

The dashboard should expose its last scan time. It should not automatically rescan on every page request. Scanning touches many working trees and may be slow or observe a repository while the operator is editing it.

The first integration should be a CLI command:

```bash
./scripts/02-rollout-tracker.py scan-deps
```

That command can call the same scanner logic or import a shared module. The dashboard can show:

```text
Dependency scan last run: 2026-05-29T17:36:00Z
```

A later version can add a button, but command-line refresh is safer for an operator tool.

## 8. State and Action Derivation

The dashboard should not require the operator to infer next actions from raw state. It should derive them.

### 8.1 Release action derivation

A repository is a release candidate when:

- `repos.state = 'main_actions_verified'`,
- no `release_url` is recorded,
- it appears in the earliest non-empty release layer for the selected train,
- all dependencies in `depends_on_json` either have release URLs or are outside the selected train.

### 8.2 Bump action derivation

A repository is a bump candidate when:

- it appears in `dependency_bump_candidates`,
- the candidate dependency has an available tag,
- the consumer is not blocked or skipped,
- the dependency edge is direct, or the operator explicitly enables indirect bumps.

### 8.3 Main verification action derivation

A repository needs main verification when:

- `state = 'merged'`, or
- `action_status` indicates pending main verification, or
- the most recent merge event has no later main verification event.

The last rule requires event timestamps. It is more robust than trusting `action_status` alone.

### 8.4 Codex action derivation

A PR needs Codex action when the latest `ggg` evidence says:

- `codex_feedback`: inspect comments and fix,
- `waiting_codex`: wait or trigger if stale,
- `ready`: merge if checks and mergeability are also satisfied.

The dashboard does not need to duplicate all of `ggg`. It should summarize and link to the evidence file.

## 9. UI Design

The UI should remain plain HTML and CSS. The purpose is fast operator comprehension, not application complexity.

### 9.1 Visual hierarchy

Use these visual groups:

1. **Summary cards** at the top.
2. **Action queues** immediately below the summary.
3. **Detailed tables** below action queues.
4. **Event/evidence tables** at the bottom or on detail pages.

The operator should see the next action without scrolling.

### 9.2 Status colors

Keep the existing state colors, but add semantic badges:

| Badge | Meaning |
|---|---|
| `release-now` | Earliest layer release candidate. |
| `bump-needed` | Downstream requires older internal version. |
| `wait-upstream` | Release depends on unreleased upstream. |
| `codex-current` | Current-head Codex feedback exists. |
| `main-pending` | Merge landed but main workflows need verification. |
| `external-failure` | Failure is unrelated to rollout gate, such as secret scanning. |

### 9.3 Tables over graphs first

A graph view is useful only after the tabular data is correct. Implement the release layer and bump tables first. Add graph visualization later for scoped exploration.

## 10. Implementation Plan

### Phase 1: Land the dependency schema and queries

- Keep `01-populate-internal-dependencies.py` as the reproducible scanner.
- Add `scan-deps` to `02-rollout-tracker.py` or factor scanner code into a shared module.
- Add query functions for `internal_modules`, `internal_dependency_edges`, `release_order_layers`, and `dependency_bump_candidates`.
- Add CLI commands:
  - `deps modules`,
  - `deps edges --repo REPO`,
  - `deps release-order --train TRAIN`,
  - `deps bumps`.

### Phase 2: Add release and bump dashboard pages

- Add routing based on request path.
- Implement `/release-train`.
- Implement `/bumps`.
- Link both pages from the overview.
- Add copyable commands for bump candidates.

### Phase 3: Add repository detail pages

- Make repository names clickable.
- Show dependencies, dependents, grouped issue logs, issue timelines, events, validations, and release actions.
- Add per-repo command snippets.
- Add filters for issue category and status so operators can focus on unresolved or high-severity issues.

### Phase 4: Add evidence and readiness panels

- Parse the latest `ggg batch ready` JSON snapshot.
- Show PR readiness queue if any PRs are open.
- Link to Codex feedback and trigger evidence files.
- Show main verification state derived from events and GitHub run URLs recorded in notes/events.

### Phase 5: Add health checks

- Add logcopter health scanner.
- Add Glazed lint health scanner.
- Add failure classification fields to events or a new `failure_events` table.
- Add blocked/skipped decision page.

## 11. Testing Strategy

The dashboard is small enough that most tests can use fixture databases.

### 11.1 Schema tests

Create a temporary SQLite DB, run schema creation, insert a minimal set of repositories, and verify that dependency scan tables can be populated.

### 11.2 Release layer tests

Use a fixture graph:

```text
A
B -> A
C -> B
D -> A
```

Expected layers:

```text
Layer 1: A
Layer 2: B, D
Layer 3: C
```

### 11.3 Bump candidate tests

Insert a consumer requiring `A v0.1.0` and an available tag `A v0.2.0`. Verify that `dependency_bump_candidates` contains one high-priority row.

### 11.4 Renderer smoke tests

Call each renderer and assert that the output contains expected repository names and no unescaped raw HTML from database fields.

### 11.5 Operator smoke test

Run against the INFRA-004 DB:

```bash
./scripts/02-rollout-tracker.py scan-deps
./scripts/02-rollout-tracker.py deps release-order --train verified_unreleased_tracker_upstreams
./scripts/02-rollout-tracker.py deps bumps
./scripts/02-rollout-tracker.py dashboard --port 8765
```

Then open the dashboard and verify that release layers and bump candidates match the SQL queries.

## 12. Design Decisions

### Decision 1: Keep SQLite as the source of truth

SQLite worked well during INFRA-004. It is inspectable, scriptable, and easy to back up with ticket artifacts. A server-side database would add operational complexity without solving the current problem.

### Decision 2: Store derived dependency tables in SQLite

The dashboard could compute dependency layers on every request, but storing them has two advantages. It gives the operator a stable snapshot, and it makes the data queryable from `sqlite3`. The scan timestamp makes staleness visible.

### Decision 3: Preserve both tracker upstreams and go.mod edges

Neither source is strictly superior. Tracker upstreams capture rollout intent. `go.mod` captures actual module requirements. The dashboard should show both and highlight disagreements.

### Decision 4: Use tables before graphs

Graphs are attractive, but release work is done row by row. A table can show action, version, tag, command, and evidence in one place. Graphs should be scoped and secondary.

### Decision 5: Keep command execution outside the dashboard

The dashboard should generate commands, not execute them. This keeps the tool safe and compatible with the existing operator workflow.

## 13. Alternatives Considered

### Alternative: Replace the dashboard with a full web application

A full web app would make complex UI easier, but it would introduce a build step, package dependencies, and more state. The current tool’s strength is that it can run inside a ticket directory with Python and SQLite.

### Alternative: Use only GitHub as the source of truth

GitHub has PRs, checks, and releases, but it does not know rollout tracks, skipped decisions, local validation notes, or dependency-layer intent. The SQLite tracker remains necessary.

### Alternative: Use only go.mod dependencies for release order

Actual `go.mod` dependencies are essential, but they do not capture every rollout constraint. The tracker’s `upstreams` field can include planning relationships and manually discovered dependencies. The dashboard should compare both.

### Alternative: Auto-run dependency scans on every dashboard refresh

Automatic scans would make the data fresh but risky. The scan walks local repositories and reads worktrees. Running it every ten seconds would be noisy and could capture half-finished edits. Manual scan refresh is safer.

## 14. Open Questions

1. Should `latest_local_tag` be replaced or supplemented with a remote GitHub tag lookup? Local tags are useful but may be stale.
2. Should skipped upstreams such as `geppetto` block dependency-converged release trains by default?
3. Should `dependency_bump_candidates` include indirect dependencies by default, or should indirect bumps be hidden unless requested?
4. Should release URLs be populated automatically from tags when `gh release view` succeeds?
5. Should `ggg` evidence be imported into SQLite, or should the dashboard read JSON evidence files directly?
6. Should secret scanning and image publishing failures get their own health category so they are visible but separate from rollout gates?

## 15. Definition of Done

The dashboard improvement is done when an operator can answer these questions without ad hoc shell queries:

- What repositories are ready to release now?
- What repositories must wait for an upstream tag?
- What downstream repositories should be bumped after this tag?
- What version does each downstream currently require?
- What evidence supports each repository’s current state?
- What blocked/skipped repositories require a manual decision?
- What is the next safest action for the rollout?

When those questions are visible in the dashboard, the tool will have moved from a status page to an operational release system.
