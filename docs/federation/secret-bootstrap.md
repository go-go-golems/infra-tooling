# Secret Bootstrap For Federated Remotes

Every source repo using the federated remote release pattern needs a standard set of GitHub secrets and variables.

This bootstrap has now been exercised successfully for:

- `go-go-app-inventory`
- `go-go-app-sqlite`

## Required Secrets

- `HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID`
- `HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY`
- `HETZNER_OBJECT_STORAGE_BUCKET`
- `HETZNER_OBJECT_STORAGE_ENDPOINT`
- `HETZNER_OBJECT_STORAGE_REGION`
- `GITOPS_PR_TOKEN`

## Required Variables

- remote public base URL variable
- platform package version variable, if the repo consumes published `@go-go-golems/os-*`

## Reusable Bootstrap Script

The Terraform-backed object-storage bootstrap is now available as:

- `scripts/federation/bootstrap_federation_source_repo_from_terraform.sh`

Usage:

```bash
scripts/federation/bootstrap_federation_source_repo_from_terraform.sh \
  <owner/repo> \
  <REMOTE_PUBLIC_BASE_URL_VAR> \
  [platform-version] \
  [bucket-name]
```

Example for sqlite:

```bash
scripts/federation/bootstrap_federation_source_repo_from_terraform.sh \
  go-go-golems/go-go-app-sqlite \
  SQLITE_FEDERATION_PUBLIC_BASE_URL \
  0.1.0-canary.5
```

This script sets:

- `HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID`
- `HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY`
- `HETZNER_OBJECT_STORAGE_BUCKET`
- `HETZNER_OBJECT_STORAGE_ENDPOINT`
- `HETZNER_OBJECT_STORAGE_REGION`
- `<REMOTE_PUBLIC_BASE_URL_VAR>`
- `GO_GO_OS_PLATFORM_VERSION` when a platform version argument is provided

It does not provision:

- `GITOPS_PR_TOKEN`
- `K3S_REPO_READ_TOKEN`

Those still need to be created separately because they do not come from the Terraform object-storage environment.

## Generic Bootstrap Commands

```bash
gh secret set HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID --repo <owner>/<repo>
gh secret set HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY --repo <owner>/<repo>
gh secret set HETZNER_OBJECT_STORAGE_BUCKET --repo <owner>/<repo>
gh secret set HETZNER_OBJECT_STORAGE_ENDPOINT --repo <owner>/<repo>
gh secret set HETZNER_OBJECT_STORAGE_REGION --repo <owner>/<repo>
gh secret set GITOPS_PR_TOKEN --repo <owner>/<repo>

gh variable set <REMOTE_PUBLIC_BASE_URL_VAR> --repo <owner>/<repo> --body "https://<bucket>.<region>.your-objectstorage.com"
gh variable set GO_GO_OS_PLATFORM_VERSION --repo <owner>/<repo> --body "<platform-version>"
```

## Recommendation

Standardize the remote public base URL variable name over time so workflows do not carry app-specific env names forever.
