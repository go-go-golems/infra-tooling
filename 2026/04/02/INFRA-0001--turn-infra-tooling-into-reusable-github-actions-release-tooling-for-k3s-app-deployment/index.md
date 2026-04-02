---
Title: Turn infra-tooling into reusable GitHub Actions release tooling for K3s app deployment
Ticket: INFRA-0001
Status: active
Topics:
    - github-actions
    - ghcr
    - gitops
    - argocd
    - k3s
    - platform
    - release-engineering
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md
      Note: Repo charter and current recommended reuse points
    - Path: /home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md
      Note: Current platform control-plane model and current duplication assumption
    - Path: /home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml
      Note: Current GHCR publish and GitOps handoff template
    - Path: /home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py
      Note: Current extracted GitOps PR helper
ExternalSources:
    - https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
    - https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
    - https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action
    - https://docs.github.com/en/actions/sharing-automations/creating-actions/metadata-syntax-for-github-actions
Summary: "Research ticket for turning infra-tooling from a copy-template repository into a versioned reusable GitHub Actions toolkit for GHCR publish, GitOps PR handoff, and future K3s app onboarding."
LastUpdated: 2026-04-02T10:27:44.390378379-04:00
WhatFor: "Use this ticket to understand how the current image-based release path works, what should become a reusable workflow versus an action image, and how to implement that extraction safely."
WhenToUse: "Read this before adding more publish-image workflows to app repos or before converting infra-tooling into a reusable GitHub Actions product surface."
---

# Turn infra-tooling into reusable GitHub Actions release tooling for K3s app deployment

## Overview

This ticket documents how to turn `infra-tooling` into the reusable home for the image-based deployment path used by go-go-golems applications that deploy to the Hetzner K3s platform. The practical problem is that the shared logic has already been extracted conceptually, but source repositories still copy workflow YAML and helper scripts locally instead of calling versioned reusable automation from this repository.

The main conclusion is that `infra-tooling` should remain the home for this work. The change needed is not another repository. The change is a packaging shift:

- keep design docs, examples, and helper code in `infra-tooling`
- add a versioned reusable workflow under `.github/workflows/`
- add a versioned action under `actions/`
- keep app-specific files such as `Dockerfile`, tests, and `deploy/gitops-targets.json` inside each source repository

The main design guide is:

- [01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md](./design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md)

Supporting investigation log:

- [01-investigation-diary.md](./reference/01-investigation-diary.md)

## Key Links

- Design guide: [01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md](./design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md)
- Diary: [01-investigation-diary.md](./reference/01-investigation-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current state of the work in this ticket:

- analysis complete
- design and implementation guide written
- implementation of reusable workflow/action not yet started

## Key Findings

- `infra-tooling` already declares itself as the neutral home for shared release and GitOps helper scripts, reusable workflow templates, example target metadata, and generic source-repo to GitOps PR automation.
- The repository already contains the required raw materials:
  - `docs/platform/source-repo-to-gitops-pr.md`
  - `templates/github/publish-image-ghcr.template.yml`
  - `scripts/gitops/open_gitops_pr.py`
  - `examples/platform/image-gitops-targets.example.json`
- The current platform doc explicitly assumes source repositories copy `scripts/open_gitops_pr.py` locally. That assumption was reasonable for the first extraction pass but is now the main source of drift.
- The correct packaging split is:
  - reusable workflow for CI orchestration
  - Docker-backed action for the GitOps PR updater toolchain
  - repo-local target metadata for deployment-specific information
- GitHub's own model supports this split:
  - reusable workflows live in `.github/workflows` and are invoked via `workflow_call`
  - composite actions package steps
  - Docker actions package toolchains and runtime dependencies

## Tasks

See [tasks.md](./tasks.md) for the phased implementation breakdown. The open tasks are intentionally implementation-facing so an intern can pick this up and build it incrementally.

## Changelog

See [changelog.md](./changelog.md) for delivery history and design milestones.
