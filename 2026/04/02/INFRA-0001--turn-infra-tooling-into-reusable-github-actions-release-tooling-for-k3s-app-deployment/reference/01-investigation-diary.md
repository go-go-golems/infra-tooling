---
Title: Investigation diary
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
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../2026-03-27--mysql-ide/.github/workflows/publish-image.yaml
      Note: Second source repo with near-identical workflow shape
    - Path: ../../../../../../../hair-booking/.github/workflows/publish-image.yaml
      Note: |-
        Example copied workflow in a source repo
        Example copied workflow reviewed during investigation
    - Path: README.md
      Note: Repo charter reviewed during investigation
    - Path: docs/platform/source-repo-to-gitops-pr.md
      Note: |-
        Current control-plane and duplication assumptions
        Current platform model reviewed during investigation
    - Path: scripts/gitops/open_gitops_pr.py
      Note: |-
        Current extracted helper implementation
        Helper script reviewed during investigation
    - Path: templates/github/publish-image-ghcr.template.yml
      Note: |-
        Current template-based workflow distribution
        Template workflow reviewed during investigation
ExternalSources:
    - https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows
    - https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action
    - https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action
Summary: Chronological investigation log for designing reusable GitHub Actions automation in infra-tooling for K3s app deployment.
LastUpdated: 2026-04-02T12:35:00-04:00
WhatFor: Use this diary to understand how the design was derived, which evidence informed it, and how the final bundle was delivered.
WhenToUse: Read this when reviewing the design doc or continuing the reusable tooling implementation later.
---


# Investigation diary

## Goal

Capture the evidence-first investigation that produced the design for turning `infra-tooling` into the versioned reusable GitHub Actions home for GHCR publish and GitOps PR handoff.

## Step 1: Map the current infra-tooling extraction state and the duplicated source-repo pattern

The first task was to understand whether `infra-tooling` was only an idea, a template stash, or already the intended product home for shared deployment automation. That mattered because the design changes depending on the answer. If the repo were only an archive of notes, the right recommendation might be a new repository. If the repo were already the neutral shared home, the right recommendation would be to package what is already here better.

The key conclusion from this step was that `infra-tooling` is already the correct home. The repo charter, the platform docs, the workflow template, and the extracted GitOps helper all point in the same direction. The missing work is packaging the current extracted mechanics as real GitHub Actions surfaces rather than continuing to copy them into source repositories.

### Prompt Context

**User prompt (verbatim):** "Ok, create a new ticket in the infra-tooling repo using docmgr --root which is about creating this reusable tooling and images to deploy further go-go-golems apps to k3s.

Create a detailed analysis / design / implementation guide that is very detailed for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file
  references.
  It should be very clear and detailed. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticket rooted in the `infra-tooling` repository, inspect the current shared release assets and their app-repo copies, analyze how those should become reusable GitHub Actions components, write an intern-facing implementation guide and diary, validate the ticket, and upload the bundle to reMarkable.

**Inferred user intent:** Stop proliferating copied workflow scripts across go-go-golems app repos and instead define the next platform extraction step clearly enough that an intern can implement it.

**Commit (code):** N/A

### What I did

- Created ticket `INFRA-0001` under `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment`.
- Read:
  - `README.md`
  - `docs/platform/source-repo-to-gitops-pr.md`
  - `templates/github/publish-image-ghcr.template.yml`
  - `scripts/gitops/open_gitops_pr.py`
  - `examples/platform/image-gitops-targets.example.json`
- Compared the shared workflow/script model against:
  - `hair-booking`
  - `mysql-ide`
- Checked official GitHub docs for:
  - reusable workflows
  - composite actions
  - Docker container actions
  - action metadata syntax

### Why

- The user asked specifically for reusable tooling and images in `infra-tooling`, so I needed to determine whether that repo was already intended to hold this work.
- The design needed to separate what belongs in app repos from what belongs in shared automation.
- GitHub Actions has multiple packaging options, and recommending the wrong one would create avoidable churn.

### What worked

- The repo README was unambiguous: `infra-tooling` already claims ownership of shared release and GitOps automation.
- The platform doc already described the correct control-plane model.
- The workflow template and helper script already contained most of the desired logic.
- Example source repos showed the copied workflow shape clearly enough that the duplication problem was easy to prove.

### What didn't work

- I initially assumed the `docmgr --root` ticket files would be created under `ttmp/`, but in this repository the docs root is the repo root itself. That caused one short path mismatch while opening generated files.
- The current platform doc still argues that copying the helper script into each app repo is intentional. That is not a tool failure, but it is the exact product assumption this ticket now overturns.

### What I learned

- `infra-tooling` is not a future idea. It is already the intended shared home, but it currently distributes shared mechanics as templates and scripts instead of callable versioned automation.
- The right packaging split is not "one big action." The workflow orchestration and the GitOps updater tool are distinct enough that they should be separate reusable surfaces.
- The app-local `deploy/gitops-targets.json` contract should remain local because it describes deployment-specific facts that do not belong in the shared repo.

### What was tricky to build

- The subtle part was not identifying duplication. It was deciding the correct GitHub Actions packaging boundary. Reusable workflows, composite actions, and Docker actions all reduce duplication, but they solve different problems. The design had to explain why the publish orchestration belongs in a reusable workflow while the GitOps updater belongs in a Docker-backed action.
- Another subtle point was preserving the current control-plane boundary. The point is not to make app repos "thinner" at any cost. The point is to centralize shared mechanics without moving app-specific Docker inputs or deployment target metadata into the wrong repo.

### What warrants a second pair of eyes

- Whether the initial reusable workflow should be Go-specific or generic from day one.
- Whether missing `GITOPS_PR_TOKEN` should skip or fail by default.
- Whether the current manifest patcher is good enough to keep initially or should be made YAML-aware before packaging.
- Whether action outputs should include PR URLs or PR numbers from the start.

### What should be done in the future

- Productize the reusable workflow under `.github/workflows/`.
- Productize the GitOps updater as a Docker action under `actions/`.
- Add tests and config validation.
- Migrate one pilot repo first, then roll out to the rest.

### Code review instructions

- Start with the main design doc:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md`
- Then inspect these core references:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py`
- Finally compare the copied workflow examples:
  - `/home/manuel/code/wesen/hair-booking/.github/workflows/publish-image.yaml`
  - `/home/manuel/code/wesen/2026-03-27--mysql-ide/.github/workflows/publish-image.yaml`

### Technical details

- Ticket creation commands:
  - `docmgr ticket create-ticket --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --ticket INFRA-0001 --title "Turn infra-tooling into reusable GitHub Actions release tooling for K3s app deployment" --topics github-actions,ghcr,gitops,argocd,k3s,platform,release-engineering`
  - `docmgr doc add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --ticket INFRA-0001 --doc-type design-doc --title "Reusable GitHub Actions and action-image toolkit for K3s app deployment"`
  - `docmgr doc add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --ticket INFRA-0001 --doc-type reference --title "Investigation diary"`
- Key local files inspected:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/examples/platform/image-gitops-targets.example.json`
  - `/home/manuel/code/wesen/hair-booking/.github/workflows/publish-image.yaml`
  - `/home/manuel/code/wesen/2026-03-27--mysql-ide/.github/workflows/publish-image.yaml`
- Official docs consulted:
  - `https://docs.github.com/en/actions/how-tos/reuse-automations/reuse-workflows`
  - `https://docs.github.com/en/actions/tutorials/create-actions/create-a-composite-action`
  - `https://docs.github.com/en/actions/tutorials/use-containerized-services/create-a-docker-container-action`
  - `https://docs.github.com/en/actions/sharing-automations/creating-actions/metadata-syntax-for-github-actions`

## Related

- Design doc: [01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md](../design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md)
- Ticket index: [index.md](../index.md)
- Tasks: [tasks.md](../tasks.md)

## Step 2: Validate the ticket and deliver the bundle to reMarkable

Once the design and diary were written, the remaining work was delivery hygiene. I needed to make sure the ticket validated cleanly under `docmgr`, resolve any vocabulary gaps in this repo's new documentation root, and then upload a single ordered PDF bundle to reMarkable.

The key result from this step was that `INFRA-0001` passed `docmgr doctor` cleanly after I added the new topic vocabulary entries, the dry-run bundle looked correct, and the final upload succeeded to the dated ticket folder on reMarkable.

### What I did

- Ran `docmgr doctor --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --ticket INFRA-0001 --stale-after 30`.
- Added topic vocabulary entries for:
  - `github-actions`
  - `ghcr`
  - `gitops`
  - `argocd`
  - `k3s`
  - `platform`
  - `release-engineering`
- Re-ran `docmgr doctor` and confirmed that all checks passed.
- Verified reMarkable status and cloud auth.
- Ran a dry-run bundle upload for:
  - `index.md`
  - the design doc
  - the diary
  - `tasks.md`
  - `changelog.md`
- Uploaded the final bundle to `/ai/2026/04/02/INFRA-0001`.

### Why

- The user asked not just for a ticket bundle, but also for reMarkable delivery.
- This repository did not yet have the needed topic vocabulary entries, so validation had to be made clean explicitly rather than ignored.
- The dry-run step confirmed both file order and remote destination before a real upload.

### What worked

- `docmgr doctor` clearly identified the only remaining issue.
- Adding the vocabulary entries was straightforward.
- The dry-run correctly showed the default layout and intended bundle contents.
- The final upload succeeded on the first real attempt.

### What didn't work

- The first validation pass surfaced new-topic warnings because this repository started with no ticket vocabulary for the K3s/GitHub Actions domain. That was expected once I saw the clean docs root state, but it still needed explicit cleanup.

### What I learned

- A new repo-level docmgr root often needs vocabulary bootstrapping before its first serious ticket passes cleanly.
- For this kind of handoff document, a bundled PDF is the right reMarkable shape because it preserves reading order and a single table of contents.

### Technical details

- Validation commands:
  - `docmgr doctor --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --ticket INFRA-0001 --stale-after 30`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug github-actions --description 'GitHub Actions workflows, actions, and CI/CD automation'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug ghcr --description 'GitHub Container Registry publishing and image distribution'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug gitops --description 'GitOps pull-request-based desired state delivery'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug argocd --description 'Argo CD application reconciliation and deployment control plane'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug k3s --description 'K3s cluster deployment and runtime concerns'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug platform --description 'Shared platform architecture and tooling'`
  - `docmgr vocab add --root /home/manuel/code/wesen/corporate-headquarters/infra-tooling --category topics --slug release-engineering --description 'Build, packaging, and release automation workflows'`
- Delivery commands:
  - `remarquee status`
  - `remarquee cloud account --non-interactive`
- `remarquee upload bundle --dry-run --non-interactive --toc-depth 2 --name "INFRA-0001 infra-tooling reusable GitHub Actions toolkit" --remote-dir "/ai/2026/04/02/INFRA-0001" 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/index.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/reference/01-investigation-diary.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/tasks.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/changelog.md`
- `remarquee upload bundle --non-interactive --toc-depth 2 --name "INFRA-0001 infra-tooling reusable GitHub Actions toolkit" --remote-dir "/ai/2026/04/02/INFRA-0001" 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/index.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/design-doc/01-reusable-github-actions-and-action-image-toolkit-for-k3s-app-deployment.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/reference/01-investigation-diary.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/tasks.md 2026/04/02/INFRA-0001--turn-infra-tooling-into-reusable-github-actions-release-tooling-for-k3s-app-deployment/changelog.md`

## Step 3: Implement the shared GitOps updater action, validator, and unit tests

After the design work was done, I started implementation with the smallest reusable slice that would unlock the rest of the system: the GitOps updater itself. The reusable workflow depends on having a stable helper contract, so the first concrete step was to package the helper logic into a real action-facing implementation instead of leaving it as only a copied source-repo script.

The key result from this step was that `infra-tooling` now has a reusable `open-gitops-pr` action directory, a canonical helper implementation behind it, a wrapper CLI for local usage, a config validator, and passing unit tests for the critical config and manifest-patching behavior.

### What I did

- Added `actions/open-gitops-pr/action.yml`.
- Added `actions/open-gitops-pr/Dockerfile`.
- Added `actions/open-gitops-pr/entrypoint.sh`.
- Added the canonical implementation under `actions/open-gitops-pr/src/gitops_pr_action/open_gitops_pr.py`.
- Replaced `scripts/gitops/open_gitops_pr.py` with a thin wrapper that imports and calls the canonical implementation.
- Added `scripts/gitops/validate_gitops_targets.py`.
- Added unit tests under `tests/gitops/test_open_gitops_pr.py`.
- Added package markers in `tests/` so `unittest` discovery works reliably.

### Why

- The helper logic is the most duplicated part of the current source-repo model, so centralizing it first gives the biggest payoff.
- The reusable workflow should call a stable action interface, not shell out to an ad hoc copied local script.
- Validation and tests had to land before more workflow code, otherwise the packaging work would have been easy to break silently.

### What worked

- The refactor preserved the existing CLI shape closely enough that the wrapper script remained simple.
- The validator worked immediately against `examples/platform/image-gitops-targets.example.json`.
- The unit tests passed once test discovery was fixed with package markers.
- `git diff --check` stayed clean during this slice.

### What didn't work

- `python3 -m unittest discover -s tests -v` initially reported `NO TESTS RAN` because the nested test directories were not package-marked yet.
- A quick dry-run call of `scripts/gitops/open_gitops_pr.py` against the example config failed with `manifest not found`, which was expected once I remembered that the example target points at a GitOps manifest path that does not exist inside `infra-tooling` itself. That was a useful reminder that the validator and the patching helper solve different problems.
- The Docker action entrypoint initially used hyphenated input env names. Those needed to be corrected to the underscore-style environment variable names GitHub Actions exposes to Docker actions.

### What I learned

- The helper logic benefits from being split into:
  - a canonical importable implementation
  - a thin repo-local wrapper CLI
  - an action entrypoint wrapper
- The config validator is worth keeping separate from the patching helper because it can be run cheaply in lint-style workflows before any GitOps write path is attempted.
- Even simple action packaging details like input env naming are easy to get subtly wrong if they are not exercised immediately.

### Technical details

- Commands run:
  - `python3 -m unittest discover -s tests -v`
  - `python3 -m unittest tests.gitops.test_open_gitops_pr -v`
  - `python3 scripts/gitops/validate_gitops_targets.py examples/platform/image-gitops-targets.example.json`
  - `python3 scripts/gitops/open_gitops_pr.py --config examples/platform/image-gitops-targets.example.json --target my-app-prod --image ghcr.io/wesen/my-app:sha-1234567 --gitops-repo-dir . --dry-run`
  - `git diff --check`

### Review instructions

- Focus on these files for the implementation slice:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/actions/open-gitops-pr/action.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/actions/open-gitops-pr/Dockerfile`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/actions/open-gitops-pr/entrypoint.sh`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/actions/open-gitops-pr/src/gitops_pr_action/open_gitops_pr.py`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/open_gitops_pr.py`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/validate_gitops_targets.py`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/tests/gitops/test_open_gitops_pr.py`

## Step 4: Add the reusable workflow, caller examples, and repo-local self-test workflow

With the action packaged, the next job was to make it usable by source repositories. That meant adding a real reusable workflow under `.github/workflows/`, updating the old template into a caller example instead of an inline implementation, and wiring a repo-local self-test workflow that exercises the reusable workflow surface.

The key result from this step was that `infra-tooling` now exposes a real `publish-ghcr-image` reusable workflow, a caller example/template, updated platform docs, and a self-test workflow that builds the action image and calls the reusable workflow in smoke-test mode.

### What I did

- Added `.github/workflows/publish-ghcr-image.yml`.
- Added `.github/workflows/test-gitops-tooling.yml`.
- Added `examples/platform/publish-image-ghcr.caller.example.yml`.
- Replaced `templates/github/publish-image-ghcr.template.yml` with a caller workflow that uses the reusable workflow.
- Updated `docs/platform/source-repo-to-gitops-pr.md`.
- Updated `README.md`.

### Why

- Without a reusable workflow, the packaged action still leaves every app repo to rebuild the same publish orchestration.
- The old template had become the wrong abstraction because it embedded the implementation instead of calling a versioned shared surface.
- A self-test workflow is necessary because reusable workflows are easy to write but hard to trust unless the repo itself exercises them.

### What worked

- The reusable workflow cleanly separates build/publish concerns from the GitOps PR step.
- Checking out `infra-tooling` into `.infra-tooling` inside the reusable workflow creates a stable path for calling the packaged action.
- Local validation stayed clean:
  - unit tests passed
  - `compileall` passed
  - `bash -n` passed for the action entrypoint
  - the action Docker image built successfully

### What didn't work

- I initially used hyphenated reusable-workflow input names and output names. That was mechanically awkward and easy to get wrong in expressions, so I normalized the workflow-facing contract to underscore-style names before proceeding.
- The same review pass surfaced that workflow/job outputs should also use underscore-style names, not hyphenated names.

### What I learned

- Reusable workflow ergonomics matter. A slightly awkward input/output contract will create friction in every caller repo, so it is worth normalizing early.
- The self-test workflow is useful not just for CI later, but as a design forcing function: it made the reusable workflow interface concrete enough to validate immediately.

### Technical details

- Commands run:
  - `python3 -m unittest discover -s tests -v`
  - `python3 -m compileall actions scripts tests`
  - `bash -n actions/open-gitops-pr/entrypoint.sh`
  - `python3 scripts/gitops/validate_gitops_targets.py examples/platform/image-gitops-targets.example.json`
  - `git diff --check`
  - `docker build -t infra-tooling-open-gitops-pr ./actions/open-gitops-pr`

### Review instructions

- Focus on these files for the workflow/documentation slice:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/.github/workflows/test-gitops-tooling.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/examples/platform/publish-image-ghcr.caller.example.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/templates/github/publish-image-ghcr.template.yml`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/docs/platform/source-repo-to-gitops-pr.md`
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/README.md`

## Step 5: Add fixture coverage and a dry-run integration test for the helper

Once the reusable workflow existed, the next useful improvement was test depth instead of more surface area. I wanted the helper to prove three things reliably:

1. it patches the correct container when there are multiple containers,
2. it restores files correctly during dry-run local validation,
3. it writes machine-readable outputs in the shape the action metadata expects.

The key result from this step was that the helper test suite now covers the multi-container patch case, the action output file case, and a dry-run integration path against a temporary Git repository.

### What I did

- Extended `tests/gitops/test_open_gitops_pr.py` with:
  - a multi-container manifest fixture test
  - a dry-run integration test using `git init` in a temp repo
  - a machine-readable output file test for `append_github_outputs`
- Re-ran the unit test suite and `git diff --check`.

### Why

- The helper is now shared infrastructure, so "works on one happy-path manifest" is not enough.
- The dry-run path is important because operators will use it to validate target config and manifest patch behavior before enabling push/PR creation.
- Action outputs are part of the reusable contract and should be treated like code, not just like logging.

### What worked

- The dry-run integration test passed with a temporary Git repository and confirmed that the manifest file is restored after validation.
- The multi-container fixture correctly verified that only the named container image changes.
- The output-file helper wrote the expected `changed`, `changed_targets`, `branch_names`, and `pr_numbers` keys.

### What didn't work

- The dry-run integration test prints the unified diff emitted by `process_target`, so the `unittest` output is slightly noisier than a pure silent unit test run. That is acceptable for now because it also makes the tested behavior more obvious during failures.

### What I learned

- The dry-run path is a real behavior boundary, not just a flag. It deserved its own integration-style test once the helper became shared tooling.
- Shared helper outputs should be tested explicitly because output key naming drift is easy to introduce during refactors.

### Technical details

- Commands run:
  - `python3 -m unittest discover -s tests -v`
  - `git diff --check`

### Review instructions

- Focus on:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/tests/gitops/test_open_gitops_pr.py`

## Step 6: Start the first pilot adoption in smailnail

With the shared workflow in place, I wanted at least one real caller repository to depend on it immediately. `smailnail` was the obvious candidate because it is the next K3s migration target and already has a production Dockerfile, but it does not yet have the K3s GitOps package or `deploy/gitops-targets.json`.

The key result from this step was a low-risk first adoption: `smailnail` now has a `publish-image` workflow that uses the shared reusable workflow for build and GHCR publish, while deliberately keeping `open_gitops_pr: false` until the target manifests exist. That exercises the shared publish path without pretending the GitOps rollout path is already ready.

### What I did

- Added `/home/manuel/code/wesen/corporate-headquarters/smailnail/.github/workflows/publish-image.yaml`.
- Pointed that workflow at `wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml@main`.
- Reused the repo's existing `go generate ./...` and `go test ./...` checks as the test command.
- Left GitOps PR creation disabled with an explicit comment explaining why.

### Why

- A shared workflow is more credible once a real application repository depends on it.
- `smailnail` is the migration target we were already analyzing, so it is a useful pilot even before its GitOps manifests are in place.
- Enabling image publish before GitOps handoff is a safe partial rollout because it exercises the build and registry parts of the contract without breaking on a nonexistent manifest target.

### What worked

- The caller workflow fits cleanly into `smailnail` without touching its existing release/tag workflows.
- `git diff --check` in `smailnail` stayed clean.
- The partial rollout constraint is explicit in the workflow comment instead of being hidden in assumptions.

### What didn't work

- `smailnail` still lacks `deploy/gitops-targets.json`, which means it cannot yet enable the GitOps PR half of the contract.
- The repo also has unrelated untracked `README.md.orig` and `go.mod.orig` files, so I needed to be careful not to include or disturb them.

### What I learned

- The shared workflow is usable before full GitOps adoption, which gives us a staged migration path:
  - adopt image publish first
  - add target metadata and K3s manifests later
  - then turn on GitOps PR creation
- `smailnail` is a good pilot repo for the publish half of the contract, but not yet for the full GitOps handoff path.

### Technical details

- Commands run:
  - `git diff --check && git status --short --branch`
  - `sed -n '1,220p' .github/workflows/publish-image.yaml`

### Review instructions

- Focus on:
  - `/home/manuel/code/wesen/corporate-headquarters/smailnail/.github/workflows/publish-image.yaml`

## Step 7: Complete the local smailnail pilot wiring and validate the shared handoff contract

The next slice was the first real integration pass across three repositories instead of one. The goal was to move `smailnail` from "publish-only pilot" to "locally validated full pilot" by adding the K3s target package, adding the source-repo target metadata, and proving that the shared updater can patch the real manifest without collateral edits.

This step also surfaced one useful shared-workflow bug that would have made the pilot look broken for the wrong reason. Fixing that before pushing anything kept the remaining validation signal clean: if the live pilot fails later, it will not be because the reusable workflow ignored its own `push_image` input.

### Prompt Context

**User prompt (verbatim):** "ok, 1, 2, 3 , 4"

**Assistant interpretation:** Execute the previously proposed next four migration steps: add the `smailnail` GitOps package in the K3s repo, add the repo-local target config in `smailnail`, enable the shared GitOps PR handoff, and validate the pilot path.

**Inferred user intent:** Stop discussing the migration as a plan and turn the `smailnail` pilot into a concrete, testable end-to-end contract across the shared tooling repo, the source repo, and the K3s GitOps repo.

**Commit (code):** `4562c9e` — "feat(smailnail): add k3s gitops manifests"; `7a0cd58` — "ci: enable smailnail gitops handoff"; `3b6b7a6` — "fix(workflows): honor push_image input"

### What I did

- Created `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/` with:
  - namespace, service account, Vault connection/auth, runtime secret, image-pull secret
  - Postgres bootstrap service account/auth/configmap/job
  - deployment, service, ingress
- Added `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/smailnail.yaml`.
- Used the merged `smailnaild` deployment shape:
  - web UI and `/api/*`
  - OIDC browser login
  - `/.well-known/oauth-protected-resource`
  - `/mcp`
- Wired the deployment for:
  - Postgres via `SMAILNAILD_DSN`
  - encryption key material from Vault
  - web OIDC against `https://auth.yolo.scapegoat.dev/realms/smailnail`
  - MCP bearer auth against the same issuer
- Added `/home/manuel/code/wesen/corporate-headquarters/smailnail/deploy/gitops-targets.json`.
- Updated `/home/manuel/code/wesen/corporate-headquarters/smailnail/.github/workflows/publish-image.yaml` to:
  - set `image_name: ghcr.io/go-go-golems/smailnail`
  - point at `deploy/gitops-targets.json`
  - enable `open_gitops_pr` only for pushes to `refs/heads/main`
- Fixed `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml` so the GHCR login step actually honors `inputs.push_image`.
- Validated the contract with:
  - `kubectl kustomize gitops/kustomize/smailnail`
  - `python3 /home/manuel/code/wesen/corporate-headquarters/infra-tooling/scripts/gitops/validate_gitops_targets.py /home/manuel/code/wesen/corporate-headquarters/smailnail/deploy/gitops-targets.json`
  - `python3 /home/manuel/code/wesen/corporate-headquarters/infra-tooling/actions/open-gitops-pr/src/gitops_pr_action/open_gitops_pr.py --config /home/manuel/code/wesen/corporate-headquarters/smailnail/deploy/gitops-targets.json --target smailnail-prod --image ghcr.io/go-go-golems/smailnail:sha-deadbee --gitops-repo-dir /home/manuel/code/wesen/2026-03-27--hetzner-k3s --dry-run`
  - `python3 -m unittest discover -s tests -v` in `infra-tooling`
  - `go generate ./... && go test ./...` in `smailnail`

### Why

- The shared action was not meaningfully proven until it patched a real target in a real GitOps checkout.
- `smailnail` is the first app that exercises the full merged-hosted shape:
  - browser OIDC
  - shared application DB
  - encrypted stored IMAP credentials
  - mounted MCP bearer-auth surface
- The K3s repo needed a stable deployment target before the source repo could safely enable GitOps PR creation.
- The workflow bug fix belonged in the same implementation window because it directly affected pilot correctness.

### What worked

- `kubectl kustomize gitops/kustomize/smailnail` succeeded, so the package is structurally sound.
- The shared validator accepted the new `deploy/gitops-targets.json`.
- The shared updater dry-run patched only `gitops/kustomize/smailnail/deployment.yaml` and then restored it cleanly.
- `infra-tooling` unit tests still passed after the reusable-workflow fix.
- After installing the locked frontend dependencies, `go generate ./... && go test ./...` passed in `smailnail`.
- The repo boundaries are now clean and intentional:
  - `infra-tooling` owns shared release orchestration
  - `smailnail` owns its target metadata and caller workflow
  - `2026-03-27--hetzner-k3s` owns the deployment package and Argo app

### What didn't work

- The first `go generate ./... && go test ./...` run in `smailnail` failed because the frontend toolchain was not installed locally. The exact failure was:
  - `sh: 1: tsc: not found`
  - `WARN  Local package.json exists, but node_modules missing, did you mean to install?`
  - `pkg/smailnaild/web/generate.go:1: running "go": exit status 1`
- That was not a product bug in `smailnail`; it was a local environment precondition for the workflow’s UI build step. I resolved it by running `corepack enable && pnpm install --frozen-lockfile` in `ui/`, then reran the workflow test command successfully.
- The pilot is still only locally validated. I did not push branches yet, so there is not yet a live GitHub Actions run or a real opened GitOps PR to inspect.

### What I learned

- The first serious pilot should always include a real dry-run against the actual GitOps repo checkout, not just unit tests or example fixtures.
- `smailnail` is a stronger pilot than simpler apps because it forces the platform model to account for:
  - web auth
  - MCP auth
  - shared database bootstrapping
  - encrypted runtime secrets
- Even with a good shared design, one mismatched reusable-workflow input expression can spoil the pilot signal. Small workflow bugs deserve the same scrutiny as Python helper code.
- The source workflow contract benefits from an explicit `image_name` override instead of relying entirely on `github.repository`, because the local repo has multiple remotes and the intended registry image is clearer when stated directly.

### What was tricky to build

- The sharp edge was not the YAML volume; it was deciding the minimum viable K3s package that is honest about runtime needs without dragging in speculative extras. `smailnail` looks like "just a web app" at first glance, but its real runtime contract includes Postgres bootstrapping, OIDC, MCP bearer auth, and encryption-key material. Copying a lighter app package would have produced a deployment that synced under Argo but failed immediately when `smailnaild` started.
- The second sharp edge was validation order. If I had enabled `open_gitops_pr` in `smailnail` before adding the target manifest and before fixing the reusable-workflow input typo, a later failure would have been ambiguous. Building the target package first, then the source config, then the dry-run patch, made it much easier to trust the result.

### What warrants a second pair of eyes

- The exact Vault runtime secret contract for `apps/smailnail/prod/runtime`:
  - `database`
  - `username`
  - `password`
  - `dsn`
  - `encryption_key_id`
  - `encryption_key_base64`
  - `oidc_client_secret`
- The initial K3s hostname and issuer assumptions:
  - `https://smailnail.yolo.scapegoat.dev`
  - `https://auth.yolo.scapegoat.dev/realms/smailnail`
- Whether the initial deployment should also require `SMAILNAILD_MCP_OIDC_AUDIENCE` or `SMAILNAILD_MCP_OIDC_REQUIRED_SCOPES` from day one, or whether that tightening should wait until the `k3s-parallel` Keycloak environment is applied.
- Whether the bootstrap image baseline in `gitops/kustomize/smailnail/deployment.yaml` should stay on `:main` until the first live pilot publish lands, or whether we should replace it immediately with the first immutable tag once the source branch is pushed.

### What should be done in the future

- Push the `smailnail`, `infra-tooling`, and K3s branches.
- Observe one live `publish-image` run in `smailnail` on `main`.
- Confirm that the live run opens a GitOps PR against `wesen/2026-03-27--hetzner-k3s`.
- Create `terraform/keycloak/apps/smailnail/envs/k3s-parallel/` before first production cutover.
- Populate the Vault runtime and image-pull secret paths before applying `gitops/applications/smailnail.yaml` to the cluster.

### Code review instructions

- Start in the shared repo:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml`
- Then inspect the source repo caller contract:
  - `/home/manuel/code/wesen/corporate-headquarters/smailnail/.github/workflows/publish-image.yaml`
  - `/home/manuel/code/wesen/corporate-headquarters/smailnail/deploy/gitops-targets.json`
- Then inspect the GitOps target:
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/deployment.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/runtime-secret.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/kustomize/smailnail/db-bootstrap-job.yaml`
  - `/home/manuel/code/wesen/2026-03-27--hetzner-k3s/gitops/applications/smailnail.yaml`
- Re-run these checks:
  - `kubectl kustomize gitops/kustomize/smailnail`
  - `python3 scripts/gitops/validate_gitops_targets.py /home/manuel/code/wesen/corporate-headquarters/smailnail/deploy/gitops-targets.json`
  - `python3 -m unittest discover -s tests -v`
  - `go generate ./... && go test ./...`

### Technical details

- K3s implementation commit:
  - `4562c9e feat(smailnail): add k3s gitops manifests`
- Source-repo implementation commit:
  - `7a0cd58 ci: enable smailnail gitops handoff`
- Shared-workflow fix commit:
  - `3b6b7a6 fix(workflows): honor push_image input`
- The local dry-run manifest patch produced the expected one-line image diff:
  - `gitops/kustomize/smailnail/deployment.yaml`
  - `ghcr.io/go-go-golems/smailnail:main -> ghcr.io/go-go-golems/smailnail:sha-deadbee`
- Local environment remediation command:
  - `cd /home/manuel/code/wesen/corporate-headquarters/smailnail/ui && corepack enable && pnpm install --frozen-lockfile`

## Step 8: Fix the smailnail pilot compilation and container-build drift

Once the pilot branches were pushed, the first live `smailnail` `publish-image` runs finally stopped failing on reusable-workflow wiring and started failing on real repo issues. That was the point where the pilot became useful. The compilation problem turned out to be two separate mismatches that only show up in different contexts: the local/CI `go generate` path assumed frontend dependencies might not exist, while the Docker image build assumed the Go builder image version still matched `go.mod`.

The key outcome from this step was that the repo now passes the real compile/test path locally, the fallback `go generate` path works even when frontend dependencies are absent, and the live GitHub Actions `publish-image` workflow completes successfully through image build and push. The remaining blocker is not compilation anymore. It is that the GitOps PR job intentionally skips because `GITOPS_PR_TOKEN` is missing.

### What I did

- Reproduced the local `smailnail` validation path:
  - `go generate ./...`
  - `go test ./...`
- Re-ran the CI-like fallback generation path with frontend dependencies intentionally removed:
  - move `ui/node_modules` aside
  - remove `ui/dist`
  - run `go generate ./pkg/smailnaild/web`
- Inspected the failed live run logs for `smailnail` run `23910254814`.
- Confirmed the compile/test phase had moved past the earlier TypeScript and `go generate` failures, then isolated the new failure to Docker build stage:
  - `go.mod requires go >= 1.26.1 (running go 1.25.8; GOTOOLCHAIN=local)`
- Updated `/home/manuel/code/wesen/corporate-headquarters/smailnail/Dockerfile` to use:
  - `FROM golang:1.26.1-bookworm AS builder`
- Reproduced the exact Docker path locally with:
  - `docker build -t smailnail-ci-check .`
- Committed and pushed the Dockerfile fix:
  - `f52bf01 fix(docker): align builder go version with module`
- Watched the next live workflow run:
  - `https://github.com/go-go-golems/smailnail/actions/runs/23910489369`

### Why

- A successful `go test ./...` is not enough if the production image build uses a different Go toolchain than the repo declares.
- The pilot needed to prove that the shared workflow is good enough to surface real application issues instead of only workflow wiring mistakes.
- Fixing the Docker builder version in the app repo is the correct boundary. This is app-owned build configuration, not shared infra-tooling logic.

### What worked

- Local `go generate ./...` and `go test ./...` both passed.
- The fallback `go generate ./pkg/smailnaild/web` path also passed with `ui/node_modules` and `ui/dist` missing, reusing the committed embedded assets as intended.
- The local container build completed successfully after the Dockerfile update.
- Live GitHub Actions run `23910489369` completed successfully:
  - `release / publish` passed in about 5 minutes 31 seconds
  - `release / Open GitOps PR` ran and exited successfully

### What didn't work

- The previous live run `23910254814` failed in Docker build because the repo had drifted:
  - `go.mod` required Go `1.26.1`
  - `Dockerfile` still used `golang:1.25.8-bookworm`
- The GitOps PR job on the successful run did not open a PR because the repo does not currently expose `GITOPS_PR_TOKEN`. The job output showed:
  - `Detect GitOps PR token availability`
  - `Skip GitOps PR creation when token is missing`

### What I learned

- The pilot is now at a healthier stage. Shared workflow and action wiring are good enough that the remaining failures are ordinary repo issues rather than framework issues.
- The fastest way to debug these rollouts is to separate three paths explicitly:
  - repo-local `go generate` and tests
  - local Docker build
  - live workflow handoff to GitOps
- The final missing piece for true end-to-end verification is secret provisioning, not more release-engineering code.

### What was tricky to build

- The subtlety here was that there were two "compilation" failures with different scopes:
  - frontend/tooling assumptions during `go generate`
  - Go toolchain mismatch inside Docker build
- It would have been easy to stop after the first local green run and miss the real production failure. The live publish workflow was necessary to expose the Dockerfile drift.

### What warrants a second pair of eyes

- Whether `smailnail` should pin the Go builder version in one more centralized way so `go.mod`, local toolchain, and Dockerfile cannot drift apart again.
- Whether the `publish-image` caller in `smailnail` should fail hard when `open_gitops_pr` is enabled but `GITOPS_PR_TOKEN` is absent, or whether the current skip behavior is the right default for rollout safety.

### What should be done in the future

- Configure `GITOPS_PR_TOKEN` for `go-go-golems/smailnail`.
- Re-run `publish-image` on `main`.
- Confirm that the second job opens a real PR against `wesen/2026-03-27--hetzner-k3s`.
- After that, decide whether to harden the shared workflow so an enabled GitOps handoff without token provisioning fails loudly instead of skipping.

### Code review instructions

- Start with the app build drift fix:
  - `/home/manuel/code/wesen/corporate-headquarters/smailnail/Dockerfile`
- Then review the app repo’s declared toolchain:
  - `/home/manuel/code/wesen/corporate-headquarters/smailnail/go.mod`
- Then inspect the successful live run:
  - `https://github.com/go-go-golems/smailnail/actions/runs/23910489369`
- Finally confirm the remaining handoff gate in the reusable workflow:
  - `/home/manuel/code/wesen/corporate-headquarters/infra-tooling/.github/workflows/publish-ghcr-image.yml`

### Technical details

- Failed live run caused by Docker build drift:
  - `23910254814`
- Successful live run after the fix:
  - `23910489369`
- Docker build failure from the earlier run:
  - `go.mod requires go >= 1.26.1 (running go 1.25.8; GOTOOLCHAIN=local)`
- Local verification commands:
  - `go generate ./...`
  - `go test ./...`
  - `docker build -t smailnail-ci-check .`
- Fix commit:
  - `f52bf01 fix(docker): align builder go version with module`
