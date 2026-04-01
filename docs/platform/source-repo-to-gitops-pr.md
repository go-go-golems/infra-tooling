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

## Standard Image-Based Variant

The most common K3s application path is not the federated-remote model. It is
the simpler image-based model:

```text
source repo
  -> run tests
  -> build Docker image
  -> publish immutable GHCR tag
  -> patch image field in GitOps deployment manifest
  -> open/update GitOps PR
  -> merge PR
  -> Argo CD reconciles
```

This repo now carries the reusable building blocks for that path:

- workflow template:
  - `templates/github/publish-image-ghcr.template.yml`
- example target metadata:
  - `examples/platform/image-gitops-targets.example.json`
- GitOps PR helper:
  - `scripts/gitops/open_gitops_pr.py`

## Required Source-Repo Files For The Image Path

An app repo using this model should normally have:

- `Dockerfile`
- `.github/workflows/publish-image.yaml`
- `deploy/gitops-targets.json`
- `scripts/open_gitops_pr.py`

The workflow template in `templates/github/publish-image-ghcr.template.yml`
assumes that the source repo keeps a local copy of `scripts/open_gitops_pr.py`.
That duplication is intentional. GitHub Actions should not depend on another
repository being cloned at runtime just to open the deployment PR.

## The `deploy/gitops-targets.json` Contract

For image-based deployments the target metadata should match the example in
`examples/platform/image-gitops-targets.example.json`.

Each target describes:

- `name`
  - human-readable deployment target name
- `gitops_repo`
  - destination GitOps repository
- `gitops_branch`
  - branch to patch and branch PRs against
- `manifest_path`
  - YAML file containing the image field to update
- `container_name`
  - exact Kubernetes container name inside that manifest

Example:

```json
{
  "targets": [
    {
      "name": "my-app-prod",
      "gitops_repo": "wesen/2026-03-27--hetzner-k3s",
      "gitops_branch": "main",
      "manifest_path": "gitops/kustomize/my-app/deployment.yaml",
      "container_name": "my-app"
    }
  ]
}
```

## Private GHCR Boundary

Publishing to GHCR is not enough if the package stays private.

For a private package, the GitOps repo and cluster also need the image-pull
side wired:

- a Vault secret containing registry credentials
- a `VaultStaticSecret` rendering a `kubernetes.io/dockerconfigjson` secret
- a service account referencing that pull secret

If those are missing, the GitOps PR can merge successfully and Argo can still
end up in `ImagePullBackOff`.

This is a cluster-side concern, not a source-repo workflow concern. Keep that
boundary explicit.

## Immutable Tag Rule

The workflow should patch immutable image tags, not floating tags.

Recommended tag shape:

```text
ghcr.io/<owner>/<repo>:sha-<short-sha>
```

Do not open GitOps PRs that point at `main` or `latest` for normal application
rollouts. Those tags are useful for human debugging, but they weaken rollback
and reviewability.

## First-Rollout Reminder

The bootstrap rule still applies to the image-based path.

If the app is brand new in the cluster, someone still needs to create the Argo
`Application` object once:

```bash
kubectl apply -f gitops/applications/<app>.yaml
kubectl -n argocd annotate application <app> argocd.argoproj.io/refresh=hard --overwrite
```

After that, the normal image-bump PR flow is enough.
