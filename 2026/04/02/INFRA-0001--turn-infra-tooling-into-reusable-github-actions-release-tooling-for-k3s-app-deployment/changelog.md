# Changelog

## 2026-04-02

- Initial workspace created
- Added the primary design and implementation guide for turning `infra-tooling` into a versioned reusable GitHub Actions toolkit
- Added the investigation diary capturing evidence from the repo, reference app repos, and official GitHub Actions docs
- Recorded the main recommendation: keep the work in `infra-tooling`, expose a reusable workflow for orchestration, and expose a Docker-backed action for the GitOps PR helper
- Validated the ticket with `docmgr doctor` after adding the required topic vocabulary entries
- Uploaded the final bundle to reMarkable as `INFRA-0001 infra-tooling reusable GitHub Actions toolkit` under `/ai/2026/04/02/INFRA-0001`
- Implemented the first shared tooling slice: packaged `open-gitops-pr` as a reusable action, moved the canonical helper under the action source tree, added config validation, and added unit tests
- Implemented the second shared tooling slice: added the reusable GHCR publish workflow, added a repo-local self-test workflow, updated the caller template/example, and updated platform docs to the versioned reuse model
- Hardened the helper tests with multi-container fixture coverage, machine-readable output coverage, and a dry-run integration test against a temporary Git repository
- Started the first pilot adoption in `smailnail` by adding a caller workflow that uses the shared publish pipeline with GitOps PR creation intentionally disabled until the K3s target manifests exist
- Completed the local `smailnail` pilot wiring by adding K3s manifests in `wesen/2026-03-27--hetzner-k3s`, adding `deploy/gitops-targets.json` in `smailnail`, enabling the shared GitOps PR path on `main`, and validating the contract with `kubectl kustomize`, the shared config validator, a local dry-run manifest patch, and the `smailnail`/`infra-tooling` test suites
- Fixed a reusable-workflow bug in `publish-ghcr-image.yml` where the GHCR login step referenced `inputs.push-image` instead of `inputs.push_image`
- Fixed the `smailnail` Docker build by aligning its builder image with `go.mod` (`golang:1.26.1-bookworm` instead of `golang:1.25.8-bookworm`)
- Verified the live `smailnail` `publish-image` run `23910489369` now completes the compile, test, and image publish path successfully; the remaining end-to-end blocker is missing `GITOPS_PR_TOKEN`, which causes the GitOps PR job to skip by design
- Completed the first real end-to-end pilot by opening and merging the live `smailnail` GitOps PR, applying the `smailnail` `k3s-parallel` Keycloak env, seeding Vault, bootstrapping Vault Kubernetes auth, and bringing the Argo application to `Healthy Synced`
- Updated the stable playbooks with the live rollout lessons: export `AWS_PROFILE` explicitly for backend-backed Terraform applies, use the in-cluster `keycloak-bootstrap-admin` secret for K3s-side Keycloak admin access, and check `origin/main` before renaming Vault role names locally

## 2026-04-02

Wrote the design bundle for productizing infra-tooling as the shared reusable GitHub Actions and GitOps handoff toolkit for future K3s app repos.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/infra-tooling/design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md — Primary design deliverable
