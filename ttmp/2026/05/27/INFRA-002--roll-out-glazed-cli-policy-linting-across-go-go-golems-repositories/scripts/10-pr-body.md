---
Title: Glazed lint rollout PR body
Ticket: INFRA-002
Status: active
Topics:
  - cli
  - automation
  - release
  - github
DocType: reference
Intent: temporary
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Reusable pull request body text for INFRA-002 Glazed lint rollout PRs.
LastUpdated: 2026-05-27T13:00:00-04:00
WhatFor: Input file used by the PR creation script.
WhenToUse: When recreating or auditing INFRA-002 PR creation.
---

## Summary

- add or normalize `make glazed-lint`
- wire Glazed CLI policy linting into local lint targets and CI lint workflow where needed
- keep legacy allow paths narrow for existing command bridge/tool code

## Validation

- `make glazed-lint`

This PR is part of INFRA-002. Please do not merge until the rollout batch is reviewed.
