# Secret Bootstrap For Federated Remotes

Every source repo using the federated remote release pattern needs a standard set of GitHub secrets and variables.

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
