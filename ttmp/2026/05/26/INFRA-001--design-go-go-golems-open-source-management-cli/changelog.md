# Changelog

## 2026-05-26

- Initial workspace created


## 2026-05-26

Created initial intern-oriented design guide for a future go-go-golems open-source management CLI, including current script inventory, GitHub/Codex APIs, release-train workflows, data model, CLI command groups, pseudocode, diagrams, gaps, and implementation phases.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/design-doc/01-go-go-golems-open-source-management-cli-design.md — Primary analysis and implementation guide
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/sources/02-xgoja-015-script-evidence.txt — Line-anchored evidence for XGOJA-015 helper scripts


## 2026-05-26

Uploaded INFRA-001 design bundle to reMarkable at /ai/2026/05/26/INFRA-001 after retrying with a longer timeout; diary updated with upload evidence and timeout note.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/reference/01-investigation-diary.md — Diary records design writing and reMarkable upload


## 2026-05-27

Created and printed a thermal almanach for the ggg project using image separators (owl, lighthouse, fox, windmill); remote render dry-run succeeded at 384x2361 and final print succeeded in 4 bitmap segments.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/various/ggg-project-almanach.json — Printed almanach layout with embedded separator images


## 2026-05-27

Reprinted the ggg project almanach at bodyScale 1.7 per user request; dry-run rendered 384x2880 and final print succeeded in 4 bitmap segments.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/various/ggg-project-almanach.json — Updated bodyScale to 1.7 for larger print


## 2026-05-27

Updated design for YAML PR lists, safe Codex trigger --force behavior, Glazed command output, and release tag-minor/tag-major verbs; implemented initial ggg CLI scaffold, pr codex-trigger, YAML PR-list parsing, and release tag patch/minor/major commands.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/cmd/ggg/main.go — New ggg CLI entry point
- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/codex_trigger.go — Initial Glazed Codex trigger command with --force and YAML input
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prlist/prlist.go — YAML PR-list parser
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/release/release.go — Initial release tag helper for patch/minor/major


## 2026-05-27

Implemented first Go PR readiness parity slice with pkg/prready classifier, GitHub GraphQL readiness query, and Glazed ggg pr ready command; go test ./... and live JSON smoke against Discord Bot PR 9 passed.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/ready.go — Glazed pr ready command
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/ghclient/readiness.go — GitHub GraphQL readiness query and decoding
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/prready.go — Go readiness classifier for checks and Codex signals


## 2026-05-27

Implemented ggg batch ready for YAML PR lists with watch flags, per-PR rows, summary row, and live JSON smoke against Discord Bot PR 9.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/batch/ready.go — Glazed YAML batch readiness command
- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/root.go — Registered real batch command group


## 2026-05-27

Added readiness exit-code parity using internal/exitcode.Error; ggg pr ready and ggg batch ready now emit rows and then return script-compatible non-ready codes.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/cmd/ggg/main.go — Root command maps typed exit errors to os.Exit(code)
- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/batch/ready.go — Batch readiness command now returns summary-based exit codes
- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/ready.go — PR readiness command now returns non-ready exit codes
- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/exitcode/exitcode.go — Typed exit-code error for script-compatible command exits


## 2026-05-27

Added detailed Codex/release hardening tasks and implemented shared Codex snapshot parsing, ggg pr codex-comments, safer Codex trigger skip behavior, and hardened release tag options/guardrails.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/codex_comments.go — Structured Codex review/comment output command
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/codex_helpers.go — Shared Codex signal helper model
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/release/release.go — Hardened release tag implementation with dirty checks
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/tasks.md — Added detailed Phase 9 and Phase 10 hardening tasks


## 2026-05-27

Implemented Codex recent-trigger cooldown and GraphQL truncation reporting; current-head truncated Codex review comments are classified conservatively and codex-trigger rows now include recent trigger fields.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/cli/pr/codex_trigger.go — Recent trigger cooldown flag and row fields
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/ghclient/readiness.go — GraphQL pageInfo/truncation decoding
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/codex_helpers.go — Recent trigger helper


## 2026-05-27

Opened live readiness test PRs 5, 6, and 7; triggered Codex, added synthetic statuses, verified ready/failed_checks/codex_feedback classifications, and fixed failed-check-kind and exit-code handling discovered during live testing.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/internal/exitcode/exitcode.go — Changed exit handling to requested process exit codes after row emission
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/prready.go — Fixed failedCheckKinds to only inspect check-related failures
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/scripts/01-create-readiness-test-prs.sh — Creates live readiness test PRs
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/scripts/03-set-readiness-test-statuses.sh — Adds synthetic StatusContext results for live PR tests


## 2026-05-27

Added minimal prready.Snapshot JSON fixtures and table-driven classifier tests for ready, failed checks, current-head Codex feedback, running Codex, stale feedback, and truncated feedback states.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/fixture_test.go — Table-driven readiness fixture tests
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/testdata/codex_feedback_current_head.json — Codex feedback fixture derived from live unsafe PR scenario
- /home/manuel/code/wesen/go-go-golems/infra-tooling/pkg/prready/testdata/ready.json — Ready fixture derived from live readiness-control scenario


## 2026-05-27

Added and ran a cleanup script for live readiness test PRs; closed PRs 5, 6, and 7 and deleted their disposable test branches after fixture creation.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/scripts/04-cleanup-readiness-test-prs.sh — Cleanup script for disposable live readiness PRs


## 2026-05-27

Closed INFRA-001 after implementing and live-validating the initial ggg management CLI slices; remaining unchecked tasks are future hardening/backlog items.

