---
Title: Tasks
Ticket: INFRA-003
Status: active
Topics:
  - cli
  - automation
  - release
  - github
  - docsctl
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Task list for designing and rolling out docsctl documentation publishing for CLI packages that can export Glazed help SQLite databases.
LastUpdated: 2026-05-27T17:30:00-04:00
WhatFor: Track analysis, candidate inventory, implementation planning, validation, and handoff work for docsctl publishing rollout.
WhenToUse: Before editing package release workflows or Terraform/Vault docsctl publisher roles.
---

# Tasks

## Phase 1: Ticket setup and evidence gathering

- [x] Create INFRA-003 ticket workspace.
- [x] Create primary design/implementation guide document.
- [x] Create investigation diary document.
- [x] Read the existing `docsctl` publishing playbook and reusable workflow.
- [x] Inventory workspace repositories for `help export --format sqlite` capability.
- [x] Validate exported SQLite help databases with `docsctl validate --package ... --version ...`.
- [x] Capture repository IDs needed by Terraform/Vault publisher roles.
- [x] Capture current docs registry package/version state.

## Phase 2: Design and implementation guide

- [x] Explain docsctl publishing architecture for a new intern.
- [x] Map all participating systems: package CLI, Glazed help export, reusable GitHub Actions workflow, Vault OIDC, docs-registry, docs.yolo frontend, Terraform roles, and `ggg` rollout tooling.
- [x] Classify candidate packages by readiness and required work.
- [x] Provide rollout workflow templates and pseudocode.
- [x] Provide phased implementation plan, test strategy, and risk analysis.

## Phase 3: Bookkeeping and delivery

- [x] Relate key source/playbook/workflow files to the design and diary.
- [x] Update changelog.
- [x] Run `docmgr doctor --ticket INFRA-003 --stale-after 30`.
- [x] Upload the document bundle to reMarkable.

## Phase 4: Future implementation rollout

- [ ] Apply Terraform/Vault publisher roles for approved candidate packages.
- [ ] Add release workflow `publish-docs` jobs for approved candidate packages.
- [ ] Open PRs, trigger Codex with `ggg`, and wait for readiness.
- [ ] Merge approved PRs and publish docs on the next package release tags.
- [ ] Verify docs.yolo package/version visibility after releases.
