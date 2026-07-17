---
Title: Sqleton Release Credential Bootstrap Runbook
Ticket: INFRA-007
Status: ""
Topics:
    - github
    - release
    - automation
DocType: reference
Intent: ""
Owners: []
RelatedFiles:
    - Path: abs:///home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf
      Note: Defines the temporary Sqleton bootstrap policy and exact OIDC claim binding.
    - Path: repo://ttmp/2026/07/17/INFRA-007--migrate-release-credentials-to-vault-and-github-app-publishing/scripts/03-store-homebrew-publisher-app.sh
      Note: Operator-only direct Vault storage helper for the new App key.
ExternalSources: []
Summary: Operator procedure for migrating existing release credentials to Vault and registering the narrow Homebrew publisher GitHub App.
LastUpdated: 0001-01-01T00:00:00Z
WhatFor: ""
WhenToUse: ""
---



# Sqleton Release Credential Bootstrap Runbook

This runbook performs the one-time transition from GitHub Actions secrets to
Vault-backed release credentials for Sqleton. It is intentionally divided into
two trust paths. Existing GoReleaser and Fury values move inside an audited
GitHub Actions runner using GitHub OIDC. The Homebrew publisher credential is
a new GitHub App key and moves from the GitHub App registration download
directly into Vault through an operator terminal.

Neither procedure prints a secret value. Do not paste a private key, token, or
license key into a ticket, PR, shell history, issue, chat, or workflow input.

## Preconditions

- Infra-tooling's reusable publisher workflow and Terraform release roles are
  merged to `main`.
- The follow-up Terraform bootstrap-role change and Sqleton bootstrap workflow
  are merged to `main` and applied to Vault.
- An operator can authenticate to the K3s Vault instance with the expected
  OIDC identity and has a clean local checkout of infra-tooling.
- The Sqleton repository has current `GORELEASER_KEY` and `FURY_TOKEN` Actions
  secrets available to a same-repository workflow. The manual workflow checks
  this before making any Vault write.
- The operator has organization permission to register and install a GitHub
  App. The normal `repo` API token is insufficient for this web-owner action.

## 1. Register and install the Homebrew publisher App

Use
`scripts/02-homebrew-publisher-app-manifest.json` as the exact manifest. In
the `go-go-golems` organization GitHub App registration screen:

1. Register a private/internal App named
   `go-go-golems-homebrew-release-publisher` (choose a unique suffix only if
   GitHub reports a collision).
2. Set the homepage URL to the infra-tooling repository.
3. Disable webhooks. The publisher receives no events.
4. Grant only repository **Contents: Read and write**. No administration,
   Actions, issues, pull requests, checks, packages, or organization
   permissions are required.
5. Install the App on the `go-go-golems` organization with **Only select
   repositories**, selecting exactly `homebrew-go-go-go`.
6. Generate a private key once. Record the numeric **App ID** (not client ID)
   and retain the downloaded PEM only long enough to store it in Vault.

GitHub installation tokens are short-lived and cannot gain access to a
repository or permission that the App installation did not grant. The shared
workflow additionally requests a token only for `homebrew-go-go-go`.

## 2. Store the App credential directly in Vault

From the infra-tooling checkout, run:

```bash
bash ttmp/2026/07/17/INFRA-007--migrate-release-credentials-to-vault-and-github-app-publishing/scripts/03-store-homebrew-publisher-app.sh \
  APP_ID /secure/path/to/downloaded-private-key.pem
```

Expected non-secret output resembles:

```json
{"version":1,"keys":["app_id","private_key"]}
```

The script validates that the App ID is numeric and the PEM parses as a private
key. It writes only
`kv/ci/github/homebrew-go-go-go/release-publisher-app` with keys `app_id` and
`private_key`. Remove the downloaded private-key file using the organization’s
approved secure-storage procedure after verifying the Vault metadata.

## 3. Migrate existing GitHub release secrets inside GitHub Actions

In `go-go-golems/sqleton`, open **Actions → bootstrap-release-credentials →
Run workflow**, choose the `main` branch, and type exactly:

```text
MIGRATE_GITHUB_SECRETS_TO_VAULT
```

The workflow requires the release environment, obtains a five-minute Vault
token through the `release-sqleton-bootstrap` role, and may write only:

- `kv/ci/release/shared/goreleaser-pro` / `license_key`
- `kv/ci/release/shared/fury-go-go-golems` / `token`

It never receives the Homebrew App private key. Its policy has no read,
delete, list, or access to unrelated paths. Save the completed GitHub Actions
run URL in the ticket diary as non-secret evidence.

If the workflow fails its availability check, do not replace the missing value
with a new unrelated credential. Identify the source system that owns the
license or Fury token, then have an authorized operator write it to the same
approved Vault key name.

## 4. Verify metadata and App capability

Run the ticket’s existing non-secret contract harness, then use the App token
verifier from the Phase 0 task to create and remove a uniquely named temporary
branch in `homebrew-go-go-go`. Verify only the run result, target repository,
branch name, token expiry, and App installation permission; do not emit the
token.

```bash
bash ttmp/2026/07/17/INFRA-007--migrate-release-credentials-to-vault-and-github-app-publishing/scripts/check_sqleton_release_contract.sh \
  /path/to/infra-tooling /path/to/terraform /path/to/sqleton
```

## 5. Remove the bootstrap path

After successful migration and verification:

1. Disable the `bootstrap-release-credentials` workflow or remove it in a
   dedicated follow-up change.
2. Remove the `release-sqleton-bootstrap` Terraform policy and role in the
   same change and apply it.
3. Record Vault metadata versions and GitHub Actions run URL, not secret
   values.
4. Only after a successful controlled release, remove the legacy GitHub
   secrets and record that removal separately.

## Failure handling

- **Workflow cannot authenticate to Vault:** inspect non-secret OIDC claim and
  role binding values. Do not weaken the role to wildcard repository or ref
  claims.
- **App token cannot write the tap:** confirm the installation selects only
  `homebrew-go-go-go` and has Contents write. Do not broaden it to all
  organization repositories.
- **Vault write failed after one path succeeded:** rerun the manual workflow
  after reviewing the run. The writes are deterministic key replacements; log
  the run identifiers and resulting Vault metadata version.
- **Private key is exposed or lost:** revoke the key in GitHub App settings,
  generate a replacement, update the same Vault key, then record the rotation.

## References

- [GitHub App installation authentication](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/authenticating-as-a-github-app-installation)
- [GitHub App key management](https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/managing-private-keys-for-github-apps)
- [Vault policies](https://developer.hashicorp.com/vault/docs/concepts/policies)
