---
Title: Tasks
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
Summary: Task list for rolling out Glazed CLI policy linting across go-go-golems repositories.
LastUpdated: 2026-05-27T11:20:00-04:00
WhatFor: Track repository-by-repository rollout, validation, PR readiness, merge, and release work.
WhenToUse: During the Glazed linting release train.
---

# Tasks

## Phase 1: Ticket setup and inventory

- [x] Create INFRA-002 ticket.
- [x] Create rollout plan and diary documents.
- [x] Add repository inventory script under this ticket's `scripts/` directory.
- [x] Run inventory script and capture candidate repositories that depend on `github.com/go-go-golems/glazed`.
- [x] Record chosen target repository set and exclusions in the rollout plan.

## Phase 2: Apply Glazed lint wiring

- [x] Add Makefile `glazed-lint-build` / `glazed-lint` targets per target repository.
- [x] Wire `glazed-lint` into existing `lint` / `lintmax` targets where present.
- [x] Wire CI lint workflows to run `make glazed-lint` where appropriate.
- [x] Keep allow paths narrow and document every allow path.

## Phase 3: Validate and fix violations

- [x] Run `make glazed-lint` per target repository.
- [ ] Fix real CLI policy violations where practical.
- [x] Add narrow allow paths only for intentional legacy bridge/tool code.
- [ ] Run repo-specific validation (`make lintmax` or `make lint`, `GOWORK=off go test ./...`, and smoke tests where relevant).

## Phase 4: Commit, push, and PRs

- [ ] Commit each repository's focused lint rollout changes.
- [ ] Push branches and open PRs.
- [ ] Store PR list as YAML under this ticket's `scripts/` directory.
- [ ] Trigger Codex with `ggg pr codex-trigger --file`.
- [ ] Watch readiness with `ggg batch ready --watch`.

## Phase 5: Merge and release

- [ ] Merge PRs only after `ggg` reports readiness.
- [ ] Run `ggg release tag-patch --dry-run` for each merged repo.
- [ ] Push patch releases after approval / when safe.
- [ ] Verify Go proxy visibility.

## Phase 6: Closeout

- [ ] Update diary after every substantial step.
- [ ] Update changelog and relate changed files.
- [ ] Run `docmgr doctor --ticket INFRA-002 --stale-after 30`.
- [ ] Close the ticket after rollout completion.
