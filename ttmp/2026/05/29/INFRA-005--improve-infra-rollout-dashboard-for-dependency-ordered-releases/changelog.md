# Changelog

## 2026-05-29

- Initial workspace created


## 2026-05-29 18:10 UTC — Initialized dependency-aware dashboard improvement ticket

- Created INFRA-005 for dashboard improvements driven by INFRA-004 rollout operations.
- Added `scripts/01-populate-internal-dependencies.py` and populated the INFRA-004 SQLite DB with normalized internal module/dependency data.
- Added derived tables: `internal_modules`, `internal_dependency_edges`, `release_order_layers`, and `dependency_bump_candidates`.
- Added analysis of dashboard improvement candidates and a design guide for dependency-aware release/bump workflows.
- Exported CSV/text snapshots of the new dependency tables under `sources/`.

## 2026-05-29 18:40 UTC — Added structured repository issue/fix history tables

- Added `scripts/02-populate-repo-issue-log.py` to derive grouped repository issue records and issue timelines from tracker `events` and `validations`.
- Extended the INFRA-004 SQLite DB with `repo_issue_log` and `repo_issue_steps`.
- Populated 342 issue-log rows and 808 issue timeline steps.
- Exported issue-log snapshots under `sources/07-repo-issue-log.csv`, `sources/08-repo-issue-steps.csv`, `sources/09-issue-log-summary.txt`, and `sources/10-sample-repo-issue-details.txt`.
- Updated the dashboard analysis and design guide to include issue/fix history on repository detail pages.

## 2026-05-29 19:15 UTC — Implemented first dependency-aware dashboard pages

- Extended the INFRA-004 rollout tracker script with dependency/release CLI queries: `deps-modules`, `deps-edges`, `deps-release-order`, `deps-bumps`, `deps-scan`, `issue-list`, and `issue-refresh`.
- Added dashboard routes for `/release-train`, `/bumps`, and `/repo?repo=<name>`.
- Repository detail pages now render tracker state, dependencies, dependents, bump candidates, grouped issue/fix history, issue timelines, validations, and raw events.
- Validated CLI queries and HTML render functions against the populated INFRA-004 SQLite DB.
- Code commit: `0973b71b5807b87140be0719795e104fc4a01b00`.

## 2026-05-29 19:45 UTC — Added logcopter and Glazed health panels

- Added `scripts/03-populate-repo-health.py` to derive lightweight logcopter and Glazed lint health checks from local repository files.
- Extended the INFRA-004 SQLite DB with `repo_health_checks` and populated 623 health-check rows.
- Added dashboard route `/health` and health-check sections on `/repo?repo=<name>` pages.
- Added tracker CLI commands `health-refresh` and `health-list`.
- Exported health snapshots under `sources/11-repo-health-checks.csv`, `sources/12-health-summary.txt`, and `sources/13-health-warnings-sample.txt`.
