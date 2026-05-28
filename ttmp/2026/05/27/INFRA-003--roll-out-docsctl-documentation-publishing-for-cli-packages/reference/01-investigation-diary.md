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
    - Path: ../../../../../../../../corporate-headquarters/loupedeck/.github/workflows/release.yaml
      Note: Step 11 release workflow that published docs
    - Path: ../../../../../../../../terraform/vault/github-actions/envs/k3s/main.tf
      Note: Step 11 applied Terraform docsctl publisher roles
    - Path: pkg/release/release.go
      Note: Step 11 ggg release-tag ergonomics fix
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

## Step 6: Patch the docsctl publishing playbook with rollout lessons

I updated the docsctl publishing playbook before changing more `ggg` code because it contained two pieces of stale or unsafe guidance discovered during the active rollout. The validation examples did not include the now-required `--package` and `--version` flags, and the workflow examples granted `id-token: write` too broadly at workflow root.

The playbook now tells operators to make package identity decisions up front, use `GOWORK=off` for local export validation, validate with package/version identity, scope OIDC permission to the `publish-docs` job, and treat `release.yaml` versus `release.yml` as an exact Vault OIDC claim input.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Execute improvement 1 by patching the docsctl publishing playbook.

**Inferred user intent:** Make the next docsctl rollout safer and less error-prone before more package workflow PRs are created.

**Commit (code):** pending — "Update docsctl publishing playbook"

### What I did
- Updated `docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md`.
- Added a package identity checklist.
- Replaced stale `docsctl validate --file ...` examples with package/version-aware validation.
- Changed workflow templates to use job-level `permissions: contents: read, id-token: write` on `publish-docs`.
- Added an exact workflow filename warning for Vault `workflow_ref` claims.
- Marked the playbook task complete.

### Why
- Codex correctly flagged broad workflow-level OIDC permissions during this rollout.
- The playbook should match the current `docsctl validate` command contract.

### What worked
- The playbook now reflects the implementation pattern used in the amended package PRs.

### What didn't work
- A first multi-edit attempt failed because the export command snippet appeared more than once. I split the edits into smaller unique replacements.

### What I learned
- The playbook had multiple copies of the same export command, so future playbook changes should search for all stale examples after editing.

### What was tricky to build
- The important nuance is that caller workflows can keep top-level `contents: write` for release jobs while granting `id-token: write` only to `publish-docs`.

### What warrants a second pair of eyes
- Confirm the reusable workflow call is valid with job-level permissions on a job that uses another workflow.

### What should be done in the future
- Consider adding a `ggg rollout docsctl patch-workflows` command after inventory/plan commands stabilize.

### Code review instructions
- Review the “Before you start”, Step 1, Step 2, checklist, minimal workflow template, and troubleshooting sections of the docsctl playbook.

## Step 7: Add compact and Markdown batch readiness reports

I added the first `ggg` ergonomics improvement: `ggg batch ready` can now produce either grouped summary rows or a copy/paste-ready Markdown report. This directly supports operator status questions during rollouts, especially when there are many PRs and the default JSON output is too verbose.

The implementation refactors batch readiness into collection and emission phases. The command now collects PR reports once, then emits detailed rows, summary rows, or Markdown from the same in-memory data.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Execute improvement 2 by adding compact and Markdown readiness reporting to `ggg batch ready`.

**Inferred user intent:** Make rollout status reporting faster and easier to paste into tickets, chat, and diary updates.

**Commit (code):** pending — "Add batch readiness reports"

### What I did
- Added flags to `ggg batch ready`:
  - `--summary-only`
  - `--markdown-report`
- Refactored `internal/cli/batch/ready.go` to collect reports before emitting rows.
- Added grouped categories for ready, Codex feedback, failed checks, merge conflicts, waiting checks, waiting Codex, missing Codex, and other states.
- Saved validation artifacts:
  - `sources/07-batch-ready-summary-only.json`
  - `sources/08-batch-ready-markdown-report.txt`

### Why
- The default batch output is precise but noisy.
- Markdown output can be pasted directly into docmgr diaries or handoff comments.

### What worked
- `go test ./...` passed.
- `ggg batch ready ... --summary-only --output json` produced grouped rows.
- `ggg batch ready ... --markdown-report` produced raw Markdown.

### What didn't work
- The live INFRA-003 batch currently exits non-zero because `css-visual-diff` has a failed check and `loupedeck` is waiting on Codex. I captured the reports with `|| true` because that non-zero state is expected for a live rollout.

### What I learned
- Separating report collection from output emission makes future report formats much easier.

### What was tricky to build
- `--markdown-report` needed to be raw stdout rather than another Glazed row to be useful with shell redirection. The implementation prints Markdown directly and emits no Glazed rows for that mode.

### What warrants a second pair of eyes
- Confirm that direct stdout printing inside a Glazed command is acceptable for this command mode.
- Consider whether `--markdown-report` should imply a zero exit code for reporting-only use; currently it preserves readiness exit semantics.

### What should be done in the future
- Add a dedicated `ggg batch report` command if more report formats appear.

### Code review instructions
- Review `internal/cli/batch/ready.go` around `collectBatchReports`, `emitBatchRows`, and `markdownBatchReport`.
- Validate with:
  - `go test ./...`
  - `ggg batch ready <prs.yaml> --summary-only --output json`
  - `ggg batch ready <prs.yaml> --markdown-report`

## Step 8: Add grouped batch Codex comments

I added `ggg batch codex-comments` so rollout operators can inspect Codex feedback across a YAML PR list without running the single-PR command repeatedly. The command can emit individual Codex comments or group them by normalized message title.

This directly addresses the pattern from the docsctl rollout where Codex flagged the same OIDC scoping issue in multiple PRs. Grouping makes repeated feedback visible as one rollout-wide fix instead of isolated PR comments.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Execute improvement 3 by adding grouped Codex feedback reporting for batch PR lists.

**Inferred user intent:** Make repeated Codex feedback across rollout PRs easier to detect and fix systematically.

**Commit (code):** pending — "Add batch Codex comment grouping"

### What I did
- Added `internal/cli/batch/codex_comments.go`.
- Registered `codex-comments` in `internal/cli/batch/root.go`.
- Added flags:
  - `--full-body`
  - `--current-head`
  - `--group-by-message`
- Reused `prlist.Load`, `ghclient.Client{}.Snapshot`, and `prready.SortedSignals`.
- Saved validation output in `sources/09-batch-codex-comments-grouped.json`.

### Why
- Rollout PRs often receive identical or closely related Codex feedback.
- The operator should be able to group repeated findings before patching only one repo.

### What worked
- `go test ./...` passed.
- `ggg batch codex-comments <prs.yaml> --group-by-message --output json` emitted grouped current-head Codex signals for the live INFRA-003 PRs.

### What didn't work
- The current live PR set mostly has satisfied Codex messages, so this validation did not show repeated actionable feedback. The command path is still validated against real PR data.

### What I learned
- Codex satisfied messages vary in wording, so exact title grouping is most useful for structured review comments rather than satisfied top-level comments.

### What was tricky to build
- Codex messages include HTML, badges, Markdown, and generated boilerplate. The title normalization strips common markup but intentionally stays conservative.

### What warrants a second pair of eyes
- The grouping key may need refinement after seeing more inline review comments with badges and severity labels.

### What should be done in the future
- Add tests around `codexMessageTitle` normalization with real Codex feedback samples.

### Code review instructions
- Review `internal/cli/batch/codex_comments.go` and `internal/cli/batch/root.go`.
- Validate with:
  - `go test ./...`
  - `ggg batch codex-comments <prs.yaml> --group-by-message --output json`

## Step 9: Add the `ggg rollout docsctl` profile

I added a profile-specific `ggg rollout docsctl` command group with `inventory`, `validate`, and `plan` subcommands. This moves the ad-hoc ticket scripts into reusable tooling while keeping the first version intentionally read-only: it discovers and validates candidates, but it does not patch workflows or Terraform yet.

The command supports package and command overrides so multi-CLI repositories can encode explicit product decisions such as `go-go-goja=./cmd/goja-repl` rather than publishing the first validating binary by accident.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Execute improvement 4 by adding docsctl inventory, validation, and planning support to `ggg rollout`.

**Inferred user intent:** Make future docsctl publishing rollouts repeatable without writing ticket-local discovery scripts.

**Commit (code):** pending — "Add docsctl rollout profile"

### What I did
- Added `internal/cli/rollout/docsctl.go`.
- Registered `docsctl` in `internal/cli/rollout/root.go`.
- Added subcommands:
  - `ggg rollout docsctl inventory`
  - `ggg rollout docsctl validate`
  - `ggg rollout docsctl plan`
- Added flags:
  - `--workspace`
  - `--include`
  - `--exclude`
  - `--package repo=package`
  - `--cmd repo=./cmd/name`
  - `--output table|json|yaml`
  - `--timeout`
  - `--version`
- Saved validation artifacts:
  - `sources/10-ggg-docsctl-inventory-css.yaml`
  - `sources/11-ggg-docsctl-validate-css.yaml`
  - `sources/12-ggg-docsctl-plan-css.yaml`
  - `sources/13-ggg-docsctl-plan-go-go-goja.yaml`

### Why
- The rollout needs repeatable candidate discovery and local validation.
- Multi-CLI repositories need explicit command/package overrides to avoid wrong public docs surfaces.

### What worked
- `go test ./...` passed.
- `ggg rollout docsctl inventory --include css-visual-diff` found both `cmd/build-web` and `cmd/css-visual-diff`.
- `ggg rollout docsctl validate --include css-visual-diff` correctly reported `cmd/build-web` as `export_failed` and `cmd/css-visual-diff` as `validate_ok`.
- `ggg rollout docsctl plan --include go-go-goja --cmd go-go-goja=./cmd/goja-repl` produced a plan for the approved `goja-repl` export.

### What didn't work
- The first version lacked command/package overrides, which would have made `go-go-goja` ambiguous. I added `--cmd` and `--package` before committing.

### What I learned
- `inventory` should intentionally include invalid candidates such as build helpers; `validate` and `plan` are where invalid commands are filtered out.

### What was tricky to build
- Validation shells out to `go run` and `docsctl`, so it needs per-command timeouts and temporary SQLite output paths. The command removes temporary export directories after each candidate.

### What warrants a second pair of eyes
- The command currently defaults `package_name` to the repo basename. That is correct for the current rollout but may need config-file input for repos whose public docs package intentionally differs.
- The command prints direct table/JSON/YAML rather than using Glazed rows. This keeps YAML plan output simple but differs from other `ggg rollout` commands.

### What should be done in the future
- Add workflow/Terraform patch subcommands once the plan format has been used once or twice.
- Add tests for `assignmentMap`, workflow detection, and candidate filtering.

### Code review instructions
- Review `internal/cli/rollout/docsctl.go` and `internal/cli/rollout/root.go`.
- Validate with:
  - `go test ./...`
  - `ggg rollout docsctl inventory --workspace <workspace> --include css-visual-diff --output yaml`
  - `ggg rollout docsctl validate --workspace <workspace> --include css-visual-diff --output yaml`
  - `ggg rollout docsctl plan --workspace <workspace> --include go-go-goja --cmd go-go-goja=./cmd/goja-repl --output yaml`

## Step 10: Final validation and installation of improved `ggg`

I ran the final validation pass for the implemented playbook and `ggg` improvements, installed the updated `ggg` binary, related the key implementation files back to the guide, and fixed a docmgr hygiene issue caused by storing a generated Markdown report as a `.md` file without frontmatter.

The final state is that all four requested improvement areas are implemented: playbook patch, batch readiness reports, grouped batch Codex comments, and docsctl rollout inventory/validate/plan commands.

### Prompt Context

**User prompt (verbatim):** (see Step 5)

**Assistant interpretation:** Finish execution, validate, and keep documentation/bookkeeping current.

**Inferred user intent:** Leave the repo in a reviewed, committed, continuation-friendly state.

**Commit (code):** pending — final bookkeeping commit after implementation commits.

### What I did
- Ran `go test ./...` successfully.
- Built and installed `~/.local/bin/ggg`.
- Related implementation files to the automation guide with `docmgr doc relate`.
- Ran `docmgr doctor --ticket INFRA-003 --stale-after 30`.
- Renamed the generated Markdown report artifact from `.md` to `.txt` because docmgr treats `.md` files as managed docs requiring frontmatter.
- Marked the final Phase 5 bookkeeping task complete.

### Why
- The installed `ggg` should match the committed implementation.
- Ticket artifacts should pass docmgr validation.

### What worked
- `go test ./...` passed.
- `docmgr doctor --ticket INFRA-003 --stale-after 30` passed after renaming the generated Markdown report artifact.

### What didn't work
- Doctor initially failed on `sources/08-batch-ready-markdown-report.md` with:
  - `frontmatter delimiters '---' not found`
- Fixed by renaming it to `sources/08-batch-ready-markdown-report.txt` and updating references.

### What I learned
- Generated Markdown snippets stored under ticket `sources/` should use `.txt` unless they are intended to be managed docmgr documents with frontmatter.

### What was tricky to build
- The code changes were straightforward; the main sharp edge was docmgr treating every `.md` under the ticket as a frontmatter-bearing document.

### What warrants a second pair of eyes
- Review the direct-output CLI choices:
  - `batch ready --markdown-report` prints raw Markdown directly.
  - `rollout docsctl` prints direct table/JSON/YAML rather than Glazed rows.

### What should be done in the future
- Add tests for the docsctl rollout profile and Codex title normalization.
- Consider adding workflow/Terraform patching after the read-only docsctl profile is used in another rollout.

### Code review instructions
- Run `go test ./...`.
- Try:
  - `ggg batch ready <prs.yaml> --summary-only --output json`
  - `ggg batch ready <prs.yaml> --markdown-report`
  - `ggg batch codex-comments <prs.yaml> --group-by-message --output json`
  - `ggg rollout docsctl plan --workspace <workspace> --include go-go-goja --cmd go-go-goja=./cmd/goja-repl --output yaml`

## Step 11: Apply Vault roles and verify the first docsctl release

I applied the merged Terraform/Vault publisher-role change, tagged `go-go-golems/loupedeck` as `v0.1.1`, watched the release workflow, and verified that the new docsctl publish job uploaded documentation to the live docs browser.

This was the first end-to-end proof that the rollout works after the Terraform-first sequencing: GitHub tag push, GoReleaser jobs, reusable docsctl workflow, Vault OIDC login, package-specific publish token minting, docs-registry upload, and public docs browsing all succeeded.

### Prompt Context

**User prompt (verbatim):** "apply terraform, tag loupedeck, check that it works once done. Keep detailed diary, esp. wrt improving ggg as needed"

**Assistant interpretation:** Apply the already-merged Terraform roles, release/tag Loupedeck, verify docs publishing, and record operational lessons plus any ggg improvements.

**Inferred user intent:** Prove the docsctl rollout is actually deployable, not just merged, and improve the automation where the live operator flow exposes sharp edges.

**Commit (code):** pending — "Avoid detached checkout when tagging releases"

### What I did
- Applied Terraform from `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s` with:
  - `AWS_PROFILE=manuel`
  - `VAULT_ADDR=https://vault.yolo.scapegoat.dev`
  - `VAULT_TOKEN` loaded from `~/.vault-token`
- Planned and applied the merged docsctl publisher roles:
  - plan: `18 to add, 1 to change, 0 to destroy`
  - apply: `Apply complete! Resources: 18 added, 1 changed, 0 destroyed.`
- Tagged Loupedeck with:
  - `ggg release tag-patch --yes --output json`
  - created `v0.1.1` at `bca2bea3ecb6f90fe0de74a2822c26510a8ccbf0`
  - proxy verification succeeded for `github.com/go-go-golems/loupedeck@v0.1.1`
- Watched release run:
  - `https://github.com/go-go-golems/loupedeck/actions/runs/26543228613`
- Verified docs publish job succeeded.
- Verified public docs at:
  - `https://docs.yolo.scapegoat.dev/loupedeck/v0.1.1`
- Saved the release verification note in:
  - `sources/14-loupedeck-release-v0.1.1-verification.txt`
- Improved `ggg release tag-*` so future tag commands do not leave the repo in detached HEAD state.
- Ran a post-apply Terraform drift check; `terraform plan -detailed-exitcode` exited `0` with `No changes`.

### Why
- Merging the workflow PR does not publish docs by itself; release-tag workflows publish docs.
- The Vault roles must exist before tag workflows can authenticate to publish package docs.

### What worked
- Terraform backend and Vault access worked once `AWS_PROFILE=manuel` and `VAULT_TOKEN` were explicitly exported.
- The release workflow succeeded end-to-end.
- The `Publish docs / publish-docs` reusable workflow job succeeded, including Vault login, JWT minting, publishing, and package API verification.
- The public docs page has title `Loupedeck v0.1.1 Documentation — Glazed Help Browser` and lists 4 sections.
- A post-apply Terraform plan returned exit code `0`, confirming the applied Vault state matches configuration.

### What didn't work
- Running Terraform earlier without AWS credentials failed with:
  - `Error: No valid credential sources found`
- The `ggg release tag-patch` implementation tagged correctly but left the local checkout detached because its plan was:
  - `git checkout --detach origin/main`
  - `git tag v0.1.1`
- This is bad operator ergonomics in repos that also have active worktrees/branches.

### What I learned
- The docsctl reusable workflow's internal verification is useful: the job reached `Verify package appears in docs API` and completed successfully before I did manual browser/curl verification.
- Public docs can be validated from the preloaded HTML state even when the exact API endpoint shape is not obvious. The page HTML included package `loupedeck`, version `v0.1.1`, and section count `4`.

### What was tricky to build
- The live release path crosses three separate systems: GitHub release workflow, Vault OIDC roles, and docs-registry. A failure in any one looks like a release failure unless the operator can quickly see which job/step failed.
- `ggg release tag-patch` did not fail, so the detached-checkout issue was an ergonomics bug rather than a correctness bug. I fixed it by changing the tag operation from checkout-then-tag to direct target tagging: `git tag <tag> <commit>`.

### What warrants a second pair of eyes
- The Terraform apply created publisher roles for all planned packages and corrected Pinocchio's workflow filename. Review the apply scope against the merged Terraform PR if audit traceability is needed.
- The docs page is live, but we should still validate one Markdown alternate URL and one section URL before tagging the rest of the packages.

### What should be done in the future
- Add a `ggg release watch` or `ggg release verify-docs` command that can wait for a tag run, identify the docs job, and test the public docs URL.
- Teach `ggg release tag-*` to optionally emit the GitHub Actions run URL after pushing a tag.
- Add docs URL/API endpoint discovery to the docsctl rollout playbook so operators do not have to guess API paths.

### Code review instructions
- Review `pkg/release/release.go`.
- Validate with:
  - `go test ./...`
  - `ggg release tag-patch --dry-run --yes --output json` in a clean repo; confirm the plan no longer includes `git checkout --detach`. Validation artifact: `sources/15-loupedeck-tag-dryrun-after-ggg-fix.json`.

### Technical details
- Terraform apply command shape:
  - `terraform plan -out=/tmp/infra-003-docsctl-publishers-apply.tfplan`
  - `terraform apply -auto-approve /tmp/infra-003-docsctl-publishers-apply.tfplan`
- Loupedeck release command:
  - `ggg release tag-patch --yes --output json`
- Release run:
  - `https://github.com/go-go-golems/loupedeck/actions/runs/26543228613`
- Public docs URL:
  - `https://docs.yolo.scapegoat.dev/loupedeck/v0.1.1`

## Step 12: Add release watch and docs verification commands

I added first-class release verification commands to `ggg` after the Loupedeck release proved that the operator needed a single path from tag workflow to public docs URL. The new commands are intentionally small: one watches a tag-triggered GitHub Actions release run, and one verifies that the docs browser has a visible package/version page.

This converts the manual curl/gh sequence from Step 11 into repeatable CLI operations that can be used for the remaining package tags.

### Prompt Context

**User prompt (verbatim):** "alright, do the ggg improvements, update the playbooks, and tag more packages for verification."

**Assistant interpretation:** Implement the release/docs verification improvements discovered during Loupedeck, document them in playbooks, and use them while tagging additional packages.

**Inferred user intent:** Reduce manual rollout risk before tagging the rest of the docsctl-enabled package releases.

**Commit (code):** pending — "Add release docs verification commands"

### What I did
- Added `ggg release verify-docs` in `internal/cli/release/verify_docs.go`.
- Added `ggg release watch` in `internal/cli/release/watch.go`.
- Registered both commands in `internal/cli/release/root.go`.
- Validated with:
  - `go test ./...`
  - `go build -o ~/.local/bin/ggg ./cmd/ggg`
  - `ggg release verify-docs --package loupedeck --version v0.1.1 --output json`
  - `ggg release watch --repo go-go-golems/loupedeck --tag v0.1.1 --no-stream --verify-docs --package loupedeck --output json`
- Saved validation artifacts:
  - `sources/17-ggg-release-verify-docs-loupedeck.json`
  - `sources/18-ggg-release-watch-loupedeck.json`

### Why
- Manual release verification involved several commands and a guessed docs URL/API shape.
- Operators need a clear pass/fail result for “did this tag publish docs successfully?”

### What worked
- `verify-docs` found the live Loupedeck `v0.1.1` page and reported 4 sections.
- `watch --no-stream --verify-docs` found the completed Loupedeck release run and verified the docs page.

### What didn't work
- The first docs verifier counted the first `sectionCount` in the page preload, which belonged to the default `glazed` package. I fixed it to match the requested package object before reading `sectionCount`.

### What I learned
- The docs browser embeds enough preloaded state for simple verification without depending on a private or unstable API path.

### What was tricky to build
- `gh run watch` streams human-readable output, while rollout documentation needs a final structured result. `ggg release watch` supports both: streaming through `gh` by default, then querying the final run JSON and optionally verifying docs.

### What warrants a second pair of eyes
- The docs verifier uses HTML/preloaded-state string matching rather than a typed docs API. It is pragmatic and works today, but a stable docs API endpoint would be cleaner.

### What should be done in the future
- Add a higher-level `ggg release tag-and-watch --verify-docs` flow once these subcommands have stabilized.
- Add tests for package-specific `sectionCount` extraction.

### Code review instructions
- Review `internal/cli/release/watch.go`, `internal/cli/release/verify_docs.go`, and `internal/cli/release/root.go`.
- Validate with the Loupedeck v0.1.1 commands above.

## Step 13: Update playbooks for release/docs verification

I updated the docsctl rollout and package release-train playbooks to use the new `ggg release watch` and `ggg release verify-docs` commands. The playbooks now describe the tag-first docs publishing model and give operators a structured path for saving verification artifacts.

This closes the documentation loop from the Loupedeck proof: after Terraform is applied, a package release tag should be watched with `ggg`, and the public docs URL should be verified with `ggg` instead of hand-written curl probes.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Update operational docs to reflect the new release/docs verification commands.

**Inferred user intent:** Make the improved workflow reusable for the next package tags and future operators.

**Commit (code):** pending — "Document release docs verification workflow"

### What I did
- Updated `docs/go-go-golems/playbooks/docsctl-docs-publishing-rollout-playbook.md`.
- Updated `docs/go-go-golems/package-publishing-release-train.md`.
- Replaced hand-pushed tag examples with `ggg release tag-*` guidance.
- Added `ggg release watch --verify-docs` examples.
- Added `ggg release verify-docs` examples for post-release proof.

### Why
- The previous playbook still implied manual `gh run watch` and curl checks were the primary path.
- The docsctl rollout now has a tool-supported verification workflow.

### What worked
- The playbooks now align with the actual Loupedeck release path.

### What didn't work
- The first edit collapsed multiline shell examples into single-line commands; I corrected the examples back to readable continuation-style snippets.

### What I learned
- Documentation examples for release operations should bias toward copy/paste-friendly multiline commands because operators often paste them into ticket scripts.

### What was tricky to build
- The playbook has both docsctl-specific and general release-train sections; both needed updates so operators do not learn two different release verification flows.

### What warrants a second pair of eyes
- Confirm whether the docsctl playbook should still mention the older `/api/packages` manual check now that `ggg release verify-docs` checks the public browser page.

### What should be done in the future
- Add a short example showing `--no-stream --output json | tee sources/...` for docmgr ticket evidence.

### Code review instructions
- Review both playbook diffs and ensure commands are copy/pasteable.

## Step 14: Tag additional packages and fix no-stream release watch

I tagged three more docsctl-enabled packages and used the new `ggg release watch --verify-docs` flow to verify that they published documentation. During this, I found and fixed a `ggg release watch --no-stream` bug: it located the run but returned immediately while the run was still in progress instead of polling to completion.

After the fix, the command successfully waited for completed release runs and verified public docs pages for `discord-bot`, `go-minitrace`, and `workspace-manager`.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Use the improved release tooling to tag more package releases and verify docs publishing end-to-end.

**Inferred user intent:** Prove the docsctl rollout works across multiple repositories, not just Loupedeck.

**Commit (code):** pending — "Poll release completion in no-stream mode"

### What I did
- Tagged `discord-bot` from `v0.1.5` to `v0.1.6`.
- Tagged `go-minitrace` from `v0.0.16` to `v0.0.17`.
- Tagged `workspace-manager` from `v0.0.4` to `v0.0.5`.
- Verified Go module proxy publication for all three through `ggg release tag-patch`.
- Watched and verified docs for all three releases with `ggg release watch --no-stream --verify-docs`.
- Fixed `ggg release watch --no-stream` so it polls until the run reaches `completed`.
- Saved artifacts:
  - `sources/19-release-watch-discord-bot-v0.1.6.json`
  - `sources/20-release-watch-go-minitrace-v0.0.17.json`
  - `sources/21-release-watch-workspace-manager-v0.0.5.json`
  - `sources/22-additional-release-tags-summary.txt`

### Why
- The first Loupedeck release proved the path once; multiple packages prove that the Terraform roles and reusable workflow pattern generalize.
- `--no-stream` is important for scripts and docmgr ticket artifacts, so it must wait to a terminal state rather than snapshotting an in-progress run.

### What worked
- `discord-bot v0.1.6` published docs with 3 sections.
- `go-minitrace v0.0.17` published docs with 20 sections.
- `workspace-manager v0.0.5` published docs with 6 sections.
- All three release runs completed successfully.

### What didn't work
- Initial `ggg release watch --no-stream --verify-docs` calls returned immediately with `ok=false` because the release run was still `in_progress`.
- Root cause: `watchRelease` only streamed to completion in interactive mode; no-stream mode fetched the run once and evaluated it immediately.
- Fix: added `waitForReleaseRunCompletion` and used it when `--no-stream` is set.

### What I learned
- A non-streaming watch mode is not the same as a non-watching mode. It should suppress live terminal output but still wait for the final state.
- The new verifier is useful enough to catch tool semantics quickly: it made the in-progress behavior obvious in JSON artifacts.

### What was tricky to build
- The release tags were straightforward, but running multiple release watches in parallel exposed that `ggg release watch` needed consistent behavior across streaming and non-streaming modes.

### What warrants a second pair of eyes
- Consider whether `ggg release watch` should expose a separate `--once` flag for current snapshot behavior rather than overloading `--no-stream`.

### What should be done in the future
- Add tests for `waitForReleaseRunCompletion` with a fake `gh` command.
- Add a batch release-watch command for a list of repo/tag/package triples.

### Code review instructions
- Review `internal/cli/release/watch.go`, especially `waitForReleaseRunCompletion`.
- Validate with any completed release tag using:
  - `ggg release watch --repo go-go-golems/<repo> --tag <tag> --no-stream --verify-docs --package <package> --output json`

## Step 15: Verify Pinocchio with `release.yml`

I tagged Pinocchio as an additional verification case because it specifically exercises the workflow filename correction from `release.yaml` to `release.yml` in the Vault bound claims.

The release and docs publish succeeded, proving the corrected `workflow_ref` claim matches the live Pinocchio release workflow.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Continue tagging packages for verification, including a package that validates a known Terraform edge case.

**Inferred user intent:** Increase confidence that both normal `release.yaml` packages and the Pinocchio `release.yml` exception work.

**Commit (code):** N/A — documentation/evidence only

### What I did
- Pulled Pinocchio `main` to include the merged docsctl workflow.
- Tagged `v0.10.29` at `764d25d7a2a18cd30f9afd18b28b8f6547a00e27` using `ggg release tag-patch --yes`.
- Watched the release with:
  - `ggg release watch --repo go-go-golems/pinocchio --workflow release.yml --tag v0.10.29 --verify-docs --package pinocchio --no-stream --output json`
- Saved the verification artifact:
  - `sources/23-release-watch-pinocchio-v0.10.29.json`

### Why
- Pinocchio uses `.github/workflows/release.yml`, not `release.yaml`.
- The Terraform apply changed Pinocchio's Vault bound `workflow_ref` to match `release.yml`; this needed a live tag proof.

### What worked
- The release run succeeded.
- The docs verification succeeded at `https://docs.yolo.scapegoat.dev/pinocchio/v0.10.29`.
- The verifier reported 112 sections.

### What didn't work
- N/A

### What I learned
- The exact workflow filename warning is not theoretical; the successful Pinocchio publish confirms the corrected Vault claim is required and sufficient.

### What was tricky to build
- The command needed an explicit `--workflow release.yml`; the default `release.yaml` is intentionally not correct for Pinocchio.

### What warrants a second pair of eyes
- Pinocchio published 112 sections, whereas earlier validation artifacts showed lower counts for older versions. This may be expected from newer docs exports, but it is worth noting.

### What should be done in the future
- Add workflow filename detection to any future batch release-watch helper so Pinocchio-style exceptions do not require manual `--workflow` overrides.

### Code review instructions
- Inspect `sources/23-release-watch-pinocchio-v0.10.29.json` for the run URL, workflow filename, and docs verification result.

## Step 16: Merge css-visual-diff and capture release failure

I merged the now-ready `css-visual-diff` docsctl PR and attempted to tag `v0.1.3` as another verification package. The tag and Go proxy verification succeeded, but the release workflow failed before docs publishing.

This failure is useful rollout evidence: docsctl was not the failing component. The release failed in the existing GoReleaser darwin build path because `go generate ./...` needed the review UI build, Dagger was unavailable on the macOS runner, and the fallback local `pnpm` executable was missing.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Continue validating package tags and record any failures clearly.

**Inferred user intent:** Discover whether docsctl publishing works in real release workflows and capture actionable failures.

**Commit (code):** N/A — evidence only

### What I did
- Confirmed `css-visual-diff` PR 10 was ready.
- Merged PR 10.
- Pulled `main` in `/home/manuel/code/wesen/go-go-golems/css-visual-diff`.
- Tagged `v0.1.3` with `ggg release tag-patch --yes`.
- Watched the release with `ggg release watch --verify-docs`.
- Saved failure evidence in `sources/24-css-visual-diff-v0.1.3-failed-release.txt`.

### Why
- `css-visual-diff` was the last open docsctl package PR and had become ready after the Chrome startup timeout fix.
- Tagging it would verify the docsctl path for another package if the release workflow succeeded.

### What worked
- The PR merged.
- The tag was created and pushed.
- The Go module proxy verification succeeded for `github.com/go-go-golems/css-visual-diff@v0.1.3`.

### What didn't work
- The release workflow failed before `publish-docs` could run:
  - run: `https://github.com/go-go-golems/css-visual-diff/actions/runs/26544778174`
  - failing job: `goreleaser-darwin`
  - failure: `go generate ./...` tried to build the review UI; Dagger could not start on the macOS runner, then fallback failed with `exec: "pnpm": executable file not found in $PATH`.

### What I learned
- A docsctl verification tag can still fail on package-specific release prerequisites before docs publishing starts.
- `ggg release watch` correctly surfaced the overall release failure, but it does not yet summarize the failed job logs.

### What was tricky to build
- The failed tag is already published as `v0.1.3`, so the fix must land on a later commit and a later tag (`v0.1.4` or similar). Do not retag `v0.1.3`.

### What warrants a second pair of eyes
- Decide whether css-visual-diff should install `pnpm` on macOS release runners or avoid UI generation in the darwin split release job.

### What should be done in the future
- Add `ggg release failed-logs` or extend `ggg release watch` to print the first failed job name and `gh run view --log-failed` hint.
- Fix css-visual-diff release workflow and tag a follow-up version.

### Code review instructions
- Inspect the failed run logs before changing css-visual-diff release workflow.

## Step 17: Add failed-log hints to release watch

I made one more `ggg release watch` improvement based on the css-visual-diff failed release: failed release results now include an exact `gh run view ... --log-failed` command.

This does not fully summarize logs yet, but it shortens the next operator step and makes JSON artifacts self-contained enough to tell a colleague how to inspect the failed job.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Improve `ggg` as rollout failures expose missing operator affordances.

**Inferred user intent:** Make failed release verification actionable without re-discovering GitHub CLI commands manually.

**Commit (code):** pending — "Add failed log hints to release watch"

### What I did
- Added `failed_log_command` to `ggg release watch` results when the watched run completes unsuccessfully.
- Validated against the failed css-visual-diff `v0.1.3` release.
- Saved artifact:
  - `sources/25-css-visual-diff-release-watch-with-failed-log-command.json`

### Why
- The failed css-visual-diff release showed that `ok=false` and a run URL are useful but not enough; the next action is almost always `gh run view <id> --log-failed`.

### What worked
- The failed release result now includes:
  - `gh run view 26544778174 --repo go-go-golems/css-visual-diff --log-failed`

### What didn't work
- N/A

### What I learned
- Failure-oriented output should include the next diagnostic command, not just the state.

### What was tricky to build
- This was small; the main decision was to avoid fetching large logs inside the default structured watch path.

### What warrants a second pair of eyes
- Decide if a future `--include-failed-logs` flag should embed a truncated log excerpt.

### What should be done in the future
- Add failed job names and annotations, not just the `gh run view` command.

### Code review instructions
- Review `internal/cli/release/watch.go` and the failed css-visual-diff artifact.

## Step 18: Patch playbooks with failed release diagnostics

I added a short note to the docsctl and release-train playbooks explaining that `ggg release watch` emits `failed_log_command` when a release fails.

This ties the css-visual-diff failure back into the operator documentation so the next failed release immediately points to the right GitHub CLI diagnostic command.

### Prompt Context

**User prompt (verbatim):** (see Step 12)

**Assistant interpretation:** Keep playbooks aligned with the latest `ggg` release-watch output.

**Inferred user intent:** Ensure operational failures improve the reusable workflow documentation.

**Commit (code):** pending — "Document release watch failed log hints"

### What I did
- Updated `docsctl-docs-publishing-rollout-playbook.md`.
- Updated `package-publishing-release-train.md`.

### Why
- Failed release outputs now include an explicit next diagnostic command; operators should know to use it.

### What worked
- The playbooks now mention `failed_log_command` in the release watch guidance.

### What didn't work
- N/A

### What I learned
- Small failure-output conventions should be documented near the command examples, not only in the CLI help.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- N/A

### What should be done in the future
- Add a richer failed-release triage section if more release failures appear.

### Code review instructions
- Review the two playbook snippets around `ggg release watch`.

## Step 19: Finish css-visual-diff and go-go-goja publication

I continued until the remaining packages had live docs. `css-visual-diff` needed two release workflow fixes before docs could publish; `go-go-goja` needed two GoReleaser configuration fixes before its release workflow could complete.

The final successful docs-published versions are `css-visual-diff v0.1.5` and `go-go-goja v0.6.3`. Earlier tags created during debugging (`css-visual-diff v0.1.3/v0.1.4` and `go-go-goja v0.6.1/v0.6.2`) should be treated as failed release attempts, not docs-published versions.

### Prompt Context

**User prompt (verbatim):** "go ahead. continue. i want them all published."

**Assistant interpretation:** Continue fixing release blockers and tagging follow-up versions until every docsctl rollout package has published docs.

**Inferred user intent:** Finish the rollout, not merely record failures; all target packages should have live docs URLs.

**Commit (code):** N/A in infra-tooling for package fixes; external package commits recorded below.

### What I did
- Fixed `css-visual-diff` release workflow:
  - `2f01366 Install pnpm for darwin release generation`
  - `4d4f668 Install review UI dependencies for darwin release`
- Tagged and verified `css-visual-diff v0.1.5`.
- Fixed `go-go-goja` GoReleaser config:
  - `e61327b Fix GoReleaser binary configuration`
  - `abf0a95 Enable CGO for GoReleaser builds`
- Tagged and verified `go-go-goja v0.6.3`.
- Saved final verification artifacts:
  - `sources/26-release-watch-css-visual-diff-v0.1.5.json`
  - `sources/27-release-watch-go-go-goja-v0.6.3.json`
  - `sources/28-final-docsctl-published-packages-summary.txt`

### Why
- The user explicitly asked for all packages to be published.
- Failed tags do not satisfy docsctl publication because the `publish-docs` jobs never ran.

### What worked
- `css-visual-diff v0.1.5` release succeeded and published 8 sections.
- `go-go-goja v0.6.3` release succeeded and published 22 sections.
- `ggg release watch --no-stream --verify-docs` cleanly verified both final releases.

### What didn't work
- `css-visual-diff v0.1.4` still failed after installing pnpm because `node_modules` had not been installed:
  - `sh: tsc: command not found`
  - `Local package.json exists, but node_modules missing, did you mean to install?`
- `go-go-goja v0.6.1` failed because `.goreleaser.yaml` still used template placeholders:
  - `couldn't find main file: stat cmd/XXX: no such file or directory`
- `go-go-goja v0.6.2` failed because the GoReleaser build disabled CGO while tree-sitter JavaScript bindings require CGO:
  - `github.com/tree-sitter/tree-sitter-javascript/bindings/go: build constraints exclude all Go files`

### What I learned
- Docs publishing often exposes latent release workflow problems because it requires a real tag, not just PR CI.
- `ggg release watch` made repeated release/debug/tag loops much faster, especially with the `failed_log_command` field.

### What was tricky to build
- For `css-visual-diff`, installing pnpm alone was insufficient; the fallback path needed dependencies installed in `web/review-site` before `go generate ./...` ran inside GoReleaser.
- For `go-go-goja`, the release config had stale scaffold placeholders and incorrect CGO settings. The canonical docs export command was `./cmd/goja-repl`, so the release binary was also set to `goja-repl`.

### What warrants a second pair of eyes
- `css-visual-diff` now has two failed tags before the successful tag. Release notes should make clear that `v0.1.5` is the docs-published version.
- `go-go-goja` now has two failed tags before the successful tag. Release notes should make clear that `v0.6.3` is the docs-published version.
- The `go-go-goja` default branch in `/home/manuel/code/wesen/go-go-golems/go-go-goja` is locally unusual (`main` tracks `wesen/main` and is ahead), so I used a clean temporary clone for release fixes against `go-go-golems/go-go-goja`.

### What should be done in the future
- Add preflight release config checks to `ggg release tag-*` or a new `ggg release preflight` command:
  - detect `.goreleaser.yaml` placeholders such as `XXX`;
  - detect CGO-disabled builds when dependencies require CGO;
  - detect frontend generation hooks that require pnpm install.
- Add batch release verification from a YAML list of repo/tag/package triples.

### Code review instructions
- Review package commits:
  - css-visual-diff `2f01366`, `4d4f668`
  - go-go-goja `e61327b`, `abf0a95`
- Review final verification artifacts in `sources/26-*`, `sources/27-*`, and `sources/28-*`.

## Step 20: Add release preflight and amend the rollout automation article

I added `ggg release preflight`, documented it in the release playbooks, and amended the Obsidian article about `ggg` rollout automation with the INFRA-003 release/docs publishing lessons.

This command is the direct follow-up to the failed `css-visual-diff` and `go-go-goja` tags: it checks the repository before a durable public tag is created and reports the kinds of mistakes that caused those failed releases.

### Prompt Context

**User prompt (verbatim):** "add it, continue"

**Assistant interpretation:** Implement the release preflight command suggested by the final publication report, update durable docs, and continue hardening the rollout workflow.

**Inferred user intent:** Prevent future rollout releases from repeating the failed-tag problems found during INFRA-003.

**Commit (code):** pending — "Add release preflight checks"

### What I did
- Added `internal/cli/release/preflight.go`.
- Added `internal/cli/release/preflight_test.go`.
- Registered `ggg release preflight` in `internal/cli/release/root.go`.
- Updated docsctl and package release-train playbooks to run preflight before tagging.
- Amended the Obsidian article:
  - `/home/manuel/code/wesen/go-go-golems/go-go-parc/Projects/2026/05/27/ARTICLE - ggg Rollout Automation - Real-World Testing and Implementation.md`
- Saved preflight validation artifacts:
  - `sources/29-release-preflight-css-visual-diff.json`
  - `sources/30-release-preflight-go-go-goja.json`

### Why
- Failed tags are permanent public release history.
- The rollout exposed predictable failure classes: stale GoReleaser placeholders, missing command paths, CGO-disabled tree-sitter builds, and frontend generation without pnpm dependencies.

### What worked
- `go test ./...` passes with new preflight tests.
- `css-visual-diff` now preflights cleanly.
- `go-go-goja` reports warnings for frontend/pnpm heuristics but no blocking errors after the release fixes.

### What didn't work
- The first main-path regex did not match YAML list items such as `- main: ./cmd/XXX`; the test caught this and I fixed the regex to allow an optional list dash.

### What I learned
- Release preflight should classify some findings as warnings because static heuristics cannot always know which frontend package a `go generate` hook actually uses.

### What was tricky to build
- The command has to be useful without becoming a full GoReleaser interpreter. It intentionally uses targeted, evidence-backed checks based on real rollout failures.

### What warrants a second pair of eyes
- The frontend/pnpm heuristic may be noisy in repositories with unrelated web directories. Review whether it should inspect `go:generate` files more deeply before warning.

### What should be done in the future
- Add a batch release preflight command for repo lists.
- Add richer GoReleaser parsing if more release config failure modes appear.

### Code review instructions
- Review `internal/cli/release/preflight.go` and `preflight_test.go`.
- Run:
  - `go test ./...`
  - `ggg release preflight --repo <repo> --output json`

## Step 21: Run release preflight across all published packages

I ran the new release preflight across the packages that were published during INFRA-003 and saved the combined output as another ticket artifact.

The goal was not to block anything after the fact, but to see how noisy the first preflight heuristic is across the real rollout set. This gives us evidence for which warnings should remain warnings and which checks are accurate enough to become strict in future automation.

### Prompt Context

**User prompt (verbatim):** (same as Step 20)

**Assistant interpretation:** Continue hardening after adding preflight by validating it across the real package set.

**Inferred user intent:** Make sure the new `ggg` command is grounded in the actual repos it is meant to protect.

**Commit (code):** pending — evidence-only follow-up

### What I did
- Ran `ggg release preflight --output json` against:
  - `css-visual-diff`
  - `discord-bot`
  - `go-go-goja`
  - `go-minitrace`
  - `loupedeck`
  - `pinocchio`
  - `workspace-manager`
- Saved output in `sources/31-release-preflight-all-published-packages.jsonl`.

### Why
- A command born from two release failures should be checked against the successful release set before relying on it in the next rollout.

### What worked
- The command runs across all published packages and catches no blocking errors in the fixed release states.

### What didn't work
- `go-go-goja` still emits frontend/pnpm warnings because the heuristic sees a frontend package and `go generate ./...`. This warning did not prevent the successful `v0.6.3` release, so it should remain a warning for now.

### What I learned
- Preflight needs a distinction between “known release blocker” and “operator attention signal.” The current `--strict` flag is the right escape hatch for rollout phases where warnings should be treated as hard gates.

### What was tricky to build
- N/A

### What warrants a second pair of eyes
- Review the combined preflight artifact and decide whether the frontend heuristic should inspect exact `go:generate` targets instead of repository-wide frontend directories.

### What should be done in the future
- Add a YAML batch input to `ggg release preflight` so a rollout can preflight all packages with one structured summary.

### Code review instructions
- Inspect `sources/31-release-preflight-all-published-packages.jsonl`.
