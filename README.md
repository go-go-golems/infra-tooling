# infra-tooling

Shared deployment and release tooling for the go-go-golems / wesen platform.

This repo is the neutral home for reusable mechanics that should not live:

- in a single app repo like `go-go-app-inventory`
- in the host repo like `wesen-os`
- or in the GitOps repo like `2026-03-27--hetzner-k3s`

## What Belongs Here

- shared release and GitOps helper scripts
- reusable workflow templates
- example target metadata files
- extracted platform docs that multiple source repos should follow
- generic source-repo to GitOps PR automation

## What Does Not Belong Here

- app-specific business logic
- live GitOps state
- repo-specific one-off workflows that cannot be reused

## Current Focus

The first extracted toolkit in this repo is the federated remote release flow:

- source repo builds remote artifact
- uploads immutable files to object storage
- computes manifest URL
- updates GitOps target
- opens or updates a GitOps PR

## Layout

```text
docs/federation/
docs/platform/
examples/federation/
templates/github/
scripts/federation/
scripts/gitops/
```

## First Extracted Toolkits

### Federation remote release

- release model and secret bootstrap docs
- `federation-manifest` target example
- host-registry patch/update helpers
- GitHub workflow template for remote publish + GitOps handoff

### Source repo -> GitOps PR flow

- generic GitOps PR opener for image-based deployments
- extracted control-plane documentation from the Hetzner K3s repo
- example image target metadata for `deploy/gitops-targets.json`
- reusable GitHub Actions workflow template for GHCR publish + GitOps PR handoff

## Current Recommended Reuse Points

If a source repo is deploying a container image through the Hetzner K3s +
Argo CD platform, start with:

- `docs/platform/source-repo-to-gitops-pr.md`
- `examples/platform/image-gitops-targets.example.json`
- `templates/github/publish-image-ghcr.template.yml`
- `scripts/gitops/open_gitops_pr.py`
