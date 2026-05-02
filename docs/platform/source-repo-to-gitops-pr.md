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

This document only covers the release and GitOps handoff portion of the
system. If the target app needs identity, Vault-backed runtime secrets, VSO, or
private-image pull credentials, the cluster-side prerequisites live elsewhere:

- [app-runtime-secrets-and-identity-provisioning-playbook.md](/home/manuel/code/wesen/2026-03-27--hetzner-k3s/docs/app-runtime-secrets-and-identity-provisioning-playbook.md)

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

- reusable workflow:
  - `.github/workflows/publish-ghcr-image.yml`
- workflow template:
  - `templates/github/publish-image-ghcr.template.yml`
- example target metadata:
  - `examples/platform/image-gitops-targets.example.json`
- GitOps PR action:
  - `actions/open-gitops-pr/`
- local validator:
  - `scripts/gitops/validate_gitops_targets.py`

The release chain is only one half of the onboarding story. The K3s repo still
needs:

- a GitOps package and Argo `Application`
- Vault auth and VSO wiring
- seeded runtime and image-pull secrets
- any DB bootstrap resources
- matching Keycloak configuration in Terraform when the app uses OIDC

## Required Source-Repo Files For The Image Path

An app repo using this model should normally have:

- `Dockerfile`
- `.github/workflows/publish-image.yaml`
- `deploy/gitops-targets.json`

An app repo using the current shared model should not need to copy the GitOps
PR helper script anymore. Instead, the caller workflow should invoke the
reusable workflow in this repository, and that workflow checks out the
versioned `infra-tooling` repo contents so it can call the packaged
`open-gitops-pr` action.

Recommended source-repo reuse points:

- keep `deploy/gitops-targets.json` local
- keep `Dockerfile` and test command local
- call `go-go-golems/infra-tooling/.github/workflows/publish-ghcr-image.yml@<ref>`
- use `examples/platform/publish-image-ghcr.caller.example.yml` as the caller reference

## Secret Expectations

The shared workflow depends on two different credential paths:

- `GITHUB_TOKEN`
  - used by `docker/login-action` to publish to GHCR
  - automatically available in GitHub Actions workflows
- GitHub Actions OIDC token
  - requested by the caller workflow through `permissions.id-token: write`
  - presented to Vault through `hashicorp/vault-action@v3`
  - exchanged at Vault `auth/github-actions` for a short-lived Vault token
- Vault-stored GitOps PR token
  - read from a repo-specific KV v2 data path such as `kv/data/ci/github/my-app/gitops-pr-token`
  - exported as `GITOPS_PR_TOKEN` for the packaged `open-gitops-pr` action
  - used by the action to clone, push, and open pull requests

The default shared workflow path is now `gitops_pr_token_source: vault`. When
`open_gitops_pr` is true, callers must provide `vault_role` and
`vault_secret_path`. The legacy `gitops_pr_token_source: secret` mode still
exists only for deliberate migration work; new repositories should not store a
long-lived `GITOPS_PR_TOKEN` directly as a source-repo GitHub secret.

Caller workflows must grant OIDC permission:

```yaml
permissions:
  contents: read
  packages: write
  pull-requests: write
  id-token: write
```

Caller workflow example:

```yaml
jobs:
  release:
    uses: go-go-golems/infra-tooling/.github/workflows/publish-ghcr-image.yml@main
    secrets: inherit
    with:
      gitops_pr_token_source: vault
      vault_role: my-app-gitops-pr
      vault_secret_path: kv/data/ci/github/my-app/gitops-pr-token
      open_gitops_pr: ${{ github.event_name != 'pull_request' && github.ref == 'refs/heads/main' }}
```

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

## Validation Before Publish

The shared workflow assumes the repo-local `deploy/gitops-targets.json`
contract is valid. For local verification or a lightweight lint step, run:

```bash
python3 scripts/gitops/validate_gitops_targets.py deploy/gitops-targets.json
```

That validator only checks the target metadata contract. It does not verify that
the referenced GitOps manifest exists in a remote repository checkout.

## First-Rollout Reminder

The bootstrap rule still applies to the image-based path.

If the app is brand new in the cluster, someone still needs to create the Argo
`Application` object once:

```bash
kubectl apply -f gitops/applications/<app>.yaml
kubectl -n argocd annotate application <app> argocd.argoproj.io/refresh=hard --overwrite
```

After that, the normal image-bump PR flow is enough.
