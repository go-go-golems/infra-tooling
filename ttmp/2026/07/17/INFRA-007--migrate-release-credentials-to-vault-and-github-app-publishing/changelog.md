# Changelog

## 2026-07-17

- Initial workspace created


## 2026-07-17

Created evidence-based Vault/OIDC release credential migration design, diary, phases, and validation plan.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/.github/workflows/publish-ghcr-image.yml — GitHub App issuance precedent
- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Vault role/policy precedent


## 2026-07-17

Validated INFRA-007 with docmgr doctor and uploaded the index, design, diary, tasks, and changelog as a reMarkable bundle to /ai/2026/07/17/INFRA-007.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/07/17/INFRA-007--migrate-release-credentials-to-vault-and-github-app-publishing/design-doc/01-vault-backed-release-credentials-and-github-app-publishing-design.md — Published primary design

## 2026-07-17

Implemented the Sqleton-first release credential migration foundation in three
separately committed feature worktrees. Terraform commit `a56efca` adds the
builder/publisher Vault roles and least-privilege credential profile; Sqleton
commit `3b14bc5` uses GoReleaser Pro split builds and the reusable publisher;
infra-tooling commits `38e5ff1`, `cf15ff3`, and `9df29e2` add the shared
workflow, research sources, detailed implementation diary, and cross-repo
contract harness. No Vault policy was applied and no secret was copied.

### Related Files

- .github/workflows/publish-goreleaser-release.yml — shared merge/publication workflow
- scripts/check_sqleton_release_contract.sh — repeatable cross-repository contract check
- sources/01-goreleaser-split-and-merge.md — official split/merge reference
- sources/02-goreleaser-v2-deprecations.md — official configuration deprecation reference

## 2026-07-17

Step 6: added one-time Sqleton GitHub-secret migration path, Homebrew publisher App manifest, direct Vault storage helper, and operator runbook (Terraform f46408b; Sqleton dd6b446).

### Related Files

- /home/manuel/code/wesen/go-go-golems/sqleton/.github/workflows/bootstrap-release-credentials.yml — Manual OIDC bootstrap workflow.
- /home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf — Least-privilege bootstrap role.
