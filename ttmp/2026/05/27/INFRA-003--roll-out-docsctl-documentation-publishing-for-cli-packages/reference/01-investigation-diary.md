---
Title: Investigation Diary
Ticket: INFRA-003
Status: active
Topics:
    - cli
    - automation
    - release
    - github
    - docsctl
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/design-doc/01-docsctl-publishing-rollout-analysis-and-implementation-guide.md
      Note: Primary guide created from the investigation
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/scripts/01-inventory-help-export.sh
      Note: Ticket-local inventory script for CLI help SQLite exports
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/scripts/02-validate-exported-sqlite.sh
      Note: Ticket-local docsctl validation script
ExternalSources: []
Summary: Chronological diary for the docsctl publishing rollout analysis and implementation guide.
LastUpdated: 2026-05-27T17:45:00-04:00
WhatFor: Record what was investigated, what worked, what failed, and how to continue the docsctl publishing rollout.
WhenToUse: Before resuming INFRA-003 implementation work or reviewing the generated design guide.
---


# Diary

## Goal

This diary records the investigation that created the INFRA-003 docsctl publishing rollout guide: ticket setup, evidence gathering, package inventory, validation, guide authoring, docmgr bookkeeping, and reMarkable delivery.

## Step 1: Create the ticket and map the docsctl publishing system

I created a dedicated docmgr ticket because docsctl publishing is a multi-repository rollout that touches package release workflows, Terraform/Vault roles, the public docs registry, and the `ggg` PR/release workflow. The first step was to read the existing `infra-tooling` playbook and reusable workflow so the guide would be anchored in current implementation rather than inferred from memory.

The initial evidence showed that the target rollout shape already exists: package CLIs export Glazed help SQLite, release workflows call `publish-docsctl.yml`, GitHub authenticates to Vault through OIDC, Vault mints a package-scoped publish JWT, and `docsctl publish` uploads to the registry.

### Prompt Context

**User prompt (verbatim):** "i merged them all. Look at the infra-tooling playbooks, can you now handle adding docsctl docs publishing for the packages that have a CLI tool capable of exporting a help sqlite? 

Create a new docmgr ticket for that, and then Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable."

**Assistant interpretation:** Create a new ticket for a future docsctl documentation publishing rollout, inventory package readiness, write a detailed intern-friendly design/implementation guide with evidence and diagrams, store it in docmgr, and upload it to reMarkable.

**Inferred user intent:** The user wants a careful planning and onboarding deliverable before implementing docs publishing across packages, so a new contributor can understand the architecture and execute the rollout safely.

**Commit (code):** N/A — documentation and ticket evidence only.

### What I did
- Created ticket `INFRA-003 — Roll out docsctl documentation publishing for CLI packages`.
- Created the design document `design-doc/01-docsctl-publishing-rollout-analysis-and-implementation-guide.md`.
- Created this diary document.
- Read `docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md`.
- Read `.github/workflows/publish-docsctl.yml`.
- Inspected Terraform/Vault publisher role examples in `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s/main.tf` and `variables.tf`.
- Inspected `docsctl validate` and publish implementation in the `glazed` workspace.

### Why
- The rollout is security-sensitive because package workflows receive publishing credentials through Vault.
- A new intern needs to understand both the happy-path YAML edits and why package identity, workflow refs, and GitHub OIDC claims must line up exactly.

### What worked
- The existing playbook and reusable workflow already provide a strong target architecture.
- Terraform already has examples for `glazed`, `pinocchio`, `remarquee`, and `sqleton`, so new packages can follow the same `local.docsctl_publishers` pattern.

### What didn't work
- The current playbook examples still show `docsctl validate --file .docsctl/help.sqlite`, but the current `docsctl validate` command requires `--package`, `--version`, and `--file`. This became a guide note and a recommended playbook fix.

### What I learned
- `docsctl validate` is stricter than the playbook examples imply, which is good for publishing safety but important for operator commands.
- The reusable workflow already validates and publishes with package/version identity, so package workflow edits should stay small.

### What was tricky to build
- The main ambiguity was scope: the user asked whether we can handle adding publishing, but also specifically requested a ticket, analysis/design/implementation guide, and reMarkable upload. I treated this turn as planning and documentation, not as editing every downstream package immediately.

### What warrants a second pair of eyes
- The package identity decisions for `workspace-manager` and `go-go-goja` should be reviewed before Terraform roles are created.
- The current default branches should be checked again immediately before implementation because the local workspace branches still reflected recent rollout branches in some repos.

### What should be done in the future
- Patch the docsctl publishing playbook validation commands to include `--package` and `--version`.
- Implement the rollout for the Tier 1 packages after approving the package set.

### Code review instructions
- Start with the design doc executive summary and candidate classification.
- Verify the source evidence against:
  - `docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md`
  - `.github/workflows/publish-docsctl.yml`
  - Terraform `main.tf`
  - `glazed/cmd/docsctl/validate.go`
- Validate docmgr with `docmgr doctor --ticket INFRA-003 --stale-after 30`.

### Technical details
- Existing docs API packages were captured from `https://docs.yolo.scapegoat.dev/api/packages`.
- Repository IDs were captured via GitHub GraphQL because this installed `gh` version does not expose `databaseId` through `gh repo view --json`.

## Step 2: Inventory help export capability and write the implementation guide

I wrote ticket-local scripts to test the active workspace repositories for `help export --format sqlite` support. The scripts store their output under the ticket's `sources/` directory so future implementers can inspect raw logs, not just the summary table in the guide.

The inventory found clear ready candidates (`css-visual-diff`, `discord-bot`, `go-minitrace`, `loupedeck`) plus two decision-needed cases (`workspace-manager`, `go-go-goja`). It also confirmed that `glazed` is already live, `pinocchio` is already visible in the docs API, and `pinocchio/cmd/web-chat` should not be published because the exported database has no sections.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Inventory package readiness and produce the requested intern-friendly design and implementation guide.

**Inferred user intent:** Provide an actionable, evidence-backed package docs publishing rollout plan.

**Commit (code):** N/A — documentation and ticket evidence only.

### What I did
- Created `scripts/01-inventory-help-export.sh`.
- Ran the inventory script over `/home/manuel/workspaces/2026-05-24/add-js-providers`.
- Created `scripts/02-validate-exported-sqlite.sh`.
- Installed `docsctl` with `GOWORK=off go install github.com/go-go-golems/glazed/cmd/docsctl@latest`.
- Validated exported SQLite files with explicit `--package` and `--version` flags.
- Captured repository IDs in `sources/repository-ids.tsv`.
- Captured current public docs packages in `sources/current-docs-packages.json`.
- Wrote the full design/implementation guide with diagrams, pseudocode, rollout templates, testing strategy, risks, and open questions.

### Why
- The rollout should only target packages whose CLI can produce a valid, non-empty Glazed help SQLite database.
- The docs registry uses package/version identity as part of the publish route and token policy, so readiness requires more than merely having a `cmd/` directory.

### What worked
- Validated candidates:
  - `css-visual-diff ./cmd/css-visual-diff`
  - `discord-bot ./cmd/discord-bot`
  - `glazed ./cmd/glaze`
  - `go-go-goja ./cmd/goja-jsdoc`
  - `go-go-goja ./cmd/goja-repl`
  - `go-go-goja ./cmd/jsverbs-example`
  - `go-go-goja ./cmd/xgoja`
  - `go-minitrace ./cmd/go-minitrace`
  - `loupedeck ./cmd/loupedeck`
  - `pinocchio ./cmd/pinocchio`
  - `workspace-manager ./cmd/wsm`
- The public docs API currently lists `glazed`, `pinocchio`, `remarquee`, and `sqleton`.

### What didn't work
- Initial validation failed for every exported database because I ran `docsctl validate --file ...` without required `--package` and `--version` flags. Exact error from one log:
  - `required field(s) missing: package, version`
- After updating the validation script to pass package/version, all real candidates validated except `pinocchio ./cmd/web-chat`.
- `pinocchio ./cmd/web-chat` failed validation with:
  - `help database contains no sections`
- `goja-git ./cmd/XXX` exited zero but produced no SQLite output, so it is not a docs publishing candidate.
- Build helper commands such as `cmd/build-web` are not docs CLI candidates even when they exit zero.

### What I learned
- A command can expose a `help` surface without being a meaningful docs publishing target. Validation must require non-empty sections.
- Multi-CLI repositories need a package identity decision before Vault roles are created.
- The installed `gh` version rejected `gh repo view --json databaseId`, so GraphQL was the reliable way to capture numeric repository IDs.

### What was tricky to build
- The inventory had to distinguish three states that all look similar at a glance: export command success, SQLite file existence, and docsctl validation success. A command that exits zero but writes no file is not ready. A command that writes a SQLite file with zero sections is also not ready.
- `go-go-goja` required special care because several tools can export valid docs, but the release configuration does not clearly identify which one is the public package docs surface.

### What warrants a second pair of eyes
- Review the candidate classification before implementing Terraform roles.
- Review whether `workspace-manager` should publish as `workspace-manager` or `wsm`.
- Review whether `go-go-goja` should publish `xgoja` docs as `go-go-goja`, split multiple docs packages, or wait for release config cleanup.

### What should be done in the future
- Use the guide's Tier 1 rollout order for the first implementation PR batch.
- Consider adding a `ggg rollout` profile for docsctl publishing if this rollout repeats across more repositories.

### Code review instructions
- Review `sources/help-export-inventory/summary.txt` and `validation.txt` first.
- Review the design doc's candidate classification and workflow YAML templates.
- Validate with:
  - `PATH="$HOME/go/bin:$PATH" scripts/02-validate-exported-sqlite.sh`
  - `docmgr doctor --ticket INFRA-003 --stale-after 30`

### Technical details
- Inventory script path:
  - `/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/scripts/01-inventory-help-export.sh`
- Validation script path:
  - `/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/scripts/02-validate-exported-sqlite.sh`
- Main guide path:
  - `/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/design-doc/01-docsctl-publishing-rollout-analysis-and-implementation-guide.md`

## Step 3: Validate docmgr hygiene and upload the guide to reMarkable

After writing the guide and diary, I ran docmgr validation and uploaded the ticket bundle to reMarkable. This makes the implementation guide available as a single PDF with table of contents under the dated AI folder.

The only validation issue was an expected vocabulary warning: `docsctl` was not yet a known topic in the infra-tooling docmgr vocabulary. I added the topic and reran doctor successfully before uploading.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete ticket validation and deliver the docsctl publishing guide to reMarkable.

**Inferred user intent:** The user wants the guide stored in docmgr and available on reMarkable for reading/review.

**Commit (code):** N/A — documentation delivery only.

### What I did
- Ran `docmgr doctor --ticket INFRA-003 --stale-after 30`.
- Added docmgr vocabulary topic `docsctl` after doctor reported it as unknown.
- Reran doctor successfully.
- Verified `remarquee status` and `remarquee cloud account --non-interactive`.
- Ran a dry-run bundle upload.
- Uploaded the bundle to `/ai/2026/05/27/INFRA-003`.
- Verified the remote listing.

### Why
- The ticket should pass docmgr validation before handoff.
- reMarkable upload should be dry-run first so formatting and destination are visible before creating the remote document.

### What worked
- `docmgr doctor --ticket INFRA-003 --stale-after 30` passed after adding the vocabulary topic.
- Dry-run upload succeeded.
- Real upload succeeded:
  - `OK: uploaded INFRA-003 docsctl publishing rollout guide.pdf -> /ai/2026/05/27/INFRA-003`
- Remote listing showed:
  - `[f] INFRA-003 docsctl publishing rollout guide`

### What didn't work
- Initial doctor run warned:
  - `unknown_topics — unknown topics: [docsctl]`
- Fixed with:
  - `docmgr vocab add --category topics --slug docsctl --description "docsctl documentation publishing, validation, and docs.yolo registry workflows"`

### What I learned
- The infra-tooling docmgr vocabulary did not yet include docsctl even though the repo already has docsctl playbooks and workflows.

### What was tricky to build
- Nothing in the upload itself was tricky. The important ordering was doctor first, then dry-run upload, then real upload, then remote listing.

### What warrants a second pair of eyes
- Review the PDF on reMarkable for diagram readability, especially the Mermaid diagrams rendered through the Markdown-to-PDF pipeline.

### What should be done in the future
- If the guide is updated after implementation starts, upload a refreshed bundle or a dated follow-up PDF.

### Code review instructions
- Confirm `ttmp/vocabulary.yaml` includes the new `docsctl` topic.
- Confirm the ticket bundle contains index, tasks, design doc, diary, and changelog.
- Validate with `docmgr doctor --ticket INFRA-003 --stale-after 30`.

### Technical details
- reMarkable destination:
  - `/ai/2026/05/27/INFRA-003`
- Bundle name:
  - `INFRA-003 docsctl publishing rollout guide`

## Step 4: Record package identity decisions for workspace-manager and go-go-goja

The user resolved the two package-scope questions that were still open after the initial analysis. `workspace-manager` will publish under the repository/package name `workspace-manager`, and `go-go-goja` will use `goja-repl` as the canonical help export command.

I updated the design guide so the rollout can proceed without ambiguity in Terraform role names, docs URLs, and release workflow `export_command` values.

### Prompt Context

**User prompt (verbatim):** "package name workspace-manager. use goja-repl for the export."

**Assistant interpretation:** Treat `workspace-manager` as the official docs package name and use `./cmd/goja-repl` for `go-go-goja` docs publishing.

**Inferred user intent:** Resolve the remaining rollout decisions so implementation can move directly to Terraform/Vault and release workflow edits.

**Commit (code):** N/A — documentation decision update only.

### What I did
- Updated the executive summary candidate table.
- Updated inventory notes for `workspace-manager` and all valid `go-go-goja` exporters.
- Replaced the Tier 2 “needs decision” section with approved rollout decisions.
- Updated the `go-go-goja` workflow template to use:
  - `GOWORK=off go run ./cmd/goja-repl help export --format sqlite --output-path .docsctl/help.sqlite`
- Marked the two package-scope decision tasks complete.

### Why
- Terraform role names, publish JWT package claims, registry package names, and public docs URLs must be stable before implementation.
- `go-go-goja` has multiple valid help-exporting commands, so the workflow must pin the intended canonical export command.

### What worked
- The guide now has concrete instructions for both previously ambiguous packages.

### What didn't work
- N/A.

### What I learned
- The preferred package identity follows the repository name for `workspace-manager`, while the preferred `go-go-goja` content surface is the REPL docs.

### What was tricky to build
- The key detail is keeping package name and binary/export command separate. `workspace-manager` publishes from binary `wsm`, and `go-go-goja` publishes from binary `goja-repl`.

### What warrants a second pair of eyes
- During implementation, verify that the `go-go-goja` release workflow can run `go run ./cmd/goja-repl` in CI with `GOWORK=off`.

### What should be done in the future
- Implement Terraform/Vault publisher entries and release workflow jobs for the approved candidate set.

### Code review instructions
- Review the guide sections “Candidate classification”, “Phase 4: package-specific workflow templates”, and “Resolved decisions”.
- Confirm the `go-go-goja` workflow template uses `./cmd/goja-repl`, not `./cmd/xgoja`.

### Technical details
- Public docs URLs implied by the decisions:
  - `https://docs.yolo.scapegoat.dev/workspace-manager/<version>`
  - `https://docs.yolo.scapegoat.dev/go-go-goja/<version>`

## Step 5: Plan the next `ggg` and playbook improvements

After the initial package PRs were open, the user asked to turn the top four improvement ideas into tracked work and implementation guidance, then execute them. I created a second design document focused on `ggg` and playbook ergonomics so the implementation has a clear target before code changes begin.

This step also records that the rollout implementation has already progressed beyond the original guide: Terraform and package PRs are open, OIDC permissions were fixed to job-level scope after Codex feedback, and the remaining package PRs are mostly waiting on checks.

### Prompt Context

**User prompt (verbatim):** "add tasks and implementation guide for 1. - 4. and then execute, committing at appropriate intervals, keeping a detailed diary"

**Assistant interpretation:** Add ticket tasks and an implementation guide for the first four recommended improvements, then implement them in focused commits while maintaining the diary.

**Inferred user intent:** The user wants to institutionalize lessons from this rollout immediately, not leave them as informal suggestions.

**Commit (code):** N/A — planning step before implementation commits.

### What I did
- Added `design-doc/02-ggg-docsctl-rollout-automation-implementation-guide.md`.
- Added Phase 5 tasks to `tasks.md` for:
  - docsctl playbook fixes;
  - `ggg batch ready --summary-only` / `--markdown-report`;
  - `ggg batch codex-comments --group-by-message`;
  - `ggg rollout docsctl inventory|validate|plan`.
- Updated Phase 4 tasks to reflect opened Terraform/package PRs and outstanding apply/merge/release steps.

### Why
- The first docsctl rollout already exposed repeated manual work and one security-sensitive playbook gap: OIDC permissions should be job-scoped.
- A guide reduces the chance that code changes drift from the intended operator workflow.

### What worked
- The implementation guide now has concrete CLI UX, pseudocode, phases, and validation commands.

### What didn't work
- N/A.

### What I learned
- The docsctl rollout automation should start with discovery/reporting primitives before attempting workflow/Terraform patch automation.

### What was tricky to build
- Scope control: “1-4” refers to a mixed set of docs and code changes. I split them into explicit phases so each can be committed independently.

### What warrants a second pair of eyes
- Review whether `--markdown-report` should be raw stdout or a Glazed row field. The guide recommends raw stdout for copy/paste usability.

### What should be done in the future
- Execute the Phase 5 tasks in order and update this diary after each commit.

### Code review instructions
- Start with `design-doc/02-ggg-docsctl-rollout-automation-implementation-guide.md`.
- Check `tasks.md` Phase 5 for the implementation checklist.

### Technical details
- The implementation guide references:
  - `internal/cli/batch/ready.go`
  - `internal/cli/pr/codex_comments.go`
  - `internal/cli/rollout/root.go`
  - `docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md`
