# Tasks

## Phase 1: Research and design

- [x] Create the INFRA-0001 ticket workspace
- [x] Inspect the existing `infra-tooling` repository charter, docs, scripts, templates, and examples
- [x] Compare the extracted template/script model against concrete source repositories using local copies
- [x] Review official GitHub Actions docs for reusable workflows, composite actions, Docker actions, and action metadata
- [x] Write the intern-facing design and implementation guide
- [x] Record the chronological investigation diary

## Phase 2: Productize the reusable workflow surface

- [x] Add `.github/workflows/publish-ghcr-image.yml` to `infra-tooling`
- [x] Use `on: workflow_call` and define explicit inputs/secrets for source repos
- [x] Decide the minimal caller contract:
  - build context
  - Dockerfile path
  - test command
  - image name override
  - target config path
  - whether to open GitOps PRs
- [x] Define stable outputs such as the published immutable image ref
- [ ] Document the caller example for Go repos and for non-Go repos

## Phase 3: Productize the GitOps PR helper as an action

- [x] Add `actions/open-gitops-pr/action.yml`
- [x] Add a Dockerfile or equivalent packaged runner for the action
- [x] Move the canonical PR helper implementation behind the action entrypoint
- [x] Define inputs for config path, image ref, target selection, push/open-pr flags, author identity, and dry-run mode
- [x] Define action outputs such as branch name, PR number, and whether a manifest changed
- [ ] Decide whether the action should keep the current line-oriented YAML patcher or switch to a YAML-aware implementation

## Phase 4: Tighten contracts and examples

- [x] Add a JSON schema or equivalent validation for `deploy/gitops-targets.json`
- [x] Add an example caller workflow that uses the reusable workflow instead of copying a template
- [x] Add an example repo-local `deploy/gitops-targets.json`
- [x] Update platform docs to describe the versioned reuse model instead of the copy-local-script model
- [x] Document secret expectations for `GITHUB_TOKEN` and `GITOPS_PR_TOKEN`

## Phase 5: Testing and verification

- [x] Add unit tests for target loading and manifest patch logic
- [x] Add fixture tests for single-target and multi-target updates
- [x] Add a dry-run integration test against a temporary Git repository
- [x] Add a self-test workflow inside `infra-tooling` that exercises the action and reusable workflow
- [ ] Verify behavior for already-updated manifests and already-open PRs

## Phase 6: Migration of existing app repos

- [x] Identify the first pilot repo to migrate to the reusable workflow
- [x] Replace its copied workflow/script path with a call into `infra-tooling`
- [ ] Verify the GitOps PR path still works end to end
  Local dry-run validation now passes against the real `smailnail` target and real `wesen/2026-03-27--hetzner-k3s` checkout. Live `smailnail` run `23910489369` also completed the publish phase successfully after the Dockerfile Go-version fix, but the GitOps handoff still skipped because `GITOPS_PR_TOKEN` is not configured for the repo.
- [ ] Migrate the remaining image-based repos incrementally
- [ ] Remove unnecessary copied local scripts once the shared action is trusted
- [ ] Push the `smailnail` pilot branches and observe a live GitHub Actions run plus GitOps PR creation
  A live `publish-image` run now exists and succeeds through image build/push. The remaining missing piece is enabling `GITOPS_PR_TOKEN` so the second job opens a real PR instead of intentionally skipping.
- [ ] Apply `gitops/applications/smailnail.yaml` in-cluster once Vault runtime secrets and the Keycloak `k3s-parallel` environment exist

## Phase 7: Delivery and follow-up

- [x] Validate this ticket bundle with `docmgr doctor`
- [x] Upload the final bundle to reMarkable
- [ ] Decide the release/tagging strategy for `infra-tooling` reusable assets
- [ ] Decide whether GitHub Marketplace publication is useful or unnecessary for the internal platform
