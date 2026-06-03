---
Title: Migrate GitOps PR automation from PAT tokens to GitHub App tokens
Slug: github-app-gitops-pr-migration-playbook
Short: A concise operator playbook for moving an app repository from a Vault-stored GitOps PR PAT to the GitHub App installation-token flow.
Topics:
- github-actions
- github-apps
- gitops
- vault
- oidc
- ci-cd
- go-go-golems
Commands:
- gh
- vault
- git
- jq
- openssl
Flags:
- source_repo
- vault_role
- app_secret_path
- gitops_repo
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

# Migrate GitOps PR automation from PAT tokens to GitHub App tokens

Use this playbook when an application repository already publishes images through `go-go-golems/infra-tooling/.github/workflows/publish-ghcr-image.yml`, but its GitOps PR step still reads a long-lived token from Vault, for example:

```yaml
gitops_pr_token_source: vault
vault_role: <app>-gitops-pr
vault_secret_path: kv/data/ci/github/<app>/gitops-pr-token
```

The target state is:

```yaml
gitops_pr_token_source: github_app
vault_role: <app>-gitops-pr
gitops_app_secret_path: kv/data/ci/github/<app>/gitops-pr-app
gitops_app_owner: wesen
gitops_app_repositories: 2026-03-27--hetzner-k3s
```

The workflow still authenticates to Vault with GitHub Actions OIDC. The difference is that Vault now returns a GitHub App ID and private key, not a PAT. The reusable workflow mints a short-lived GitHub App installation token and passes it to `actions/open-gitops-pr` as `GITOPS_PR_TOKEN`.

## Target architecture

```text
GitHub Actions push on source repo main
  -> Vault JWT login with repo-bound role
  -> read kv/data/ci/github/<app>/gitops-pr-app
  -> actions/create-github-app-token mints installation token
  -> open-gitops-pr clones/pushes wesen/2026-03-27--hetzner-k3s
  -> GitOps PR updates the target manifest image
```

## Preconditions

- `go-go-golems/infra-tooling` is referenced at a revision that supports `gitops_pr_token_source: github_app`.
- The GitHub App is installed on the GitOps repository, normally `wesen/2026-03-27--hetzner-k3s`.
- The App has at least:
  - Contents: read/write
  - Pull requests: read/write
  - Metadata: read-only
- The source workflow has `permissions.id-token: write`.
- The existing Vault role already admits trusted `push` runs on `refs/heads/main` for the source repository.

## Step 1: identify the current source workflow

In the app repo, inspect the publish workflow:

```bash
rg -n "gitops_pr_token_source|vault_secret_path|vault_role|publish-ghcr-image" .github/workflows -S
```

Record:

| Field | Example |
|---|---|
| source repo | `wesen/2026-03-16--gec-rag` |
| workflow path | `.github/workflows/publish-image.yaml` |
| Vault role | `coinvault-gitops-pr` |
| old PAT path | `kv/data/ci/github/coinvault/gitops-pr-token` |
| new App path | `kv/data/ci/github/coinvault/gitops-pr-app` |
| GitOps repo | `wesen/2026-03-27--hetzner-k3s` |

## Step 2: store GitHub App credentials in Vault

If the same GitHub App is already used by another source repo, copy its `app_id` and `private_key` into an app-specific Vault path. Keep one source-repo-specific secret path per workflow role so Vault policies stay narrow.

Example using an existing working App credential:

```bash
export VAULT_ADDR=https://vault.yolo.scapegoat.dev
vault login -method=oidc role=operators
export VAULT_TOKEN="$(<"$HOME/.vault-token")"

src_path=kv/ci/github/retro-obsidian-publish/gitops-pr-app
dst_path=kv/ci/github/<app>/gitops-pr-app

tmpdir="$(mktemp -d)"
vault kv get -format=json "$src_path" > "$tmpdir/app.json"
app_id="$(jq -r '.data.data.app_id' "$tmpdir/app.json")"
jq -r '.data.data.private_key' "$tmpdir/app.json" > "$tmpdir/private-key.pem"

vault kv put "$dst_path" \
  app_id="$app_id" \
  private_key=@"$tmpdir/private-key.pem"

rm -rf "$tmpdir"
```

Verify shape without printing the key:

```bash
vault kv get -format=json kv/ci/github/<app>/gitops-pr-app \
  | jq '{keys:(.data.data|keys), app_id:.data.data.app_id, private_key_len:(.data.data.private_key|length)}'
```

Expected keys:

```text
app_id
private_key
```

## Step 3: update the Vault policy for the source workflow role

In the K3s GitOps repo, edit the app-specific GitHub Actions policy:

```text
/home/manuel/code/wesen/2026-03-27--hetzner-k3s/vault/policies/github-actions/<app>-gitops-pr.hcl
```

Replace the old PAT read path:

```hcl
path "kv/data/ci/github/<app>/gitops-pr-token" {
  capabilities = ["read"]
}
```

with the GitHub App credential path:

```hcl
path "kv/data/ci/github/<app>/gitops-pr-app" {
  capabilities = ["read"]
}
```

Keep the token self-management stanzas:

```hcl
path "auth/token/lookup-self" { capabilities = ["read"] }
path "auth/token/renew-self" { capabilities = ["update"] }
path "auth/token/revoke-self" { capabilities = ["update"] }
```

Apply the policy live:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
vault policy write gha-<app>-gitops-pr vault/policies/github-actions/<app>-gitops-pr.hcl
vault policy read gha-<app>-gitops-pr | rg 'gitops-pr-app|gitops-pr-token|capabilities'
```

Commit and push the policy change in the GitOps repo. Preserve unrelated local `ttmp/` changes.

## Step 4: update the app workflow

In the app repo, replace the old token-source inputs:

```yaml
gitops_pr_token_source: vault
vault_role: <app>-gitops-pr
vault_secret_path: kv/data/ci/github/<app>/gitops-pr-token
```

with:

```yaml
gitops_pr_token_source: github_app
vault_role: <app>-gitops-pr
gitops_app_secret_path: kv/data/ci/github/<app>/gitops-pr-app
gitops_app_owner: wesen
gitops_app_repositories: 2026-03-27--hetzner-k3s
```

The caller workflow must still include:

```yaml
permissions:
  contents: read
  packages: write
  pull-requests: write
  id-token: write
```

Do not set `vault_secret_path` for `github_app` mode. The reusable workflow reads `gitops_app_secret_path` and exports a minted installation token as `GITOPS_PR_TOKEN` for the `open-gitops-pr` action.

## Step 5: update app deployment docs

If the app repo has deploy docs, replace language that says “GitOps PR token from Vault” with “GitHub App credentials from Vault; short-lived installation token minted by the reusable workflow.”

Mention the exact credential path:

```text
kv/data/ci/github/<app>/gitops-pr-app
```

and the GitOps installation target:

```text
owner: wesen
repository: 2026-03-27--hetzner-k3s
```

## Step 6: validate locally

Validate workflow YAML:

```bash
python3 - <<'PY'
import yaml
from pathlib import Path
yaml.safe_load(Path('.github/workflows/publish-image.yaml').read_text())
print('workflow yaml ok')
PY
```

Verify the App can mint an installation token and write to the GitOps repo. If the `publish-vault` ticket scripts are available, reuse the write verifier:

```bash
export VAULT_ADDR=https://vault.yolo.scapegoat.dev
export VAULT_TOKEN="$(<"$HOME/.vault-token")"
export GITOPS_APP_SECRET_PATH=kv/ci/github/<app>/gitops-pr-app
export GITOPS_OWNER=wesen
export GITOPS_REPO=2026-03-27--hetzner-k3s

/home/manuel/code/wesen/go-go-golems/publish-vault/ttmp/2026/05/31/RETRO-GITOPS-008--automate-gitops-pr-credentials-with-github-app-tokens/scripts/06-verify-github-app-gitops-write-access.sh
```

Expected output:

```text
Installation token minted OK: expires_at=...
Remote branch push OK: verify/github-app-token-...
Remote branch cleanup OK: verify/github-app-token-...
```

If this fails, fix Vault policy, Vault secret shape, or GitHub App installation before opening the app PR.

## Step 7: open the app PR

Commit the app workflow/doc changes on a branch and open a PR.

The PR body should include:

- workflow changed from `vault` PAT mode to `github_app` mode,
- Vault secret path used,
- live Vault policy was updated,
- App installation-token write verification passed,
- any dependency/security bump needed to satisfy pre-push or CI.

## Step 8: merge and verify the first main run

After merging the app PR to `main`, watch the source repo workflow run:

```bash
gh run list --workflow publish-image.yaml --limit 5
gh run watch <run-id>
```

The important log sequence is:

```text
Read GitHub App credentials from Vault
Mint GitHub App token for GitOps repository
Open GitOps pull requests for published image
```

A successful run should open or update a PR in:

```text
https://github.com/wesen/2026-03-27--hetzner-k3s
```

Merge that GitOps PR after review, then verify Argo CD and rollout health for the app.

## Common failure modes

### `Invalid username or token` while cloning the GitOps repo

The action received an invalid or expired token. Confirm the workflow is no longer in old PAT mode and that `GITOPS_PR_TOKEN` comes from `actions/create-github-app-token`.

### Vault can log in but cannot read the app secret

The Vault policy still points at `gitops-pr-token`, or it has not been applied live. Run:

```bash
vault policy read gha-<app>-gitops-pr
```

### `actions/create-github-app-token` cannot find installation

The GitHub App is not installed on `wesen/2026-03-27--hetzner-k3s`, or `gitops_app_owner` / `gitops_app_repositories` is wrong.

### The workflow does not get a Vault token

Check the Vault role bound claims. The role should match the exact source repo, `refs/heads/main`, and `push` event for main-branch deploy runs.

## Migration checklist

- [ ] Existing workflow uses `publish-ghcr-image.yml@main` or another ref that supports `github_app` mode.
- [ ] GitHub App installed on `wesen/2026-03-27--hetzner-k3s`.
- [ ] Vault secret `kv/ci/github/<app>/gitops-pr-app` exists with `app_id` and `private_key`.
- [ ] Vault policy `gha-<app>-gitops-pr` reads only the app credential path plus token self-management paths.
- [ ] Policy applied live with `vault policy write`.
- [ ] Source workflow uses `gitops_pr_token_source: github_app`.
- [ ] Source workflow passes `gitops_app_secret_path`, `gitops_app_owner`, and `gitops_app_repositories`.
- [ ] Workflow YAML parses.
- [ ] GitHub App write verifier can push and delete a temporary branch.
- [ ] App PR opened and merged.
- [ ] First `main` workflow opens a GitOps PR successfully.
