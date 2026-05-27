---
Title: Tasks
Ticket: INFRA-001
Status: active
Topics:
  - cli
  - github
  - release
  - automation
DocType: reference
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Phased task list for researching, designing, and implementing a go-go-golems open-source management CLI.
LastUpdated: 2026-05-26T23:35:00-04:00
WhatFor: Track analysis, design, implementation, validation, and reMarkable delivery.
WhenToUse: Use while executing INFRA-001.
---

# Tasks

## Analysis and design

- [x] Inventory current infra-tooling scripts, docs, examples, and generated evidence from XGOJA-015.
- [x] Map GitHub APIs, GraphQL fields, gh CLI calls, PR readiness states, reactions, checks, and merge/release interactions.
- [x] Map Go module/release-train APIs and workflows: tags, svu, go list proxy verification, GOWORK=off validation, dependency graph ordering.
- [x] Identify missing functionality, brittleness, and future CLI/package building blocks.
- [x] Write intern-oriented design/implementation guide with file references, API sketches, pseudocode, diagrams, and phased plan.
- [x] Keep chronological investigation diary in the ticket.
- [x] Relate key repository files and script evidence to the ticket docs.
- [x] Run docmgr doctor and resolve metadata/vocabulary issues.
- [x] Upload the initial design bundle to reMarkable and verify successful upload output.

## Phase 1: CLI scaffold and Glazed command foundation

- [x] Initialize a Go module in `infra-tooling`.
- [x] Add Glazed, Cobra, and YAML dependencies.
- [x] Create `cmd/ggg/main.go`.
- [x] Create root command and command groups: `pr`, `batch`, `repo`, `release`, and `train`.
- [x] Ensure new verbs are Glazed commands emitting row-oriented data.
- [x] Add concise human defaults and compatibility for `--with-structured-output`.
- [x] Validate with `go test ./...` and a command-tree smoke invocation.

## Phase 2: PR references, YAML PR lists, and Codex trigger safety

- [x] Implement PR reference parsing for URLs and `owner/repo#number`.
- [x] Implement YAML PR-list loading for string and object entries.
- [x] Implement a `GitHubClient` interface and initial `gh`-backed client.
- [x] Implement Codex in-progress detection via `EYES` reactions.
- [x] Implement `ggg pr codex-trigger` with `--force`, `--dry-run`, and `--file prs.yaml`.
- [x] Add tests for PR parsing and YAML PR-list loading.

## Phase 3: PR readiness parity

- [x] Port the Python GraphQL readiness query and typed decoding.
- [x] Port check classification.
- [x] Port Codex signal collection, stale reviewed-commit detection, and inline review comment extraction.
- [x] Implement `ggg pr ready` with Glazed row output and current state names.
- [x] Preserve current exit-code semantics.
- [ ] Add golden fixtures for observed XGOJA-015 states.

## Phase 4: Batch readiness with YAML input

- [x] Implement `ggg batch ready prs.yaml`.
- [x] Support `--watch`, `--interval`, `--timeout`, and `--trigger-missing-codex`.
- [x] Preserve batch exit codes including partial-ready exit `5`.
- [ ] Add tests for aggregation and partial readiness.

## Phase 5: Release verbs and Go module verification

- [ ] Implement module-path detection from `go.mod`.
- [ ] Implement highest semver tag discovery and next patch/minor/major calculation.
- [ ] Implement `ggg release tag-patch`.
- [ ] Implement `ggg release tag-minor`.
- [ ] Implement `ggg release tag-major`.
- [ ] Add guardrails for clean worktree, target commit, pushing only the new tag, and Go proxy verification.
- [ ] Add temporary-git-repo tests for release calculation.

## Phase 6: Validation profiles

- [ ] Define validation profile YAML schema.
- [ ] Implement validation runner with env, workdir, timeout, dry-run, and log capture.
- [ ] Implement `ggg repo validate`.
- [ ] Port XGOJA-015 focused validations into a sample profile.

## Phase 7: Release-train orchestration

- [ ] Define release-train YAML schema.
- [ ] Implement dependency graph loading and topological sort.
- [ ] Implement `ggg train status`.
- [ ] Implement `ggg train next`.
- [ ] Add merge gates for readiness and visible upstream tags.

## Phase 8: Reporting and docmgr integration

- [ ] Generate Markdown release reports from run-state files.
- [ ] Generate docmgr changelog snippets.
- [ ] Evaluate whether reMarkable upload should remain manual/documented or become a CLI command.

## Phase 9: Codex comment and trigger hardening

- [x] Refactor Codex signal parsing so readiness and trigger safety share one model.
- [x] Make `ggg pr codex-trigger` use the shared readiness/Codex snapshot instead of the simplified `CodexStatus` query.
- [x] Add precise trigger actions and skip reasons: `triggered`, `would_trigger`, `skipped_running`, `skipped_current_feedback`, and `skipped_recent_trigger`.
- [x] Add `ggg pr codex-comments` to emit Codex-authored review bodies and inline review comments as structured rows.
- [x] Include reviewed commit, current/stale status, path, line, body preview/full body, and URL in Codex comment rows.
- [x] Add tests for stale-vs-current Codex inline comments and trigger skip decisions.
- [x] Document pagination limitations for review comments and add a follow-up task for GraphQL pagination.

## Phase 10: Release command hardening

- [x] Add clean-worktree checks before tagging; refuse dirty repos unless `--allow-dirty` is set.
- [x] Add explicit `--commit` / `--from` target support with default `origin/main`.
- [x] Add `--yes` confirmation guard for non-dry-run tag pushes.
- [x] Detect existing tag collisions and verify whether the tag already points at the expected commit.
- [x] Keep pushing only `refs/tags/<tag>` and never broad `git push --tags`.
- [x] Add Go proxy verification retry/backoff with structured status rows.
- [x] Expand dry-run output to include module path, current tag, next tag, target commit, dirty status, and planned commands.
- [ ] Add temporary-git-repo tests for patch/minor/major and existing-tag behavior.
