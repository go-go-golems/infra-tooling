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
LastUpdated: 2026-04-02T10:27:44.403924503-04:00
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
