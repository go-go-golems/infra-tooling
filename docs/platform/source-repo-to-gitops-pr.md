# Source Repo To GitOps PR Flow

This document captures the reusable deployment control-plane model that was first documented in the Hetzner K3s repo and then exercised in:

- `draft-review`
- `wesen-os`
- `mysql-ide`
- `coinvault`

The point is to make the control-plane boundaries explicit for future source repos.

## The Release Chain

```text
source repo
  -> test and build in CI
  -> publish immutable artifact
  -> patch GitOps target
  -> open or update GitOps PR
  -> merge PR
  -> Argo CD reconciles
  -> Kubernetes rolls out
```

Do not collapse these steps mentally. Each arrow is a contract boundary with its own failure modes.

## Ownership Boundaries

### Source repo owns

- source code
- tests
- artifact build logic
- publish workflow
- target metadata
- helper that opens or updates GitOps PRs

### GitOps repo owns

- desired deployment state
- canonical manifest locations
- Argo `Application` objects
- namespace, ingress, and secret topology

### Cluster owns

- actual pods
- service routing
- TLS
- rollout health
- reconciliation status

## Most Common Misunderstanding

Publishing an artifact is not deployment.

Updating the GitOps repo is the deployment handoff. Argo CD only reacts to changes in the GitOps repo that are already reachable through an existing `Application` object.

For a brand-new app, the first rollout still requires one bootstrap step:

```bash
kubectl apply -f gitops/applications/<app>.yaml
kubectl -n argocd annotate application <app> argocd.argoproj.io/refresh=hard --overwrite
```

After that, normal GitOps PR merges are enough for Argo to reconcile the app.

## Why This Belongs Here

This model is not specific to one host repo or one cluster repo. It is the reusable operating contract for source repos that deploy through:

- GitHub Actions
- a package registry or object storage
- a separate GitOps repo
- Argo CD

Future repo-specific docs should reference this file instead of re-explaining the entire chain from scratch.
