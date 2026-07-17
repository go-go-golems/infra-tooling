---
Title: Vault-backed release credentials and GitHub App publishing design
Ticket: INFRA-007
Status: active
Topics:
    - github
    - release
    - automation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: abs:///home/manuel/code/wesen/go-go-golems/go-minitrace/.github/workflows/release.yaml
      Note: Second current GitHub-secret release merge consumer
    - Path: abs:///home/manuel/code/wesen/go-go-golems/pinocchio/.github/workflows/release.yml
      Note: Current GitHub-secret release merge consumer
    - Path: abs:///home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf
      Note: Existing exact-claim Vault JWT role and policy conventions
    - Path: abs:///home/manuel/workspaces/2026-07-07/prod-tiny-idp/tiny-idp/.github/workflows/release.yml
      Note: Pilot candidate release workflow
    - Path: repo://.github/workflows/publish-docsctl.yml
      Note: Existing GitHub OIDC and short-lived publish-token flow
    - Path: repo://.github/workflows/publish-ghcr-image.yml
      Note: Existing Vault-to-GitHub-App installation-token implementation
    - Path: repo://docs/go-go-golems/playbooks/github-app-gitops-pr-migration-playbook.md
      Note: Operator migration and token-write verification precedent
ExternalSources: []
Summary: An evidence-based design for moving GoReleaser-era release credentials from GitHub secrets into narrowly scoped Vault/OIDC roles, replacing the Homebrew tap PAT with GitHub App installation tokens, and preparing a later keyless-signing migration.
LastUpdated: 2026-07-17T17:21:34.381925143-04:00
WhatFor: Designing and implementing shared release credential issuance for go-go-golems repositories.
WhenToUse: Use before changing Terraform, Vault, shared workflows, or a repository release workflow that currently consumes long-lived release secrets.
---


# Vault-backed release credentials and GitHub App publishing design

## Executive Summary

The current Go-Go-Golems artifact-release workflows publish a GitHub release,
sign checksums with GPG, update a Homebrew tap, and sometimes upload packages to
Fury. Representative repositories—Pinocchio and go-minitrace—receive the
GoReleaser Pro license, GPG private key and passphrase, Homebrew tap token, and
Fury token directly from GitHub Actions secrets. tiny-idp was intentionally
aligned with that existing model during its release-readiness work.

The platform already demonstrates a stronger model in two adjacent systems.
The reusable `publish-docsctl.yml` workflow authenticates to Vault using a
GitHub Actions OIDC token and receives a short-lived, package-scoped publishing
JWT. The reusable `publish-ghcr-image.yml` workflow can read a GitHub App
credential from Vault and mint a short-lived installation token for its GitOps
repository. This ticket adapts those proven patterns to binary artifact
releases.

The target state has two credential classes. Static vendor secrets, where no
short-lived protocol exists, are stored once in Vault and released only to a
tag- and workflow-bound Vault role. GitHub repository-write authority is not
stored as a personal access token: Vault stores a GitHub App ID and private key,
and the workflow mints an installation token limited to the Homebrew tap. The
first implementation preserves GPG checksum signing so it is a migration rather
than a release-format redesign. A later phase replaces exportable GPG material
with keyless Sigstore/Cosign signing where practical.

The design deliberately treats “copy the secrets to Vault” as an intermediate
operation, not the completed security outcome. A successful migration verifies
the Vault-backed release path, revokes/removes the corresponding GitHub secret,
records the rotation owner, and proves that the tag workflow cannot read secrets
outside its declared release profile.

## Problem Statement

### The release credential problem

A release workflow needs several unrelated capabilities. The default
`GITHUB_TOKEN` creates a GitHub release in the caller repository. GoReleaser
Pro requires a vendor license. GPG checksum signing requires private key
material and a passphrase. Homebrew publishing needs write access to a separate
repository. Fury publishing uses a vendor API token. Treating all of these as
one generic “release secret” makes access reviews, rotation, incident response,
and least-privilege policy unnecessarily difficult.

Today, representative workflows put those capabilities in their GitHub secret
namespace. Pinocchio's merge job imports
`GO_GO_GOLEMS_SIGN_KEY` and `GO_GO_GOLEMS_SIGN_PASSPHRASE`, then passes
`GORELEASER_KEY`, `TAP_GITHUB_TOKEN`, and `FURY_TOKEN` to GoReleaser
(`~/code/wesen/go-go-golems/pinocchio/.github/workflows/release.yml`).
go-minitrace follows the same pattern in
`~/code/wesen/go-go-golems/go-minitrace/.github/workflows/release.yaml`.
The GoReleaser configurations consume `TAP_GITHUB_TOKEN` for the
`go-go-golems/homebrew-go-go-go` tap and `FURY_TOKEN` for a curl publisher.

This is operationally workable but has four weaknesses:

- GitHub repository or organization secret access is the primary authorization
  boundary, instead of an explicit policy that expresses repository, tag,
  workflow, and credential capability.
- The Homebrew token is a long-lived bearer credential with write authority to
  a separate repository. It has a wider lifetime and usually a wider blast
  radius than an installation token.
- Credential inventory, rotation metadata, and audit review are split across
  GitHub and operator knowledge rather than being centralized in Vault.
- A copied secret can silently remain active in GitHub after a migration,
  leaving two sources of truth and defeating a clean access revocation story.

### Scope

This design covers release credentials used by GitHub Actions tag workflows in
the `go-go-golems` organization. It defines shared infrastructure, Terraform
contracts, migration mechanics, tests, and repository adoption. It uses
Pinocchio, go-minitrace, Glazed, and tiny-idp as representative consumers.

In scope:

- Vault KV paths, JWT roles, policies, and Terraform representation.
- A shared workflow/action interface for preparing release credentials in the
  job that runs GoReleaser merge/publish.
- Homebrew tap publishing with a GitHub App installation token.
- Safe migration of GoReleaser license, GPG, and Fury credentials to Vault.
- Release evidence, rollback, auditing, and credential-removal procedure.
- A deliberate follow-up toward keyless Cosign/Sigstore signing.

Out of scope for the first delivery:

- Changing project versioning, artifact formats, package names, or the
  GoReleaser split/merge build topology.
- Replacing Fury as a package registry.
- Building a generic secret-management framework for runtime application
  secrets.
- Automatically reading, printing, or copying secret values in CI logs or this
  ticket.

### Terms

**Caller repository** is the project being released, such as tiny-idp.
**Release merge job** is the job that combines platform artifacts and carries
out side effects: GitHub release creation, signing, tap updates, and package
publication. **Vault JWT role** is a Vault authentication role that validates
GitHub's OIDC assertion. **Credential profile** is the explicit, named set of
release capabilities allowed for one repository and workflow. **GitHub App
installation token** is a short-lived token minted for a single installed App
and target repository; it replaces the Homebrew personal access token.

## Current-State Architecture and Evidence

### Release execution topology

Both studied projects use three jobs. Linux and macOS jobs run GoReleaser with
`release --split`, upload `dist` as artifacts, and a merge job downloads and
combines the artifacts before invoking `goreleaser continue --merge`. This is
not an arbitrary convention: CGO builds need native or explicitly configured
cross compilation, while release-side effects should occur only once after all
platform artifacts are available.

```text
tag push
  |
  +--> Linux split build --------> dist-linux artifact --+
  |                                                    |
  +--> macOS split build --------> dist-darwin artifact +--> merge/publish job
                                                               |
                                                               +--> GitHub release
                                                               +--> checksums + GPG signature
                                                               +--> Homebrew tap update
                                                               +--> optional Fury upload
                                                               +--> docs publisher
```

Pinocchio's release workflow shows this directly: the first two jobs have only
the GoReleaser key plus GitHub token, while the merge job adds signing, tap, and
Fury credentials. go-minitrace adds a separately built frontend but keeps the
same privilege boundary. It also places its merge job in the GitHub `release`
environment; this is a useful human-approval layer but is not a replacement for
Vault policy.

### Current GoReleaser credential contract

The observed GoReleaser contract is intentionally small:

```yaml
checksum:
  name_template: checksums.txt

signs:
  - artifacts: checksum
    args: ["--batch", "-u", "{{ .Env.GPG_FINGERPRINT }}", ...]

brews:
  - repository:
      owner: go-go-golems
      name: homebrew-go-go-go
      token: "{{ .Env.TAP_GITHUB_TOKEN }}"

publishers:
  - name: fury.io
    cmd: curl -F package=@{{ .ArtifactName }} https://{{ .Env.FURY_TOKEN }}@push.fury.io/go-go-golems/
```

The project workflow must therefore prepare environment variables before
GoReleaser runs. It does not need to expose secrets as reusable-workflow
outputs, commit them to configuration, or pass them through artifacts.

One observed discrepancy matters for cleanup: Pinocchio and go-minitrace pass
`COSIGN_PWD` into their merge job, but their inspected GoReleaser files contain
no Cosign signing/publisher configuration. The migration must inventory actual
consumers before moving a secret; an unused secret should be removed rather
than re-homed.

### Existing Vault/OIDC model for documentation

`infra-tooling/.github/workflows/publish-docsctl.yml` performs
three distinct operations:

1. GitHub Actions authenticates to Vault using `hashicorp/vault-action` with
   `method: jwt`.
2. The Vault token can call `identity/oidc/token/<role>` to mint a short-lived
   docs-registry JWT.
3. `docsctl publish` receives the minted token from a temporary file, while the
   workflow logs only selected non-sensitive JWT claims.

Terraform models those permissions in
`~/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf`.
The `docsctl_publishers` map carries repository, immutable repository ID, and
exact workflow reference. The resulting JWT role binds the owner, repository,
repository ID, tag reference, `push` event, and workflow reference. Its TTL is
five minutes and policy has only the mint-token capability plus token
self-management endpoints. This is the strongest available local precedent for
least-privilege release publishing.

### Existing GitHub App model for GitOps

`infra-tooling/.github/workflows/publish-ghcr-image.yml`
already supports `gitops_pr_token_source: github_app`. In that mode it:

```text
GitHub Actions OIDC assertion
  -> hashicorp/vault-action reads app_id and private_key from one KV path
  -> actions/create-github-app-token@v2 mints an installation token
  -> token is masked and exported as GITOPS_PR_TOKEN
  -> open-gitops-pr writes only to the selected GitOps repository
```

The migration playbook at
`infra-tooling/docs/go-go-golems/playbooks/github-app-gitops-pr-migration-playbook.md`
documents the exact App requirements: Contents read/write, Pull requests
read/write where PR creation is needed, and Metadata read-only. A Homebrew
publisher only needs to commit a generated formula/cask directly; Contents
read/write and Metadata read-only are the baseline permissions to validate.

### Constraints inferred from current implementation

- Vault policy alone cannot distinguish two ordinary jobs in the same caller
  workflow if their OIDC claims are otherwise identical. Passing a more
  privileged role name from an untrusted build job would not create a security
  boundary.
- A shared reusable workflow has a distinct `job_workflow_ref` claim. It can
  make the publish step an independently attestable boundary while receiving
  build artifacts through GitHub Actions artifacts.
- The release merge job is the only job that requires Homebrew, GPG, or Fury.
  It is therefore the correct place to retrieve those secrets.
- The GoReleaser Pro license may be required in split-build jobs. It has much
  lower authority than a repository-writing token, but it must still be
  stored and scoped deliberately if the organization chooses a complete Vault
  migration.

## Target Architecture

### Security boundary map

```text
              ┌─────────────────────────────────────────────┐
              │ caller repository: tag vX.Y.Z                │
              │ Linux / macOS split build jobs               │
              │ - GitHub contents: read                      │
              │ - optional Vault build-license role          │
              └───────────────┬─────────────────────────────┘
                              │ upload immutable dist artifacts
                              v
 ┌──────────────────────────────────────────────────────────────────────┐
 │ infra-tooling reusable publish workflow                               │
 │ job_workflow_ref = infra-tooling release-publish workflow             │
 │ permissions: contents: write, id-token: write                         │
 │                                                                      │
 │ GitHub OIDC -> Vault release-publish role                             │
 │   ├─ read GPG secret (temporary migration only)                       │
 │   ├─ read Fury token only for credential profile that needs it         │
 │   └─ read Homebrew App id/private key                                  │
 │                 -> GitHub App installation token for tap              │
 │                                                                      │
 │ merge artifacts -> import GPG -> goreleaser continue --merge          │
 └─────────────────────────────────────┬────────────────────────────────┘
                                       │
              ┌────────────────────────┼───────────────────────────┐
              v                        v                           v
      caller GitHub release     homebrew-go-go-go             Fury registry
      built-in GITHUB_TOKEN     App installation token        scoped vendor token
```

### Release credential classes

| Capability | First-delivery source | Effective authority | Planned end state |
|---|---|---|---|
| GitHub release in caller repository | GitHub-provided `GITHUB_TOKEN` | caller repo contents/releases | retain |
| GoReleaser Pro license | Vault KV | vendor-license use | Vault KV or license-free release architecture |
| GPG checksum signing | Vault KV | exports private signing material into ephemeral runner | replace with keyless signing |
| Homebrew tap publishing | Vault GitHub App credential | short-lived token for tap repo | retain |
| Fury publishing | Vault KV | package upload to organization namespace | retain until vendor supports federation |
| Docs publishing | existing Vault Identity OIDC JWT | package/version publication | retain |

The table distinguishes a Vault-stored App private key from a generated
installation token. The App credential remains sensitive and is guarded by
Vault; the job never receives a long-lived tap PAT. The generated token is
masked, short-lived, and restricted by the App installation to the tap.

### Repository profile data model

Add a Terraform local map named `release_publishers`, parallel to
`docsctl_publishers`. Each entry is declarative authorization data, not secret
material.

```hcl
locals {
  release_publishers = {
    tinyidp = {
      repository       = "go-go-golems/tiny-idp"
      repository_id    = "<immutable GitHub repository id>"
      workflow_ref     = "go-go-golems/tiny-idp/.github/workflows/release.yml@refs/tags/v*"
      publish_workflow = "go-go-golems/infra-tooling/.github/workflows/publish-goreleaser-release.yml@refs/tags/v*"
      profile          = "gpg-homebrew"
      needs_fury       = false
    }
    pinocchio = {
      repository       = "go-go-golems/pinocchio"
      repository_id    = "802670903"
      workflow_ref     = "go-go-golems/pinocchio/.github/workflows/release.yml@refs/tags/v*"
      publish_workflow = "go-go-golems/infra-tooling/.github/workflows/publish-goreleaser-release.yml@refs/tags/v*"
      profile          = "gpg-homebrew-fury"
      needs_fury       = true
    }
  }
}
```

`profile` is intentionally an allowlisted enum implemented by the shared
workflow, not an arbitrary string interpolated into a KV path. It makes review
of which releases may receive Fury or signing material straightforward.

### Vault secret layout

Paths below show intent; create them only through the approved operator/Vault
bootstrap procedure. The ticket must never contain raw values.

```text
kv/ci/release/shared/goreleaser-pro
  license_key

kv/ci/release/shared/gpg-checksum-signing
  private_key
  passphrase
  fingerprint                 # non-secret; supports deterministic import

kv/ci/github/homebrew-go-go-go/release-publisher-app
  app_id
  private_key

kv/ci/release/shared/fury-go-go-golems
  token
```

Using a shared path does not create broad access: Terraform creates a separate
policy per caller release publisher and grants only paths compatible with its
profile. A future repository that only creates GitHub releases receives none of
the above paths.

### Shared workflow API

The preferred first implementation is a reusable workflow rather than emitting
secrets as outputs. Secrets stay inside the side-effecting job, while caller
repositories pass only declarative inputs and artifact names.

```yaml
jobs:
  publish:
    needs: [goreleaser-linux, goreleaser-darwin]
    permissions:
      contents: write
      id-token: write
    uses: go-go-golems/infra-tooling/.github/workflows/publish-goreleaser-release.yml@<pinned-ref>
    with:
      dist_artifacts: dist-linux,dist-darwin
      goreleaser_config: .goreleaser.yaml
      credential_profile: gpg-homebrew
      vault_role: release-tinyidp-publisher
      homebrew_owner: go-go-golems
      homebrew_repository: homebrew-go-go-go
```

The reusable workflow validates inputs before requesting any Vault credential.
It downloads only named artifacts, merges the expected `dist` directories,
retrieves the profile's credentials with `hashicorp/vault-action`, mints the
Homebrew App token, imports the GPG key if selected, and invokes
`goreleaser continue --merge`. The action writes masked variables to the
environment only after validation.

#### Pseudocode: publish workflow

```text
validate(input.profile in {gpg-homebrew, gpg-homebrew-fury, homebrew})
validate(input.dist_artifacts has exactly expected safe artifact names)
validate(input.homebrew owner/repository for profiles needing Homebrew)

checkout caller repository at immutable tag revision
download each input dist artifact into dist-parts/
merge dist-parts/*/dist into dist/

vault_login_oidc(role=input.vault_role, audience=platform_audience)
credentials = vault_read_allowlisted_profile(profile)
mask(credentials.static_values)

if profile needs homebrew:
    app_token = create_github_app_token(
        app_id=credentials.homebrew_app_id,
        private_key=credentials.homebrew_app_private_key,
        owner=input.homebrew_owner,
        repositories=[input.homebrew_repository])
    set_masked_environment("TAP_GITHUB_TOKEN", app_token)

if profile needs gpg:
    import_gpg(credentials.gpg_private_key, credentials.gpg_passphrase,
               credentials.gpg_fingerprint)
    set_environment("GPG_FINGERPRINT", imported_fingerprint)

set_masked_environment("GORELEASER_KEY", credentials.goreleaser_license)
if profile needs fury: set_masked_environment("FURY_TOKEN", credentials.fury_token)
goreleaser_continue_merge(config=input.goreleaser_config)
```

### Terraform authorization contract

For each `release_publishers` entry, Terraform creates:

1. A policy with read capability only for paths selected by the profile, plus
   token lookup/renew/revoke self-management endpoints.
2. A JWT role that requires the exact organization, repository, immutable
   repository ID, tag `refs/tags/v*`, `push` event, caller workflow reference,
   and shared publish workflow reference.
3. A five-minute token TTL, bounded maximum TTL, and a small token-use budget.

```hcl
resource "vault_jwt_auth_backend_role" "release_publish" {
  for_each = local.release_publishers

  backend   = vault_jwt_auth_backend.github_actions.path
  role_name = "release-${each.key}-publisher"
  role_type = "jwt"

  bound_audiences   = [var.github_actions_audience]
  bound_claims_type = "glob"
  bound_claims = {
    repository_owner = "go-go-golems"
    repository       = each.value.repository
    repository_id    = each.value.repository_id
    ref_type         = "tag"
    ref              = "refs/tags/v*"
    event_name       = "push"
    workflow_ref     = each.value.workflow_ref
    job_workflow_ref = each.value.publish_workflow
  }

  token_policies = [vault_policy.release_publish[each.key].name]
  token_ttl      = 300
  token_num_uses = 8
}
```

The exact claim availability must be verified with a non-production diagnostic
workflow before applying the policy. The docs workflow proves that
`workflow_ref` and `job_workflow_ref` are useful claims in this platform, but a
release reusable-workflow call must be tested rather than assumed identical.

## Decision Records

### Decision: use Vault as the sole release-secret authority

- **Context:** GitHub release secrets are currently the operational source for
  GPG, Homebrew, Fury, and GoReleaser credentials.
- **Options considered:** Keep GitHub secrets; keep copies in both systems;
  move static values to Vault and remove GitHub copies after verification.
- **Decision:** Vault becomes the source of truth; GitHub copies are removed
  per credential after a successful production-like test.
- **Rationale:** Two active copies make revocation and access review ambiguous.
  Vault policies and audit logs provide an explicit, centralized control plane.
- **Consequences:** Vault availability becomes a release dependency. The
  release runbook needs a tested rollback path and an operator owns rotation.
- **Status:** proposed

### Decision: use a GitHub App, not a tap PAT

- **Context:** Homebrew publishing writes to a repository separate from the
  caller repository.
- **Options considered:** Long-lived PAT in GitHub; PAT stored in Vault; Vault
  App credential followed by a GitHub App installation token.
- **Decision:** Store the App credential in Vault and mint an installation token
  for the tap during each publish job.
- **Rationale:** The existing GitOps workflow proves the platform mechanism.
  The installation token is short lived and repository-scoped.
- **Consequences:** The App must be installed on the tap and its Contents
  permission must be tested. GoReleaser must receive the minted token as
  `TAP_GITHUB_TOKEN`.
- **Status:** proposed

### Decision: make release publishing a reusable workflow boundary

- **Context:** The current merge job is the only place that needs powerful
  credentials, but a single caller workflow does not reliably distinguish build
  and publish jobs in Vault claims.
- **Options considered:** Give all jobs one role; use a caller-job composite
  action; use a reusable publish workflow with a distinct job workflow ref.
- **Decision:** Publish side effects run in an infra-tooling reusable workflow;
  a composite action may be used internally by that workflow.
- **Rationale:** It keeps credential values in the side-effecting job and gives
  Terraform a more precise policy anchor.
- **Consequences:** Artifact and configuration inputs become a supported API.
  The reusable workflow must be versioned, tested, and compatible with current
  GoReleaser v2 configurations.
- **Status:** proposed

### Decision: preserve GPG first, plan keyless signing separately

- **Context:** Existing releases publish GPG signatures for checksums, while
  some workflows also pass an apparently unused `COSIGN_PWD`.
- **Options considered:** Keep GitHub GPG secret; migrate GPG into Vault;
  remove GPG immediately in favor of Cosign; run both signature systems.
- **Decision:** Migrate GPG into Vault in the first release-credential change,
  inventory/remove unused Cosign secret wiring, and create a separately tested
  keyless signing rollout.
- **Rationale:** It avoids changing artifact verification semantics at the same
  time as credential transport. Keyless signing is desirable but deserves its
  own compatibility and consumer-validation work.
- **Consequences:** The first workflow still imports an exportable private key
  into a short-lived runner. This remains a tracked residual risk.
- **Status:** proposed

## Implementation Plan

### Phase 0 — Inventory, ownership, and non-production proof

**Goal:** establish facts before moving any credential.

1. Inventory every release workflow and GoReleaser configuration. Record which
   variables are actually consumed, not merely passed through job environment.
2. Name an owner and rotation cadence for GPG, GoReleaser Pro, Homebrew App,
   Fury, and any Cosign material.
3. Confirm whether Homebrew Casks and formulas accept a GitHub App
   installation token in a temporary tap branch.
4. Add a non-production OIDC diagnostic workflow that prints only approved
   claim names/values and confirms `workflow_ref` plus `job_workflow_ref`.
5. Decide whether GoReleaser Pro license retrieval is required in split jobs
   or can be limited to the reusable publish job.

**Exit criteria:** reviewed inventory; no secret values in source control or
logs; known claim contract; App installation and minimum permissions confirmed.

### Phase 1 — Terraform release-publisher primitive

**Goal:** make a reviewable, least-privilege policy family.

1. Add `release_publishers` and profile validation to
   `terraform/vault/github-actions/envs/k3s/main.tf`.
2. Implement profile-to-path mapping in Terraform without user-provided string
   interpolation into policy paths.
3. Create `vault_policy.release_publish` and
   `vault_jwt_auth_backend_role.release_publish` resources.
4. Bind repository ID, tag ref, event, caller workflow ref, and reusable
   publish workflow ref.
5. Set short token TTLs, bounded maximum TTLs, and minimal token uses.
6. Add Terraform tests/format/plan evidence and review the rendered policy.

**Exit criteria:** plan proves each pilot repository can read only profile
paths, and unrelated repositories cannot authenticate as its publish role.

### Phase 2 — Shared reusable publish workflow

**Goal:** add a stable, testable interface in infra-tooling.

1. Create `.github/workflows/publish-goreleaser-release.yml` with
   `workflow_call` input schema.
2. Validate artifact names, profile enum, config path, and Homebrew target
   before Vault login.
3. Download/merge split artifacts using current Linux/Darwin conventions.
4. Reuse the GitOps workflow pattern for Vault JWT login and GitHub App token
   minting, but export the result as `TAP_GITHUB_TOKEN` only in this job.
5. Retrieve static secret fields from explicit, profile-controlled Vault paths;
   immediately add GitHub masks and never expose them as action outputs.
6. Import GPG only for a profile that requires it; invoke GoReleaser merge.
7. Write unit/YAML contract tests and an integration test against a test App
   installation or a non-destructive tap branch.

**Exit criteria:** reusable workflow passes validation tests and a synthetic
release can merge artifacts without secrets appearing in logs.

### Phase 3 — Seed and verify Vault secrets

**Goal:** introduce controlled Vault copies before switching callers.

1. Operators create the KV entries using secure local files and avoid terminal
   history or ticket artifacts containing raw secret values.
2. Verify only key names, expected fingerprint, and lengths; never print key
   content or tokens.
3. Configure/install the Homebrew release-publisher App with minimal tap
   permissions.
4. Run a purpose-built verifier that mints a token, creates a uniquely named
   temporary branch in a test target, and cleans it up.
5. Record evidence identifiers, timestamp, and credential version/rotation
   owner in the operations diary.

**Exit criteria:** every profile path exists, its policy is readable only by
the intended role, and App token mint/write/delete has passed.

### Phase 4 — Pilot with tiny-idp

**Goal:** prove the pattern with the newest release workflow before legacy
repositories adopt it.

1. Add `tinyidp` to Terraform with the immutable repository ID and exact
   release workflow path.
2. Change the tiny-idp merge job to call the pinned reusable publish workflow.
3. Leave split jobs unchanged except for any unavoidable license retrieval.
4. Run a release-candidate/snapshot test and inspect artifacts, signature,
   release metadata, and tap update behavior.
5. Create a real tag only after code review, Terraform deployment, and
   environment approval.
6. Remove the corresponding GitHub Actions secrets only after the first
   production tag succeeds and rollback evidence is captured.

**Exit criteria:** release, signed checksums, and Homebrew update succeed using
Vault/App credentials; no retired GitHub secret is needed.

### Phase 5 — Adopt Pinocchio and go-minitrace

**Goal:** migrate the two established release patterns and eliminate stale
credential wiring.

1. Add profiles with Fury for each repository.
2. Align each workflow’s merge job with the shared interface while preserving
   project-specific frontend or SPA preparation.
3. Verify GPG signatures, tap commits, Fury package availability, and docs
   publication on a planned tag.
4. Determine whether the existing `COSIGN_PWD` has an actual consumer. Remove
   it if not; create a dedicated Cosign task if it is intended.
5. Remove GitHub secrets only after each repository independently succeeds.

**Exit criteria:** both releases use Vault-backed profiles; stale Cosign wiring
is removed or tracked by a concrete successor ticket.

### Phase 6 — Operations, observability, and keyless signing follow-up

**Goal:** make the system maintainable after rollout.

1. Add a release credential inventory report that reads Terraform declarations
   and validates workflow profile use without accessing Vault secret values.
2. Document incident response: disable App installation, revoke/rotate vendor
   token, revoke GitHub App private key, and re-run a controlled release.
3. Add a quarterly access review verifying role bindings, App installation
   scope, and GitHub secret absence.
4. Design keyless Cosign signing using GitHub OIDC and publish verification
   guidance for users.

## Testing and Validation Strategy

### Static checks

- `terraform fmt -check` and `terraform validate` for the Vault environment.
- A policy-render test asserting that a `gpg-homebrew` profile cannot read Fury
  and cannot read another repository's credentials.
- YAML parser/linter tests for all reusable workflow input examples.
- A repository scanner that fails if an adopted workflow still references
  `secrets.HOMEBREW_TAP_TOKEN`, `secrets.GO_GO_GOLEMS_SIGN_KEY`, or a retired
  secret name.
- GoReleaser `check` for each project configuration.

### Runtime tests

| Test | Evidence | Must not do |
|---|---|---|
| OIDC claim diagnostic | role accepts only expected tag + workflow | print Vault token |
| negative role test | altered repo/ref/workflow claim is denied | broaden policy to make it pass |
| App mint test | installation token has expected expiry and repository | log token |
| temporary branch write | token writes then deletes a unique verification branch | touch production cask/formula |
| synthetic merge | GoReleaser reads merged artifacts and signs snapshot | create a public release |
| pilot tag | public assets, checksum signature, tap commit, package registry | delete old secret before success |

### Acceptance tests for every repository migration

```text
given a tag-triggered workflow from the exact approved repository and workflow
when the shared publish workflow requests its Vault role
then Vault grants only its declared profile paths
and GoReleaser creates the expected release artifacts
and the Homebrew change is authored through an App installation token
and no GitHub Actions secret is required for migrated credentials
```

## Migration and Rollback Procedure

1. Create secret values in Vault and apply Terraform policy first; do not
   remove GitHub secrets yet.
2. Merge caller and infra-tooling changes through ordinary pull requests.
3. Exercise a non-production verifier and inspect masked logs.
4. Run one controlled production release with both values available but configure
   the workflow to read Vault only.
5. Verify release artifacts and downstream publication.
6. Remove the unused GitHub secret, verify a second controlled release or
   credential preflight, and record the removal.

If the Vault path, policy, or App installation fails before the release is
published, cancel the job and repair the configuration. Do not silently fall
back to a GitHub secret; that turns a Vault outage into an unnoticed security
regression. A documented, operator-approved emergency rollback may temporarily
restore a known GitHub secret, but it must create an incident record and an
expiry task.

## Risks, Alternatives, and Open Questions

### Risks

- **Vault availability:** release publication depends on the Vault control
  plane. Mitigate with preflight checks, short operation windows, and an
  explicit emergency procedure rather than hidden fallbacks.
- **GitHub App scope error:** App installation on the wrong owner/repository or
  missing Contents write fails late. Mitigate with an isolated token verifier.
- **GPG runner exposure:** Vault centralizes the key but the runner still holds
  it briefly. Keyless signing remains the long-term mitigation.
- **Incorrect OIDC claim assumptions:** exact claim semantics differ between
  caller and reusable workflows. Mitigate by capturing diagnostic claims before
  policy enforcement and testing denied cases.
- **Artifact interface drift:** repositories use different GoReleaser
  configurations and frontend preparation. Keep the shared workflow focused on
  post-build artifact merge/publish and preserve project-local build steps.

### Alternatives considered

**Vault-only static Homebrew PAT.** Better inventory and access control than a
GitHub secret, but still a long-lived bearer token. Rejected in favor of the
already-proven GitHub App issuance pattern.

**One broad organization release role.** Easier to bootstrap, but a compromised
repository could use it to publish another repository or the shared tap.
Rejected because repository IDs and exact workflow references are available in
the existing Terraform model.

**Composite action invoked directly in each merge job.** Simpler artifact
wiring, but it does not create an independent `job_workflow_ref` boundary. It
may still be useful internally after the reusable publish workflow boundary is
established.

**Immediate keyless-signing conversion.** Removes GPG exposure but combines
release-distribution and verification UX changes with credential migration.
Deferred to a successor design and rollout.

### Open questions to resolve in Phase 0

1. Does the GitHub OIDC token for a reusable publish workflow expose the exact
   `job_workflow_ref` anticipated by the Terraform role?
2. Does the selected GitHub App installation token work with every current
   GoReleaser Homebrew formula/cask publisher mode?
3. Is `GORELEASER_KEY` required in split build jobs, and can its Vault role be
   safely narrower than the publisher role?
4. Which release consumers verify existing GPG signatures, and what is the
   compatibility plan for a future keyless signature?
5. Is Fury still a required distribution channel for every current profile?

## References

### Primary implementation references

- `infra-tooling/.github/workflows/publish-ghcr-image.yml` — Vault/App token
  minting sequence and input validation.
- `infra-tooling/.github/workflows/publish-docsctl.yml` — GitHub OIDC login,
  short-lived token minting, and non-sensitive claim logging.
- `infra-tooling/docs/go-go-golems/playbooks/github-app-gitops-pr-migration-playbook.md`
  — operational App migration and temporary write verification.
- `terraform/vault/github-actions/envs/k3s/main.tf` — existing docs and GitOps
  role/policy lifecycle, exact-claim binding, and TTL conventions.
- `pinocchio/.github/workflows/release.yml` and `.goreleaser.yaml` — current
  split/merge release and Fury/Homebrew/GPG environment contract.
- `go-minitrace/.github/workflows/release.yaml` and `.goreleaser.yaml` — a
  second existing release consumer with environment protection.
- `tiny-idp/.github/workflows/release.yml` and `.goreleaser.yaml` — the new
  production-oriented pilot candidate.

## Proposed Solution

<!-- Describe the proposed solution in detail -->

## Design Decisions

<!-- Document key design decisions and rationale -->

## Alternatives Considered

<!-- List alternative approaches that were considered and why they were rejected -->

## Implementation Plan

<!-- Outline the steps to implement this design -->

## Open Questions

<!-- List any unresolved questions or concerns -->

## References

<!-- Link to related documents, RFCs, or external resources -->
