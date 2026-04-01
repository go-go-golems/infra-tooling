# Federated Remote Release Model

This document defines the reusable release model for browser-loaded federated remotes.

It is extracted from active implementation work in:

- `wesen-os`
- `go-go-app-inventory`
- `2026-03-27--hetzner-k3s`

The purpose is to give future source repos one canonical model to follow.

## Release Flow

```text
source repo
  -> build remote artifact
  -> upload immutable files to object storage
  -> compute manifest URL
  -> patch host GitOps target
  -> open/update GitOps PR
  -> merge PR
  -> Argo sync
  -> deployed host loads remote
```

## Ownership Boundaries

### Source repo owns

- remote build
- object-storage upload
- target metadata
- GitOps PR creation

### GitOps repo owns

- canonical host registry layout
- canonical target paths
- deployment state

### infra-tooling owns

- reusable scripts
- reusable workflow templates
- example metadata
- extracted operational docs

## Shared Helpers

The current shared helper split is:

- immutable remote upload:
  - [publish_federation_remote.py](../../scripts/federation/publish_federation_remote.py)
- host registry patch + GitOps PR handoff:
  - [open_federation_gitops_pr.py](../../scripts/federation/open_federation_gitops_pr.py)

That means a source repo only needs to keep:

- its app-specific build step
- its repo-local `deploy/federation-gitops-targets.json`
- its workflow wiring and secret/variable names

## Deployment Unit

The deployment unit is the immutable manifest URL, for example:

```text
https://<bucket>.<region>.your-objectstorage.com/remotes/<remote-id>/versions/<version>/mf-manifest.json
```

Do not deploy moving aliases like `latest` directly in production host config.

## Generic Metadata Shape

See:

- [example target file](../../examples/federation/federation-gitops-targets.example.json)

The core fields are:

- `kind`
- `remote_id`
- `gitops_repo`
- `gitops_branch`
- `target_file`
- `config_key`

## Generic Patch Behavior

The host config patch operation is:

1. parse the target YAML file
2. locate the embedded `federation.registry.json`
3. find the matching `remoteId`
4. replace `manifestUrl`
5. ensure `enabled=true`

The reusable helper for that lives in:

- [patch_federation_registry_target.py](../../scripts/federation/patch_federation_registry_target.py)
