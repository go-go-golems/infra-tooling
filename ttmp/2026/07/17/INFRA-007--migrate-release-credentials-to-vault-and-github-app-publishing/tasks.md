# Tasks

## TODO

### Phase 0 — Inventory and proof

- [ ] Inventory all release workflows and actual GoReleaser credential consumers.
- [ ] Assign credential owners, rotation cadence, and incident contacts.
- [ ] Add a non-production reusable-workflow OIDC claim diagnostic.
- [ ] Prove the Homebrew GitHub App can mint a token and write/delete a temporary branch.
- [ ] Decide and document whether split builds require the GoReleaser Pro license.

### Phase 1 — Terraform least privilege

- [ ] Add `release_publishers` model and allowlisted profile-to-path mapping.
- [ ] Implement Vault release-publisher policies with per-profile read paths.
- [ ] Implement tag/caller-workflow/shared-workflow JWT role bindings.
- [ ] Add Terraform validation and negative policy tests.

### Phase 2 — Shared publish workflow

- [ ] Add reusable `publish-goreleaser-release.yml` workflow API.
- [ ] Validate profile, artifact, configuration, and Homebrew target inputs before login.
- [ ] Implement artifact download and split-dist merge.
- [ ] Implement Vault static-secret loading with masking.
- [ ] Implement GitHub App token minting as `TAP_GITHUB_TOKEN`.
- [ ] Implement conditional GPG import and GoReleaser merge execution.
- [ ] Add workflow contract and integration tests.

### Phase 3 — Vault bootstrap and verification

- [ ] Seed approved Vault paths without exposing raw values in logs or tickets.
- [ ] Verify key names/fingerprint/lengths only.
- [ ] Install the Homebrew publisher App with minimum permissions.
- [ ] Run and retain evidence from the token mint/write/delete verifier.

### Phase 4 — tiny-idp pilot

- [ ] Add tiny-idp release-publisher Terraform entry and deploy policy.
- [ ] Move tiny-idp merge publication to the shared workflow.
- [ ] Run a release-candidate/snapshot validation.
- [ ] Complete a reviewed production tag and verify release/signature/tap.
- [ ] Remove migrated tiny-idp GitHub secrets after success.

### Phase 5 — Existing release consumers

- [ ] Add Pinocchio and go-minitrace profiles including Fury.
- [ ] Migrate their merge jobs while preserving project-local build steps.
- [ ] Verify release, GPG, Homebrew, Fury, and docs behavior per repository.
- [ ] Inventory/remove stale `COSIGN_PWD` wiring or create a successor task.
- [ ] Remove migrated GitHub secrets after each independent success.

### Phase 6 — Operations and keyless signing

- [ ] Add a non-secret release credential inventory/preflight report.
- [ ] Publish rotation and incident-response playbook.
- [ ] Schedule quarterly access review.
- [ ] Design and plan keyless Cosign/Sigstore adoption.
