---
Title: Investigation diary
Ticket: INFRA-007
Status: active
Topics:
    - github
    - release
    - automation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: abs:///home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf
      Note: Inspected role and policy evidence
    - Path: repo://.github/workflows/publish-ghcr-image.yml
      Note: Inspected App credential and token-minting evidence
ExternalSources: []
Summary: Chronological evidence and decisions for INFRA-007.
LastUpdated: 2026-07-17T17:21:34.47023215-04:00
WhatFor: Enable a future engineer to retrace the release credential design investigation.
WhenToUse: Read before implementing, reviewing, or resuming INFRA-007.
---


# Investigation diary

## Goal

Record the evidence, commands, decisions, and residual questions behind the
Vault-backed release credential migration design.

## Step 1: Map current release and Vault credential boundaries

This step created INFRA-007 and established the design baseline. The work did
not read or modify any raw credential value. It inspected configuration and
Terraform declarations to determine which credentials are actually used, which
existing Vault/OIDC patterns can be reused, and where the proposed system needs
new infrastructure.

### Prompt Context

**User prompt (verbatim):** "ok, can you make a docmgr ticket in ~/code/wesen/go-go-golems/infra-tooling and Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new infra-tooling ticket, research the
current release/GitOps/Vault mechanisms, write an intern-ready design and
implementation plan for release credential migration, and upload the finished
bundle to reMarkable.

**Inferred user intent:** Make release credential access centralized,
least-privilege, auditable, and easier to operate across Go-Go-Golems projects.

**Commit (code):** N/A — design ticket creation and documentation only.

### What I did

- Created ticket `INFRA-007` with the `github`, `release`, and `automation`
  vocabulary topics.
- Created the primary design document and this diary using docmgr.
- Inspected `publish-ghcr-image.yml`'s three token-source paths and its GitHub
  App token mint sequence.
- Inspected `publish-docsctl.yml` and the active Vault Terraform environment
  for the package-scoped OIDC role pattern.
- Compared Pinocchio and go-minitrace release workflows plus GoReleaser files.
- Inspected the GitHub App GitOps migration playbook, which includes an
  isolated write/delete verifier.

### Why

The release credential design should reuse established platform patterns where
they match the security problem. The goal is not merely relocating a PAT from
GitHub to Vault. It is issuing short-lived, repository-scoped authority for
GitHub writes and using exact OIDC claim binding for unavoidable static vendor
secrets.

### What worked

- `docmgr status --summary-only` confirmed infra-tooling already has a
  structured ticket workspace and vocabulary.
- `docmgr ticket create-ticket --ticket INFRA-007 ...` created the expected
  ticket with `index.md`, `tasks.md`, and `changelog.md`.
- Existing reusable workflows provided concrete API and policy conventions;
  no speculative new authentication system is required.
- The evidence shows Pinocchio and go-minitrace use GitHub Actions secrets for
  GoReleaser, GPG, Homebrew, and Fury, while docs use Vault/OIDC.

### What didn't work

- Initial command: `docmgr ticket list --with-glaze-output json`
- Result: `Too many arguments`.
- Resolution: used `docmgr ticket list` directly. The command's help states
  that structured output requires separate `--with-glaze-output` and output
  selection flags; it does not accept a positional `json` value.

### What I learned

- Terraform's `docsctl_publishers` entries already bind immutable repository
  IDs, release tags, `push`, exact `workflow_ref`, and
  `job_workflow_ref`, with a five-minute Vault token. This is the appropriate
  model for release publisher roles.
- `publish-ghcr-image.yml` obtains App credentials from Vault, then uses
  `actions/create-github-app-token@v2` and exports only the minted token. Its
  structure directly supports a Homebrew tap publisher adaptation.
- The existing workflows pass `COSIGN_PWD`, but inspected GoReleaser configs do
  not configure Cosign. The ticket treats it as inventory/cleanup work rather
  than assuming it should be migrated.

### What was tricky to build

The main design challenge is authorization granularity. A Vault role that only
binds a caller repository and workflow cannot necessarily distinguish a split
build job from the high-privilege merge job in the same workflow. The proposed
solution uses an infra-tooling reusable publish workflow so a
`job_workflow_ref` claim can form a separately testable boundary. This must be
verified against live non-production OIDC claims before policy rollout.

### What warrants a second pair of eyes

- Validate the exact OIDC claim set when a caller invokes the proposed reusable
  workflow; do not deploy a policy based solely on documentation workflow
  claims.
- Confirm GitHub App authorization is sufficient for all current GoReleaser
  Homebrew formula and cask modes.
- Review the first GPG-in-Vault migration with the signing-key owner because
  the key still enters an ephemeral runner until keyless signing is adopted.

### What should be done in the future

- Execute Phase 0 inventory and OIDC/App proof work before any secret copy or
  caller workflow change.
- Implement the Terraform and reusable workflow phases in this ticket.
- Create or link a successor ticket for keyless signing after consumer
  verification requirements are known.

### Code review instructions

- Start with the primary design document's “Current-State Architecture and
  Evidence” and “Target Architecture” sections.
- Review the referenced workflow steps in `publish-ghcr-image.yml` and
  `publish-docsctl.yml` against the Terraform bindings in `main.tf`.
- Validate the proposed role with a test OIDC workflow and inspect denied as
  well as allowed cases.

### Technical details

Commands used:

```text
docmgr status --summary-only
docmgr vocab list --category topics
docmgr ticket create-ticket --ticket INFRA-007 ...
rg -n -i "vault-action|create-github-app-token|goreleaser|HOMEBREW|FURY" ...
sed -n '180,330p' .github/workflows/publish-ghcr-image.yml
sed -n '230,430p' terraform/vault/github-actions/envs/k3s/main.tf
```

No secret-bearing Vault read command was run. All evidence came from checked-in
workflow, GoReleaser, Terraform, and playbook files.

## Step 2: Validate and publish the design bundle

The completed design package was validated as a docmgr ticket and rendered as a
single reMarkable PDF bundle. The upload contains the ticket index, the primary
design, this diary, the long-term task checklist, and changelog, so an intern
can read the architectural rationale and immediately locate the implementation
work that remains.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Deliver the ticket materials to reMarkable after
the design has been written and validated.

**Inferred user intent:** Make the design available for offline review without
losing the ticket's operational task and diary context.

**Commit (code):** pending — documentation-only ticket commit follows this
delivery step.

### What I did

- Ran `docmgr doctor --ticket INFRA-007 --stale-after 30`.
- Ran `remarquee upload bundle --dry-run` with the five ticket documents.
- Uploaded the same bundle to `/ai/2026/07/17/INFRA-007`.

### Why

The doctor check catches malformed metadata and broken relationships before a
document is circulated. The dry-run verifies the exact PDF input set and remote
path before an upload. A bundle keeps the design and its execution checklist
together rather than scattering multiple PDFs on the device.

### What worked

- Doctor result: `All checks passed` for INFRA-007.
- Dry-run included `index.md`, primary design, diary, `tasks.md`, and
  `changelog.md`.
- Upload result: `OK: uploaded INFRA-007 Vault Release Credentials Design.pdf
  -> /ai/2026/07/17/INFRA-007`.

### What didn't work

N/A.

### What I learned

The final bundle is self-contained enough for offline review: it opens with
the ticket purpose, explains current evidence and target design, then provides
implementation phases and diary evidence.

### What was tricky to build

The bundle has a long primary document and several supporting files. Including
the task list and changelog is important: without them, a reader can understand
the target but cannot see which individual rollout actions remain pending. The
upload command's ToC depth of two preserves navigability without generating an
overly deep outline.

### What warrants a second pair of eyes

- The document's proposal to use reusable-workflow `job_workflow_ref` as a
  credential boundary needs a real non-production OIDC claim test before
  Terraform implementation.
- Secret-path names and profile mapping should be reviewed by the Vault and
  release-key owners before secrets are seeded.

### What should be done in the future

- Start Phase 0 tasks in `tasks.md`; do not migrate or delete any secret before
  the inventory and App/OIDC proofs are complete.

### Code review instructions

- Verify the reMarkable package corresponds to the current files in INFRA-007.
- Use `docmgr doctor --ticket INFRA-007 --stale-after 30` after any later
  document update, then upload a replacement bundle only when deliberate.

### Technical details

```text
docmgr doctor --ticket INFRA-007 --stale-after 30
remarquee upload bundle --dry-run ... --remote-dir /ai/2026/07/17/INFRA-007
remarquee upload bundle ... --remote-dir /ai/2026/07/17/INFRA-007
```

## Step 3: Re-scope the first pilot and implement its Vault boundary

### Prompt Context

**User prompt (verbatim):** “alright, add tasks if you haven't already, and
then implement step by step, cmomit at appropriate itnervals, keeping a
detailed diary as you go. Start with ~/code/wesen/go-go-golems/sqleton as a
first repository, until it works, and then we can advise how to continue once
that works. terraform is in ~/code/wesen/terraform”

**Assistant interpretation:** Treat Sqleton as the first concrete adopter,
track the work at task granularity, implement only the infrastructure that can
be safely validated without copying secrets or changing live Vault state, and
commit that boundary independently.

**Inferred user intent:** Prove the design against an operational release
workflow before advising a wider migration. The proof must use the real
Terraform repository, preserve auditability, and avoid inventing replacement
credentials for values GitHub cannot reveal.

**Commit (Terraform):** `a56efca feat(vault): add Sqleton release publisher
roles` in the dedicated Terraform feature worktree.

### What I did

- Created clean feature worktrees for infra-tooling, Sqleton, and Terraform;
  the active Terraform checkout had unrelated work and was left untouched.
- Captured the official GoReleaser split/merge documentation using Defuddle at
  `sources/01-goreleaser-split-and-merge.md`.
- Inspected Sqleton's actual release configuration. It builds Linux and macOS
  with CGO and publishes a GitHub release, GHCR images, a Homebrew formula, and
  Fury packages. Its current workflow imports GPG and Cosign-related secrets,
  but its committed `.goreleaser.yaml` defines neither a `signs` stanza nor a
  Cosign publisher.
- Reordered the rollout plan so Sqleton is Phase 4, ahead of tiny-idp. The
  task checklist now separately records local implementation, reviewed Vault
  apply, bootstrap, and production-release evidence.
- Added Terraform `release_credential_profiles` and `release_publishers` for
  Sqleton. The only initial profile is `homebrew-fury`, with fixed policies for
  the GoReleaser Pro license, Homebrew publisher App credential, and Fury
  token.
- Added separate build and publisher Vault JWT roles. The build role reads
  only the GoReleaser Pro license. The publisher role is bound to the calling
  repository, immutable repository ID, `v*` tag push, caller workflow path,
  and the shared reusable workflow's `job_workflow_ref`.
- Added a non-secret `release_publishers` Terraform output for review and
  inventory tooling.
- Ran `terraform fmt` and a backend-free `terraform validate` after downloading
  only the pinned providers into the temporary worktree. No `terraform plan`
  or `terraform apply` was run, and no live Vault state was changed.

### Why

GoReleaser's documented split/merge flow is a Pro-only feature. Sqleton's
existing two-platform workflow cannot safely be converted to the shared merge
design without supplying the Pro license to each split build. Keeping the
license in a build-only role limits a compromised build job to the capability
it needs; it cannot update Homebrew or publish to Fury. The merge role receives
the remaining static credentials only when the caller is the approved
reusable workflow.

Sqleton is a good first consumer because it exercises several publication
channels but avoids an otherwise confounding GPG migration. Retiring unused
GPG/Cosign wiring is safer than carrying it forward merely because a secret
exists.

### What worked

- Defuddle successfully saved the official GoReleaser source after the ticket
  `sources/` directory was created.
- `terraform fmt` normalized the new HCL.
- `terraform -chdir=vault/github-actions/envs/k3s init -backend=false` installed
  the lockfile-pinned Vault and Kubernetes providers in the temporary
  worktree.
- `terraform -chdir=vault/github-actions/envs/k3s validate` reported
  `Success! The configuration is valid.`
- The Terraform implementation was committed as `a56efca` without staging
  unrelated files.

### What didn't work

The first Defuddle command failed because the newly created ticket did not yet
have a `sources/` directory:

```text
Error: ENOENT: no such file or directory, open '.../sources/01-goreleaser-split-and-merge.md'
```

Creating the standard folder and repeating the identical capture succeeded.

The first validation attempt stopped at formatting drift, as designed:

```text
vault/github-actions/envs/k3s/main.tf
vault/github-actions/envs/k3s/outputs.tf
```

After formatting, validation initially could not use the copied worktree's
absent provider cache:

```text
Error: missing or corrupted provider plugins:
  - registry.terraform.io/hashicorp/kubernetes: there is no package ... cached
  - registry.terraform.io/hashicorp/vault: there is no package ... cached
```

`terraform init -backend=false` resolved that local, expected condition. It
did not use the S3 backend or contact Vault.

### What I learned

- The previous docsctl publisher entry for Sqleton names `release.yaml`, but
  the real release file is `release.yml`. The new release role uses the actual
  path, so claim binding is not inherited from a stale documentation entry.
- The standard reusable workflow reference is expected in the OIDC
  `job_workflow_ref` claim at `@refs/heads/main`, even though callers write
  `@main`; existing docsctl roles demonstrate this convention.
- GitHub Actions secret values cannot be inventoried or copied back out of
  GitHub. A later bootstrap phase must obtain current values from their owners,
  place them directly in the approved Vault paths, and verify only metadata
  such as key names and fingerprint.

### What was tricky to build

The security boundary is intentionally split in two dimensions. A broad
repository-tag role would be simpler, but then every platform build could
read credentials that update an external Homebrew repository and vendor
registry. Conversely, binding only the shared workflow would make it
impossible for platform builds to read the Pro license. The two-role design
preserves the smallest useful authority:

```text
Sqleton v* tag build job
  -> release-sqleton-builder
  -> GoReleaser Pro license only

Sqleton v* tag merge job calling approved reusable workflow
  -> release-sqleton-publisher
  -> license + Homebrew App credentials + Fury token
```

The role does not grant list, create, or update on KV paths, and it does not
grant an arbitrary input-selected path. The profile-to-path map is Terraform
data, reviewed with the authorization policy.

### What warrants a second pair of eyes

- Confirm that the chosen Homebrew GitHub App is installed only for
  `go-go-golems/homebrew-go-go-go` and has the minimum contents permission to
  update the formula.
- Confirm the intended Vault secret owners and rotation cadence before paths
  are seeded. The source repository has no `kv/ci/release` entries yet.
- Review whether GoReleaser Pro licensing permits the planned CI use and
  whether its exact action distribution/version should be pinned before the
  first tag.

### What should be done in the future

- Implement the Sqleton caller workflow and validate it locally with
  GoReleaser configuration checks.
- Add a workflow contract test and a policy-render negative test, then run a
  non-production OIDC claim diagnostic.
- Seed secrets and apply Terraform only after owner approval. A real tag and
  secret deletion are explicitly later tasks, not part of this implementation
  commit.

### Code review instructions

- Review Terraform `release_build` before `release_publish`; verify the former
  cannot read the Homebrew or Fury paths.
- Check every `bound_claims` entry against the exact Sqleton workflow path and
  the reusable workflow's checked-out branch ref.
- Check that all secret paths use KV v2 `kv/data/...` policy paths and that the
  output includes no secret values.
- Re-run the backend-free initialization and validation commands below in a
  clean worktree. Do not apply without reviewing the remote-state plan.

### Technical details

```text
defuddle parse https://goreleaser.com/customization/general/partial/ --md -o \
  ttmp/.../sources/01-goreleaser-split-and-merge.md
terraform fmt vault/github-actions/envs/k3s/main.tf \
  vault/github-actions/envs/k3s/outputs.tf
terraform -chdir=vault/github-actions/envs/k3s init -backend=false
terraform -chdir=vault/github-actions/envs/k3s validate
git commit -m "feat(vault): add Sqleton release publisher roles"
```

## Step 4: Convert Sqleton to least-privilege split and merge publication

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Implement the first repository adoption rather
than stopping at platform plumbing, retain a testable release workflow, and
record any production-relevant limitation instead of masking it with a broad
configuration rewrite.

**Inferred user intent:** The first pilot should be a realistic release path:
the platform builds should be able to create artifacts and Linux container
images, while only a final merge job may create the GitHub release or mutate
external publishers.

**Commit (Sqleton):** `3b14bc5 ci(release): split Sqleton build and Vault
publish` in the dedicated Sqleton feature worktree.

### What I did

- Replaced Sqleton's old sequential Linux-then-macOS release process with two
  independent GoReleaser Pro split jobs and a reusable merge/publish job.
- Restricted the release trigger from every tag to `v*`, matching the Vault
  role's exact tag claim.
- Applied job-level permissions:
  - Linux build: `contents: read`, `packages: write`, `id-token: write`.
  - macOS build: `contents: read`, `id-token: write`.
  - shared publication call: `contents: write`, `id-token: write`.
- Changed the Linux GHCR login to the repository-scoped `github.token` and
  removed use of the long-lived `RELEASE_ACTION_PAT`.
- Had both build jobs authenticate to Vault as `release-sqleton-builder` and
  retrieve only `license_key` for GoReleaser Pro.
- Uploaded independent `sqleton-dist-linux` and `sqleton-dist-darwin`
  artifacts. The shared publisher consumes those names, validates inputs,
  reloads the merged credentials through `release-sqleton-publisher`, mints a
  Homebrew App installation token, and runs `goreleaser continue --merge`.
- Removed workflow references to `RELEASE_ACTION_PAT`,
  `GO_GO_GOLEMS_SIGN_KEY`, `GO_GO_GOLEMS_SIGN_PASSPHRASE`, `COSIGN_PWD`, and
  the direct GitHub `FURY_TOKEN` secret.
- Ran YAML parsing, GoReleaser configuration validation, a static retired
  secret scan, and `go test ./...`.

### Why

The old workflow granted `contents: write` globally and gave both platform
jobs all publisher-related GitHub secrets, despite its GoReleaser
configuration not using GPG or Cosign. The new structure communicates the
operational split directly in the workflow graph:

```text
v* push
  ├── Linux split build (Pro license + GHCR package write)
  ├── macOS split build (Pro license only)
  └── reusable merge publication (Vault publisher role)
          ├── GitHub Release via caller GITHUB_TOKEN
          ├── Homebrew formula via short-lived App token
          └── Fury packages via Vault vendor token
```

This matches GoReleaser's documented split/merge model: split jobs produce
platform artifacts and the merge command performs final release publication.

### What worked

- Ruby's YAML parser reported `YAML syntax OK` for the new workflow.
- `go test ./...` passed after downloading normal Go module dependencies.
- The installed `goreleaser-pro` (v2.13.3) accepted the configuration under
  `goreleaser check --soft --config .goreleaser.yaml` and reported `1
  configuration file(s) validated`.
- The static scan found no retired release-secret references in
  `.github/workflows/release.yml`.
- The caller workflow was committed as `3b14bc5` independently of Terraform
  and infra-tooling changes.

### What didn't work

Strict `goreleaser check --config .goreleaser.yaml` exited nonzero even though
the configuration parsed, because the committed config uses v2-deprecated
properties:

```text
DEPRECATED: snapshot.name_template should not be used anymore
DEPRECATED: archives.format should not be used anymore
DEPRECATED: archives.builds should not be used anymore
dockers and docker_manifests are being phased out ... dockers_v2
brews is being phased out in favor of homebrew_casks
```

`goreleaser check --soft` is explicitly a syntax-only validation mode and
passed. This is not treated as an ignored production error. A new Phase 4 task
requires a reviewed deprecation resolution or explicit temporary approval
before a real production tag. The official deprecation reference is captured
as `sources/02-goreleaser-v2-deprecations.md`.

### What I learned

- The local development environment has GoReleaser Pro installed, so it can
  validate the existing v2 configuration without fabricating a license.
- Existing Pinocchio and go-minitrace workflows are useful current examples:
  they use `goreleaser-action@v7`, `actions/upload-artifact@v7`, and
  `actions/download-artifact@v8`. The new implementation follows those
  project-local conventions rather than retaining Sqleton's older action
  versions.
- Artifact download layouts can vary between direct and directory-wrapped
  uploads. The shared workflow accepts both `dist-parts/<platform>/` and
  `dist-parts/<platform>/dist/` layouts before calling GoReleaser merge.

### What was tricky to build

The credential roles must correspond to the actual execution graph. It would
be incorrect to give each split job the same Vault role as the reusable
publisher merely because both invoke GoReleaser. The split commands need the
license; they do not need to update the Homebrew tap or call Fury. Conversely,
the merge command must still receive the caller repository's GitHub token for
the GitHub Release, but no long-lived token for the caller repository is
stored in Vault.

Another subtle constraint is trigger alignment. The Vault policy accepts only
`refs/tags/v*`; keeping the old workflow's `'*'` trigger would result in
unexplained Vault authentication failures for tags outside that convention.

### What warrants a second pair of eyes

- Verify GitHub's effective permissions for the nested reusable workflow when
  `contents: write` is passed by the caller; the real run should confirm the
  GitHub Release can be created.
- Verify that the repository `GITHUB_TOKEN` has package-write access for GHCR
  in Sqleton's Actions settings.
- Review the GoReleaser deprecation migration separately. Replacing Docker or
  Homebrew configuration changes release artifacts and should not be bundled
  into a credential-boundary change without focused tests.

### What should be done in the future

- Commit the small action-major-version correction in the shared infra-tooling
  workflow, then run the same YAML validation there.
- Perform the Phase 3 bootstrap steps and a real, non-production end-to-end
  run only after the Homebrew App and Vault values exist.
- Do not delete old GitHub secrets until the approved production tag has
  demonstrated GitHub Release, GHCR, Homebrew, and Fury behavior.

### Code review instructions

- Compare `release-linux` and `release-darwin` environment blocks: neither
  should contain a tap, Fury, GPG, or Cosign credential.
- Verify that the only job with `contents: write` is `publish`, which calls
  the shared workflow after both artifact producers complete.
- Verify that `GGOOS` is `linux` or `darwin` in the respective split jobs and
  artifact names match the shared-workflow inputs exactly.
- Re-run the commands below and inspect the deprecation output rather than
  assuming a zero exit from `--soft` means all production quality gates pass.

### Technical details

```text
ruby -e 'require "yaml"; YAML.load_file(ARGV.fetch(0))' .github/workflows/release.yml
goreleaser check --config .goreleaser.yaml       # reports existing deprecations
goreleaser check --soft --config .goreleaser.yaml
go test ./...
rg -n 'secrets\\.(RELEASE_ACTION_PAT|GO_GO_GOLEMS_SIGN_KEY|GO_GO_GOLEMS_SIGN_PASSPHRASE|COSIGN_PWD|FURY_TOKEN)' .github/workflows/release.yml
git commit -m "ci(release): split Sqleton build and Vault publish"
```

## Step 5: Add a cross-repository release contract harness

### Prompt Context

**User prompt (verbatim):** (same as Step 3)

**Assistant interpretation:** Make the first pilot repeatedly reviewable rather
than relying on a single manual reading of three repositories. The harness must
be safe to run locally and must not read Vault secrets.

**Inferred user intent:** A later intern should have a concrete command that
detects drift between the Sqleton caller workflow, the shared workflow, and
Terraform authorization before a release reaches GitHub Actions.

**Commit (documentation/harness):** pending this diary entry's commit.

### What I did

- Added
  `scripts/check_sqleton_release_contract.sh` to INFRA-007.
- Made the script accept explicit infra-tooling, Terraform, and Sqleton roots,
  so it can test clean feature worktrees without assuming a user's directory
  layout.
- Checked caller invariants: `v*` trigger, split `GGOOS` values, builder and
  publisher role names, shared-workflow reference, and absence of retired
  GitHub secret references.
- Checked shared-workflow invariants: approved `homebrew-fury` profile,
  fixed Homebrew destination, fixed Vault credential fields, App-token mint,
  and GoReleaser merge command.
- Checked Terraform invariants: the builder policy reads the Pro license but
  not Homebrew/Fury paths; the publisher policy derives paths from the
  allowlisted profile map; and its role binds `job_workflow_ref`.
- Ran the script against the three dedicated worktrees. It ended with
  `Sqleton release contract: PASS`.
- Related the script and both newly captured official sources to the ticket
  index and re-ran `docmgr doctor` successfully.

### Why

The change is distributed across three repositories. Each file can look
reasonable in isolation while the release fails because an artifact name,
Vault role, tag pattern, or shared workflow reference differs. The harness
turns the security architecture into a small set of executable assertions.
Most importantly, it includes a negative assertion: a split builder must not
gain publisher credential paths.

### What worked

The final command completed without reading a secret or contacting a remote
service:

```text
bash .../check_sqleton_release_contract.sh \
  /tmp/infra-tooling-release-vault-sqleton \
  /tmp/terraform-release-vault-sqleton \
  /tmp/sqleton-release-vault-sqleton
Sqleton release contract: PASS
```

`docmgr doctor --ticket INFRA-007 --stale-after 30` also returned `All checks
passed` after linking the sources and harness.

### What didn't work

The harness's first version expected the fully expanded publisher paths inside
the rendered `vault_policy.release_publish` HCL block. Terraform correctly
derives those paths through
`local.release_credential_profiles[each.value.profile].secret_paths`, so the
first run reported:

```text
contract violation: publisher policy is missing kv/data/ci/release/shared/goreleaser-pro
```

The check was corrected to assert both parts of the real representation: the
policy must reference the allowlisted profile map, and that map must contain
each required path. The second run passed. This was a harness expectation bug,
not a policy defect.

### What I learned

Terraform's declarative indirection is part of the authorization design. A
static test should verify the indirection rather than flattening it into a
different mental model. The map is what prevents a workflow input from
selecting an arbitrary Vault path, and the policy's reference to the map is
what makes the map authoritative.

### What was tricky to build

The script deliberately uses only file inspection. Running `terraform plan`
would additionally require remote state, Vault credentials, and Kubernetes
configuration; running a GitHub workflow would require currently absent
credential paths and could publish real artifacts. The harness instead covers
the contractual gaps that must be correct before either external action is
safe.

### What warrants a second pair of eyes

- The harness detects intended current contracts, not every possible GitHub
  Actions semantic error. A test release remains necessary after Vault
  bootstrap.
- As additional credential profiles are introduced, extend the script or
  replace its profile checks with a typed parser so every profile is covered.

### What should be done in the future

- Add this script to a non-release CI workflow after the three feature
  branches are integrated or expose it through an infra-tooling command.
- Add an OIDC claim diagnostic that proves Vault's actual
  `job_workflow_ref` observations match the policy's expected value.

### Code review instructions

- Read the negative builder-policy assertion first; it guards the central
  least-privilege property.
- Verify every string asserted by the script matches the intended public
  workflow contract, not merely the current implementation accidentally.
- Run the command with a modified, deliberately unsafe fixture (for example,
  add `FURY_TOKEN` to a build job) to confirm the failure is understandable.

### Technical details

```text
bash ttmp/2026/07/17/INFRA-007--.../scripts/check_sqleton_release_contract.sh \
  /tmp/infra-tooling-release-vault-sqleton \
  /tmp/terraform-release-vault-sqleton \
  /tmp/sqleton-release-vault-sqleton
docmgr doc relate --ticket INFRA-007 --file-note ...
docmgr doctor --ticket INFRA-007 --stale-after 30
```

## Quick Reference

<!-- Provide copy/paste-ready content, API contracts, or quick-look tables -->

## Usage Examples

<!-- Show how to use this reference in practice -->

## Related

<!-- Link to related documents or resources -->
