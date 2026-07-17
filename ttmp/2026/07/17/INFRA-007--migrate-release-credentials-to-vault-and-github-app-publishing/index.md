---
Title: Migrate release credentials to Vault and GitHub App publishing
Ticket: INFRA-007
Status: active
Topics:
    - github
    - release
    - automation
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://.github/workflows/publish-goreleaser-release.yml
      Note: Reusable Vault-backed merge publication workflow.
    - Path: repo://ttmp/2026/07/17/INFRA-007--migrate-release-credentials-to-vault-and-github-app-publishing/sources/01-goreleaser-split-and-merge.md
      Note: Official GoReleaser Pro split/merge source captured with Defuddle.
ExternalSources: []
Summary: Design and rollout plan for Vault-backed GoReleaser credentials and GitHub App Homebrew publishing.
LastUpdated: 2026-07-17T17:21:34.284068781-04:00
WhatFor: Track platform work to centralize release credential access in Vault.
WhenToUse: Use when planning or reviewing release workflow credential migration.
---


# Migrate release credentials to Vault and GitHub App publishing

## Overview

This ticket designs a shared platform migration from GitHub Actions release
secrets to Vault/OIDC-issued credentials. It preserves existing artifact
publication behavior while replacing the Homebrew tap PAT with a short-lived
GitHub App installation token. The ticket is currently in design review; no
Terraform, Vault, or release workflow behavior has been changed.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active — design package complete; implementation phases pending**

## Topics

- github
- release
- automation

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
