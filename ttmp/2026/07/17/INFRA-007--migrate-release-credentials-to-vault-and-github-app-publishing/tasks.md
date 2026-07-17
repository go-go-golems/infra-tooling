# Tasks

## TODO

### Phase 0 — Inventory and proof

- [x] Inventory all release workflows and actual GoReleaser credential consumers.
- [ ] Assign credential owners, rotation cadence, and incident contacts.
- [ ] Add a non-production reusable-workflow OIDC claim diagnostic.
- [x] Add a manually confirmed Homebrew App verifier and least-privilege Vault role.
- [ ] Prove the Homebrew GitHub App can mint a token and write/delete a temporary branch.
- [x] Decide and document whether split builds require the GoReleaser Pro license.

### Phase 1 — Terraform least privilege

- [x] Add `release_publishers` model and allowlisted profile-to-path mapping.
- [x] Implement Vault release-publisher policies with per-profile read paths.
- [x] Implement tag/caller-workflow/shared-workflow JWT role bindings.
- [x] Add Terraform validation and negative policy tests.

### Phase 2 — Shared publish workflow

- [x] Add reusable `publish-goreleaser-release.yml` workflow API.
- [x] Validate profile, artifact, configuration, and Homebrew target inputs before login.
- [x] Implement artifact download and split-dist merge.
- [x] Implement Vault static-secret loading with masking.
- [x] Implement GitHub App token minting as `TAP_GITHUB_TOKEN`.
- [x] Implement the Sqleton `homebrew-fury` GoReleaser merge execution.
- [ ] Implement conditional GPG import for the later `gpg-homebrew*` profiles.
- [x] Add workflow contract and integration tests.

### Phase 3 — Vault bootstrap and verification

- [x] Add a one-time, repository-bound GitHub-secret migration workflow and write-only Vault role.
- [x] Store the Homebrew publisher App credential at its approved Vault path and verify non-secret metadata.
- [x] Seed approved Vault paths without exposing raw values in logs or tickets.
- [x] Verify key names/fingerprint/lengths only.
- [x] Install the Homebrew publisher App with minimum permissions.
- [x] Run and retain evidence from the token mint/write/delete verifier.
- [ ] Merge and apply removal of the temporary bootstrap and App-verifier roles/workflows.

### Phase 4 — Sqleton pilot

- [x] Add the Sqleton release-publisher Terraform entry and validate it locally.
- [ ] Apply the reviewed Sqleton release-publisher policy to Vault.
- [x] Move Sqleton to split/merge publication with the shared workflow.
- [ ] Run a release-candidate/snapshot validation.
- [ ] Resolve or explicitly approve Sqleton's GoReleaser v2 configuration deprecations before a production tag.
- [ ] Complete a reviewed production tag and verify release, GHCR, Homebrew, and Fury behavior.
- [ ] Remove migrated Sqleton GitHub secrets after success.

### Phase 5 — tiny-idp pilot

- [ ] Add tiny-idp release-publisher Terraform entry and deploy policy.
- [ ] Move tiny-idp merge publication to the shared workflow.
- [ ] Run a release-candidate/snapshot validation.
- [ ] Complete a reviewed production tag and verify release/signature/tap.
- [ ] Remove migrated tiny-idp GitHub secrets after success.

### Phase 6 — Existing release consumers

- [ ] Add Pinocchio and go-minitrace profiles including Fury.
- [ ] Migrate their merge jobs while preserving project-local build steps.
- [ ] Verify release, GPG, Homebrew, Fury, and docs behavior per repository.
- [ ] Inventory/remove stale `COSIGN_PWD` wiring or create a successor task.
- [ ] Remove migrated GitHub secrets after each independent success.

### Phase 7 — Operations and keyless signing

- [ ] Add a non-secret release credential inventory/preflight report.
- [ ] Publish rotation and incident-response playbook.
- [ ] Schedule quarterly access review.
- [ ] Design and plan keyless Cosign/Sigstore adoption.
