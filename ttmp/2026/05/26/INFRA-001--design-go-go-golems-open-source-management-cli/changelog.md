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

