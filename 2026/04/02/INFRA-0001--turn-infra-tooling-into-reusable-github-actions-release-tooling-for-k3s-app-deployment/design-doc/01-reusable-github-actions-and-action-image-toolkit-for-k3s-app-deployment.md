---
Title: Reusable GitHub Actions and action-image toolkit for K3s app deployment
Ticket: INFRA-0001
Status: active
Topics:
    - github-actions
    - ghcr
    - gitops
    - argocd
    - k3s
    - platform
    - release-engineering
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-03-27--mysql-ide/.github/workflows/publish-image.yaml
      Note: |-
        Second app repo showing near-identical workflow shape
        Second copied workflow example in a source repo
    - Path: ../../../../../../../hair-booking/.github/workflows/publish-image.yaml
      Note: |-
        Example app repo using a copied workflow/script pattern
        Copied workflow example in a source repo
    - Path: ../../../../../../../hair-booking/scripts/open_gitops_pr.py
      Note: Example copied helper script in a source repo
    - Path: README.md
      Note: |-
        Repo charter defining what belongs in infra-tooling
        Repo charter for shared tooling ownership
    - Path: docs/platform/source-repo-to-gitops-pr.md
      Note: |-
        Current image-based control-plane model
        Current control-plane model and copy-local-script assumption
    - Path: examples/platform/image-gitops-targets.example.json
      Note: |-
        Current target metadata contract example
        Current target metadata example
    - Path: scripts/gitops/open_gitops_pr.py
      Note: |-
        Current extracted GitOps PR helper
        Current canonical GitOps PR helper
    - Path: templates/github/publish-image-ghcr.template.yml
      Note: |-
        Current reusable-by-copy workflow template
        Current GHCR publish template
ExternalSources:
    - https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
    - https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
    - https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action
    - https://docs.github.com/en/actions/sharing-automations/creating-actions/metadata-syntax-for-github-actions
Summary: Detailed design for converting infra-tooling from a template-and-script repository into a versioned reusable GitHub Actions toolkit for GHCR publish, GitOps PR creation, and K3s app onboarding.
LastUpdated: 2026-04-02T10:27:44.40010878-04:00
WhatFor: Use this document to implement reusable workflows and actions in infra-tooling without rediscovering the platform model or GitHub Actions packaging tradeoffs.
WhenToUse: Read this before building a new shared workflow/action, migrating app repos off copied scripts, or deciding between reusable workflows, composite actions, and Docker actions.
---


# Reusable GitHub Actions and Action-Image Toolkit for K3s App Deployment

## Executive Summary

`infra-tooling` is already the conceptual home for shared release and GitOps automation, but it still delivers the image-based K3s deployment path as copyable assets rather than as versioned reusable GitHub Actions components. The repository has the right ingredients today:

- a control-plane document describing the source repo -> GitOps PR -> Argo CD model
- a workflow template for GHCR publish + GitOps PR handoff
- a generic Python helper that patches a GitOps manifest and opens a PR
- an example `deploy/gitops-targets.json` contract

The problem is packaging, not architecture. Source repositories still carry their own copies of `.github/workflows/publish-image.yaml` and `scripts/open_gitops_pr.py`, which means bug fixes, CLI contract changes, PR title conventions, and manifest patch behavior drift over time.

The recommended design is:

1. Keep `infra-tooling` as the shared home.
2. Add a reusable workflow under `.github/workflows/` for CI orchestration.
3. Add a Docker-backed action under `actions/` for the GitOps PR updater toolchain.
4. Keep per-application build inputs and deployment target metadata local to each source repository.
5. Migrate existing app repositories incrementally from copied assets to versioned references such as `@v1`.

The guiding principle is simple:

```text
shared orchestration belongs in reusable workflows
shared toolchain logic belongs in actions
app-specific release inputs stay in app repos
live desired state stays in the GitOps repo
```

## Problem Statement and Scope

The go-go-golems platform now has multiple application repositories that deploy to the Hetzner K3s cluster by publishing immutable GHCR image tags and then opening GitOps pull requests that update the deployment manifest in a separate infra repository. `infra-tooling` was created to be the neutral home for reusable mechanics that do not belong in a single application repository or in the GitOps repository itself. The repo README states exactly that mission and explicitly lists shared release and GitOps helper scripts, reusable workflow templates, example metadata, and generic source-repo to GitOps PR automation as things that belong here ([README.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md#L5), [README.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md#L13), [README.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md#L64)).

However, the current image-based path still ships as a template-plus-copy model. The platform doc says an app repo should keep `.github/workflows/publish-image.yaml`, `deploy/gitops-targets.json`, and `scripts/open_gitops_pr.py` locally, and it states that this duplication is intentional because the workflow template assumes a local copy of the script ([source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md#L104), [source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md#L113)).

That was a reasonable first extraction step, but it no longer matches the desired operating model. The platform now wants a versioned reusable toolkit that can be called by future K3s app repos without copying the same helper code across repositories.

This document covers:

- the current architecture and why duplication exists
- which responsibilities belong in a reusable workflow
- which responsibilities belong in an action image
- how `deploy/gitops-targets.json` should remain local
- how to expose stable caller contracts for future application repositories
- how to migrate from copied local assets to shared versioned automation

This document does not cover:

- application-specific Dockerfiles
- GitOps manifests for any specific app
- Vault/VSO runtime secret wiring inside the cluster
- direct deployment from source repositories into Kubernetes

## Current-State Analysis

### 1. What `infra-tooling` already is

The repo charter is already aligned with the requested future. The README says `infra-tooling` is the neutral home for reusable mechanics and explicitly names these current reuse points for image-based K3s deployments:

- `docs/platform/source-repo-to-gitops-pr.md`
- `examples/platform/image-gitops-targets.example.json`
- `templates/github/publish-image-ghcr.template.yml`
- `scripts/gitops/open_gitops_pr.py`

That is direct evidence that the platform has already extracted the logic conceptually, even though the packaging is still primitive ([README.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md#L69)).

### 2. What the platform flow already assumes

The current platform document describes the release chain like this:

```text
source repo
  -> run tests
  -> build Docker image
  -> publish immutable GHCR tag
  -> patch image field in GitOps deployment manifest
  -> open/update GitOps PR
  -> merge PR
  -> Argo CD reconciles
```

That is the right mental model and should remain unchanged after extraction. The control-plane boundaries are already well defined:

- source repo owns source code, tests, artifact build logic, publish workflow, target metadata, and helper logic
- GitOps repo owns desired deployment state
- cluster owns runtime state

Those boundaries are explicit in the doc and should be preserved by the new tooling surface, not blurred ([source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md#L12), [source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md#L27), [source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md#L79)).

### 3. What is currently reusable only by copying

The current workflow template performs two jobs:

1. build/test/publish the image
2. call a local Python helper to open the GitOps PR

The template hardcodes a Python setup step and then calls `python3 scripts/open_gitops_pr.py` with the repo-local config path and image ref ([publish-image-ghcr.template.yml](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml#L75), [publish-image-ghcr.template.yml](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml#L90)).

The extracted helper itself already has the core generic logic:

- load `deploy/gitops-targets.json`
- select one or more targets
- clone the GitOps repo
- patch the container image in the manifest
- create a branch and commit
- optionally push and open a PR

That is visible directly in the implementation ([open_gitops_pr.py](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py#L29), [open_gitops_pr.py](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py#L70), [open_gitops_pr.py](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py#L205), [open_gitops_pr.py](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py#L260)).

The target metadata contract is also already explicit and minimal:

```json
{
  "targets": [
    {
      "name": "my-app-prod",
      "gitops_repo": "wesen/2026-03-27--hetzner-k3s",
      "gitops_branch": "main",
      "manifest_path": "gitops/kustomize/my-app/deployment.yaml",
      "container_name": "my-app"
    }
  ]
}
```

This file is intentionally app-local because it names the app's own deployment target and manifest path ([image-gitops-targets.example.json](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/examples/platform/image-gitops-targets.example.json#L1)).

### 4. Evidence of duplication in source repos

The copied workflow shape is not theoretical. `hair-booking` and `mysql-ide` both carry almost the same `publish-image.yaml` pipeline:

- run Go tests
- build/push to GHCR
- install Python
- call `python3 scripts/open_gitops_pr.py`

See:

- [hair-booking publish-image.yaml](/home/manuel/code/wesen/hair-booking/.github/workflows/publish-image.yaml#L19)
- [mysql-ide publish-image.yaml](/home/manuel/code/wesen/2026-03-27--mysql-ide/.github/workflows/publish-image.yaml#L19)

Likewise, `hair-booking` still carries a copied `open_gitops_pr.py` implementation with the same core mechanics as the extracted infra-tooling version ([hair-booking open_gitops_pr.py](/home/manuel/code/wesen/hair-booking/scripts/open_gitops_pr.py#L17)).

Observed during investigation:

- multiple app repos now contain `scripts/open_gitops_pr.py`
- multiple app repos now contain `deploy/gitops-targets.json`
- several repos use near-identical `publish-image.yaml` workflows

That means the platform has crossed the threshold where copy-paste costs more than centralization.

## Gap Analysis

The current system is close, but not finished. The gaps are mostly packaging and lifecycle gaps.

### Gap 1: No versioned reusable workflow API

`infra-tooling` has a template file, but GitHub reusable workflows are discovered under `.github/workflows` and are invoked via `workflow_call`. The current repo does not yet expose the publish pipeline in that form. According to GitHub's documentation, reusable workflows must live in `.github/workflows` and include `on: workflow_call` in order to be callable from other repositories. That means the current template is documentation, not a product surface yet.

### Gap 2: No versioned action for the GitOps updater toolchain

The current GitOps PR helper is a Python script. Source repos can reuse it only by copying the file or by cloning another repo at runtime, which the current docs explicitly avoid. This is exactly the scenario where an action is useful: package the toolchain once, give it a stable input/output interface, and let callers reference it by version.

### Gap 3: Unclear split between reusable workflow and action responsibilities

Right now the copied template mixes concerns:

- repository-specific build/test logic
- registry publish plumbing
- GitOps PR handoff

Those should not all be packaged the same way. The build/test orchestration belongs in a reusable workflow because it spans multiple jobs and GitHub workflow permissions. The GitOps PR updater belongs in an action because it is a discrete tool invocation with inputs and outputs.

### Gap 4: No stable migration path for existing repos

Even if the reusable assets are built, source repos still need a documented migration path:

- what changes in the caller workflow
- what stays local
- which secrets remain required
- how to validate the transition safely

### Gap 5: Testing and contract validation are under-specified

The current helper script has no obvious repo-local test harness in `infra-tooling`, no JSON schema for the config file, and no self-test workflow proving that the packaged reusable assets actually work end to end.

## Proposed Architecture

### Recommendation Summary

Keep everything in `infra-tooling`. Do not create a separate repository yet.

Add two first-class reusable surfaces:

1. reusable workflow:
   - `.github/workflows/publish-ghcr-image.yml`
2. reusable Docker action:
   - `actions/open-gitops-pr/action.yml`
   - `actions/open-gitops-pr/Dockerfile`
   - package or wrap the canonical GitOps PR helper implementation

Keep these repo-local in each application repository:

- `Dockerfile`
- test command or test workflow inputs
- `deploy/gitops-targets.json`
- any app-specific metadata such as image description

### Why this split is correct

GitHub provides three relevant reuse mechanisms:

1. reusable workflows
2. composite actions
3. Docker actions

The official docs distinguish reusable workflows from composite actions clearly: composite actions bundle a series of steps into a single step, while reusable workflows allow one workflow to call another workflow. That is exactly the distinction we need.

For this platform:

- reusable workflow is the right abstraction for multi-job orchestration, permissions, triggers, and registry publishing
- Docker action is the right abstraction for a packaged toolchain that needs `git`, `gh`, Python, and predictable execution dependencies
- composite action is less attractive here because the GitOps helper is not just a handful of shell steps; it is a proper tool with dependencies and likely future tests

## High-Level System Diagram

```text
app repo
  ├─ Dockerfile
  ├─ deploy/gitops-targets.json
  └─ .github/workflows/release.yml
        uses -> infra-tooling/.github/workflows/publish-ghcr-image.yml@v1
                     ├─ run repo-specific test command
                     ├─ build/push immutable GHCR image
                     └─ uses -> infra-tooling/actions/open-gitops-pr@v1
                                 ├─ load target config from caller repo
                                 ├─ clone GitOps repo
                                 ├─ patch image field
                                 ├─ commit/push branch
                                 └─ open or reuse PR

GitOps repo
  └─ manifest image updated by PR

Argo CD
  └─ reconciles merged desired state

K3s cluster
  └─ rolls new immutable image
```

### Proposed Files in `infra-tooling`

```text
.github/workflows/
  publish-ghcr-image.yml

actions/open-gitops-pr/
  action.yml
  Dockerfile
  entrypoint.sh

scripts/gitops/
  open_gitops_pr.py
  validate_gitops_targets.py

docs/platform/
  source-repo-to-gitops-pr.md
  reusable-github-actions-toolkit.md

examples/platform/
  image-gitops-targets.example.json
  caller-workflow.example.yml
```

### Action and Workflow API Design

#### Reusable workflow contract

Recommended caller shape:

```yaml
jobs:
  release:
    uses: wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml@v1
    permissions:
      contents: read
      packages: write
      pull-requests: write
    secrets: inherit
    with:
      dockerfile: ./Dockerfile
      build-context: .
      test-command: go test ./...
      image-name: ghcr.io/${{ github.repository }}
      gitops-target-config: deploy/gitops-targets.json
      open-gitops-pr: true
```

Recommended workflow inputs:

- `dockerfile`
- `build-context`
- `test-command`
- `image-name`
- `platforms`
- `gitops-target-config`
- `open-gitops-pr`
- `gitops-target`
- `gitops-all-targets`

Recommended workflow secrets:

- `GITOPS_PR_TOKEN`

Recommended workflow outputs:

- `image-ref`
- `image-tag`
- `gitops-pr-opened`

#### Docker action contract

Recommended `action.yml` interface:

```yaml
name: open-gitops-pr
description: Patch GitOps deployment manifests to a new immutable image and open or update pull requests.
inputs:
  config:
    required: true
  image:
    required: true
  target:
    required: false
  all-targets:
    required: false
    default: "true"
  push:
    required: false
    default: "true"
  open-pr:
    required: false
    default: "true"
  git-author-name:
    required: false
  git-author-email:
    required: false
outputs:
  branch-name:
    description: Created branch name when a manifest changed
  pr-number:
    description: Open PR number when available
runs:
  using: docker
  image: Dockerfile
```

The exact output set can evolve, but the key point is that the action should act like a real product with a declared input/output contract instead of "run this script if you copied it correctly."

### Why Docker action instead of composite action

Composite actions are good when you want to bundle a sequence of workflow steps and let the runner provide the environment. That is not a perfect fit here because the GitOps helper needs a packaged execution environment:

- `git`
- `gh`
- Python runtime
- any future validation or YAML tooling

A Docker action is better for that because it packages the toolchain and reduces caller-side environment drift. The GitHub docs for Docker actions and action metadata are the relevant implementation references.

Observed tradeoff:

- reusable workflow:
  - best for orchestration
  - can run multiple jobs
  - can define permissions and secrets at workflow level
- Docker action:
  - best for encapsulating the GitOps updater toolchain
  - stable runtime
  - easier to test as one unit
- composite action:
  - simpler, but depends more heavily on caller runner environment
  - weaker fit for a growing CLI-style helper

## Detailed Flow Design

### Flow A: Normal successful publish and PR creation

```text
caller repo workflow starts
  -> reusable workflow receives inputs
  -> test command runs
  -> image metadata generated
  -> GHCR login occurs for non-PR events
  -> immutable image tag is pushed
  -> workflow computes final image ref
  -> Docker action loads deploy/gitops-targets.json
  -> action clones GitOps repo target branch
  -> action patches manifest image for matching container
  -> action commits + pushes branch
  -> action opens PR unless an identical branch PR already exists
  -> workflow exposes image-ref output
```

### Flow B: Manifest already points at target image

```text
action loads config
  -> action reads manifest
  -> target container image already matches desired immutable tag
  -> no commit
  -> no push
  -> no PR
  -> action emits changed=false
```

### Flow C: No GitOps token configured

```text
workflow builds and pushes image
  -> workflow sees no GITOPS_PR_TOKEN
  -> workflow either:
       1. skips GitOps PR with clear notice
       2. fails fast if strict mode is enabled
```

The recommended default is:

- skip with explicit notice for early adoption
- add a strict mode input later if desired

## Pseudocode

### Reusable workflow pseudocode

```text
on workflow_call(inputs, secrets):
  validate required inputs

  job test_and_publish:
    checkout caller repo
    optionally setup language runtime
    run inputs.test-command
    docker metadata-action
    if non-PR event:
      login to ghcr
      build and push immutable tag
    output image_ref = ghcr image with sha-short tag

  job gitops_pr:
    needs test_and_publish
    if inputs.open-gitops-pr and branch == main:
      checkout caller repo
      invoke open-gitops-pr action with:
        config = inputs.gitops-target-config
        image = needs.test_and_publish.outputs.image_ref
        token = secrets.GITOPS_PR_TOKEN
```

### Action pseudocode

```text
main():
  parse action inputs
  load target config
  select targets
  for each target:
    clone target.gitops_repo at target.gitops_branch
    patch target.manifest_path container image to desired image
    if no change:
      continue
    create deterministic branch name
    commit manifest update
    if push enabled:
      push branch
    if open-pr enabled:
      create or reuse pull request
  emit outputs
```

### Suggested deterministic branch format

```text
automation/<app-name>-<target-name>-<sha-tag>
```

This matches the existing helper's spirit and keeps PR deduplication tractable.

## Implementation Plan

### Phase 1: Normalize the canonical helper inside `infra-tooling`

Goal: make `scripts/gitops/open_gitops_pr.py` the single authoritative implementation.

Tasks:

1. Review differences across copied app-repo variants.
2. Fold the best improvements back into the infra-tooling copy.
3. Add tests for:
   - target loading
   - target selection
   - manifest patch success
   - no-op patch
   - missing container/image errors
4. Add a validator for `deploy/gitops-targets.json`.

Files:

- `scripts/gitops/open_gitops_pr.py`
- `scripts/gitops/validate_gitops_targets.py`
- `tests/...` or repo-appropriate test directory

### Phase 2: Add Docker action packaging

Goal: expose the helper as a proper reusable action.

Tasks:

1. Create `actions/open-gitops-pr/action.yml`.
2. Create `actions/open-gitops-pr/Dockerfile`.
3. Add a minimal entrypoint wrapper that maps GitHub Action inputs to CLI args and environment variables.
4. Emit action outputs using GitHub Actions output conventions.

Files:

- `actions/open-gitops-pr/action.yml`
- `actions/open-gitops-pr/Dockerfile`
- `actions/open-gitops-pr/entrypoint.sh`

### Phase 3: Add reusable workflow packaging

Goal: stop handing source repos a copy template and instead give them a callable workflow.

Tasks:

1. Move the template logic into `.github/workflows/publish-ghcr-image.yml`.
2. Replace hardcoded assumptions with workflow inputs.
3. Ensure workflow permissions and outputs are documented.
4. Call the shared action instead of `python3 scripts/open_gitops_pr.py`.

Files:

- `.github/workflows/publish-ghcr-image.yml`
- `templates/github/publish-image-ghcr.template.yml`

Likely outcome:

- keep the template only as a caller example or deprecate it entirely

### Phase 4: Update docs and examples

Goal: make onboarding obvious for an intern.

Tasks:

1. Update `docs/platform/source-repo-to-gitops-pr.md`.
2. Add a new doc explaining:
   - how to call the reusable workflow
   - what remains local
   - how version pinning works
   - how to do a dry-run rollout
3. Add an example caller workflow.
4. Keep the example `deploy/gitops-targets.json` file.

Files:

- `docs/platform/source-repo-to-gitops-pr.md`
- `docs/platform/reusable-github-actions-toolkit.md`
- `examples/platform/caller-workflow.example.yml`

### Phase 5: Migrate pilot repos

Goal: prove the reusable surface on one or two real repositories before mass migration.

Good pilot candidates:

- `hair-booking`
- `mysql-ide`
- `sanitize`

Migration sequence:

1. replace copied workflow with reusable workflow call
2. remove local copied GitOps helper if no longer needed
3. keep `deploy/gitops-targets.json`
4. run a real release to confirm PR handoff still works

## Testing and Validation Strategy

### Unit tests

Add tests for:

- config validation
- selecting targets
- branch name generation
- PR body generation
- manifest patch logic

### Fixture tests

Use small fixture manifests for:

- single-container deployment
- multiple containers
- already-current image
- missing target container

### Integration tests

At minimum:

1. create temp Git repo representing GitOps repo
2. run helper in dry-run mode
3. verify manifest diff
4. run helper in commit mode against local repo dir
5. verify commit created

### Self-hosted repo validation

Inside `infra-tooling` itself:

- add a workflow that lint-checks the action metadata
- add a workflow that invokes the reusable workflow from an example caller repo or fixture
- ensure the reusable workflow can call the local action cleanly

### Manual operator validation

For the first migrated app:

1. run the source repo workflow on `main`
2. confirm immutable image published to GHCR
3. confirm GitOps PR created against the infra repo
4. review manifest diff
5. merge PR
6. confirm Argo sees the change
7. confirm Kubernetes rolls the new image

## Risks, Alternatives, and Open Questions

### Risk 1: Over-generalizing the workflow

If the reusable workflow tries to encode every build style from day one, it will become harder to use than the copied template. Start with the dominant Go + Docker path and add inputs only where variation is real.

### Risk 2: Hidden runner assumptions

A composite action would inherit more of the runner environment and could fail differently across callers. The Docker action reduces that risk but introduces its own maintenance cost. That trade is still favorable here.

### Risk 3: Weak config validation

If `deploy/gitops-targets.json` remains loosely validated, callers will get runtime failures after publishing an image. Add schema validation early.

### Risk 4: Breaking existing source repos during migration

Do not migrate every repo at once. Use one or two pilot repos first and compare the produced PRs against the current local-script path.

### Alternative A: Keep the template-and-copy model

Pros:

- no GitHub reusable action packaging work
- repos stay self-contained

Cons:

- bug fixes keep getting copied manually
- behavior drifts
- harder to evolve branch naming, PR body format, config validation, and patch logic consistently

This alternative is no longer attractive.

### Alternative B: Separate repository just for actions

Pros:

- tighter product boundary
- cleaner public reuse story

Cons:

- another repository to maintain
- unnecessary split given `infra-tooling` already exists for this purpose

This is not recommended yet.

### Open Questions

1. Should the reusable workflow support only Go repos initially, or should it accept a generic `test-command` and remain language-agnostic from day one?
2. Should missing `GITOPS_PR_TOKEN` skip with warning or fail in strict mode?
3. Should the action remain Python-based or be rewritten later in Go for easier static distribution and testing?
4. Should the action patch only plain `Deployment` manifests, or should it support more Kustomize/Helm-like shapes later?
5. What versioning and release process should govern `@v1`, `@v2`, and breaking changes?

## Official GitHub Actions Reference Notes

These docs are relevant to implementation:

- Reusable workflows:
  - GitHub documents that reusable workflows live in `.github/workflows` and use `on: workflow_call`.
- Composite actions:
  - GitHub distinguishes composite actions from reusable workflows and describes composite actions as bundling steps into a single action step.
- Docker actions:
  - GitHub documents Docker container actions as packaged actions with their own runtime environment.
- Metadata syntax:
  - GitHub documents `action.yml` schema and `runs.using` values for action packaging.

Use the official URLs in the frontmatter `ExternalSources` list as the canonical external references when implementing.

## References

### Core repo files

- [README.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md)
- [source-repo-to-gitops-pr.md](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md)
- [publish-image-ghcr.template.yml](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml)
- [open_gitops_pr.py](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py)
- [image-gitops-targets.example.json](/home/manuel/code/wesen/corporate-headquarters/infra-tooling/examples/platform/image-gitops-targets.example.json)

### Example source repos

- [hair-booking publish-image.yaml](/home/manuel/code/wesen/hair-booking/.github/workflows/publish-image.yaml)
- [mysql-ide publish-image.yaml](/home/manuel/code/wesen/2026-03-27--mysql-ide/.github/workflows/publish-image.yaml)
- [hair-booking open_gitops_pr.py](/home/manuel/code/wesen/hair-booking/scripts/open_gitops_pr.py)

### Official docs

- https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
- https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
- https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action
- https://docs.github.com/en/actions/sharing-automations/creating-actions/metadata-syntax-for-github-actions
