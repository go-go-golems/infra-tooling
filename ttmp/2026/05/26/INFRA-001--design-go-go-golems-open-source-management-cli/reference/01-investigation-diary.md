---
Title: Investigation Diary
Ticket: INFRA-001
Status: active
Topics:
  - cli
  - github
  - release
  - automation
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Chronological diary for researching a go-go-golems open-source management CLI.
LastUpdated: 2026-05-26T23:00:00-04:00
WhatFor: Preserve investigation steps, evidence gathered, failures, and continuation notes.
WhenToUse: Read before resuming INFRA-001 research or implementation.
---

# Diary

## Goal

This diary records the research and design work for a future Go package plus CLI toolbox that manages go-go-golems open-source repository operations. The ticket studies the shell/Python scripts, GitHub API interactions, release-train workflows, PR readiness states, and XGOJA-015 operational experience before proposing a cleaner implementation.

## Step 1: Initialize the ticket and gather the first evidence set

I created a dedicated `infra-tooling` docmgr workspace because this work is no longer about one xgoja release train. The goal is to turn repeated operational scripts and lessons into a reusable management tool for future go-go-golems open-source maintenance.

The first investigation pass inventoried current infra-tooling scripts and docs, then explicitly pulled in the ticket-local scripts from XGOJA-015. The user called out those scripts because they contain the real ad-hoc operational knowledge: Codex inline review extraction, PR check summaries, and focused downstream validation commands.

### Prompt Context

**User prompt (verbatim):** "Ok, we are shifting to @infra-tooling/. Create a new docmgr ticket to create a go go golem CLI management tool based on all the scripts and experience we have gathered. First, I want to analyze all the APIs, functionality, actions, data we have been dealing with, the APIs, formats, interactions, etc... ONce we cover all the existing functionality and implementation and usages (and missing things too), we will be able to put together the building blocoks of a go package + CLI toolbox to do go-go-golems opensource management in the future.

Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

Keep a diary as you work"

**Assistant interpretation:** Create a new `infra-tooling` docmgr ticket, perform an evidence-backed analysis of the existing release/readiness scripts and workflows, write a detailed intern-oriented design/implementation guide for a future Go CLI/package, maintain a diary, and upload the result to reMarkable.

**Inferred user intent:** The user wants to consolidate lessons from recent multi-repo release trains into a deliberate tool design instead of accumulating one-off scripts.

**Commit (code):** N/A — documentation/research setup only so far.

### What I did
- Initialized docmgr in `/home/manuel/code/wesen/go-go-golems/infra-tooling` because the repo had no `ttmp` root yet.
- Added vocabulary topics: `cli`, `github`, `release`, and `automation`.
- Created ticket `INFRA-001 — Design go-go-golems open-source management CLI`.
- Created:
  - `design-doc/01-go-go-golems-open-source-management-cli-design.md`
  - `reference/01-investigation-diary.md`
  - `sources/01-line-anchored-evidence.txt`
  - `sources/02-xgoja-015-script-evidence.txt`
- Read current infra-tooling scripts and docs under:
  - `scripts/go-go-golems/`
  - `docs/go-go-golems/`
  - `examples/go-go-golems/`
- Read the XGOJA-015 scripts that the user explicitly referenced:
  - `08-extract-codex-review-comments.sh`
  - `09-pr-check-summary.sh`
  - `10-validate-downstream-focused.sh`
- Replaced the default task list with a concrete investigation/delivery checklist.

### Why
- The future CLI should be based on observed workflows and APIs, not only on desired commands.
- The XGOJA-015 scripts capture missing functionality that is not yet in the reusable infra-tooling script directory.

### What worked
- `docmgr init --seed-vocabulary` created a clean ticket workspace for `infra-tooling`.
- The existing repository already has a useful split between scripts, docs, examples, and action/tooling directories.
- The XGOJA-015 ticket scripts gave concrete examples of functionality that should be promoted into the future CLI.

### What didn't work
- `docmgr status --summary-only` initially failed because `infra-tooling/ttmp` did not exist yet:

```text
Error: root directory does not exist: /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp
```

I fixed that by running `docmgr init --seed-vocabulary`.

### What I learned
- The reusable infra-tooling scripts cover PR readiness and Codex triggering, while XGOJA-015 adds operational helpers for extracting Codex review comments, checking PR checks, and running repo-specific validation.
- The future CLI needs to model both reusable primitives and release-train-specific configuration.

### What was tricky to build
- The evidence spans two repositories: reusable scripts in `infra-tooling` and historical ticket scripts in `go-go-goja`. The ticket sources now preserve line-numbered snapshots of both so the design doc can cite them without relying on memory.

### What warrants a second pair of eyes
- Whether the future CLI should live in `infra-tooling` as a Go module directly, or whether `infra-tooling` should hold docs/scripts while the CLI gets a dedicated repo.
- Whether XGOJA-015 ticket scripts should be copied into `infra-tooling/scripts/go-go-golems/` immediately or only used as design input.

### What should be done in the future
- Write the full design guide with explicit command groups and Go package boundaries.
- Upload the completed docs to reMarkable.

### Code review instructions
- Start with the evidence files in `sources/` and compare them against the current scripts in `scripts/go-go-golems/`.
- Validate docmgr ticket state with `docmgr doctor --ticket INFRA-001 --stale-after 30`.

### Technical details

Evidence capture commands included:

```bash
rg --files scripts docs examples templates tests
find scripts/go-go-golems -maxdepth 2 -type f -print | sort
find docs/go-go-golems -maxdepth 3 -type f -print | sort
nl -ba scripts/go-go-golems/01-pr-ready-check.py
nl -ba /home/manuel/workspaces/2026-05-24/add-js-providers/go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/10-validate-downstream-focused.sh
```

## Step 2: Write the design guide and upload the bundle

The second step turned the evidence inventory into the first full design document. The guide explains the existing scripts, the GitHub and Go module APIs they use, the current release-train state machine, and the proposed Go package/CLI architecture.

I also uploaded the design guide and diary to reMarkable. The first upload attempt timed out at the tool-call level after 300 seconds, but no upload process remained running. Re-running the same upload with a longer timeout succeeded.

### Prompt Context

**User prompt (verbatim):** "look at the scripts you wrote as well, righ. @go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/10-validate-downstream-focused.sh  etc..."

**Assistant interpretation:** Include the ticket-local scripts from the XGOJA-015 release train as first-class evidence, especially `10-validate-downstream-focused.sh` and related helper scripts.

**Inferred user intent:** The user wants the future CLI design to be grounded in the actual scripts used during the release train, not just the reusable infra-tooling scripts.

**Commit (code):** N/A — documentation/research update.

### What I did
- Wrote `design-doc/01-go-go-golems-open-source-management-cli-design.md`.
- Included current functionality inventory for:
  - PR readiness checks;
  - Codex signal parsing;
  - batch readiness/watch behavior;
  - Codex triggering;
  - inline review extraction;
  - PR check summaries;
  - focused downstream validation;
  - release-train dependency and publication workflows;
  - dependency bump Makefile snippets.
- Proposed Go package boundaries, CLI command groups, data models, pseudocode, diagrams, configuration shape, missing functionality, and phased implementation plan.
- Related key files to the design doc with `docmgr doc relate`.
- Updated the changelog and checked completed tasks.
- Uploaded the design bundle to reMarkable:
  - `OK: uploaded INFRA 001 Go Go Golems Management CLI Design.pdf -> /ai/2026/05/26/INFRA-001`

### Why
- The future CLI should consolidate observed operational behavior into typed, testable building blocks.
- The XGOJA-015 scripts show missing reusable commands that should be promoted into the Go CLI design.

### What worked
- Line-anchored evidence files made it easy to cite current scripts and playbooks.
- `docmgr doctor --ticket INFRA-001 --stale-after 30` passed before upload.
- The reMarkable upload succeeded with a longer timeout.

### What didn't work
- The first reMarkable upload attempt timed out after 300 seconds:

```text
Command timed out after 300 seconds
```

- I checked for lingering processes with:

```bash
pgrep -af 'remarquee|pandoc|xelatex' || true
```

No upload/render process was left running, so I reran the same upload with a longer timeout.

### What I learned
- The combined design+diary PDF can take longer than 300 seconds to render/upload in this environment.
- For long technical design bundles, use a longer tool timeout rather than changing the document structure prematurely.

### What was tricky to build
- The design needed to distinguish reusable policy from release-train-specific configuration. For example, `GOWORK=off` validation is a global invariant, but the exact package list for `discord-bot` or `loupedeck` belongs in validation-profile configuration.

### What warrants a second pair of eyes
- Naming of the future CLI (`ggg`, `go-go-golems`, or another binary name).
- Whether the first implementation should shell out to `gh` for parity or use a direct GitHub GraphQL client immediately.
- Whether ticket-local XGOJA-015 scripts should be copied into `infra-tooling/scripts/go-go-golems/` before the Go CLI exists.

### What should be done in the future
- Implement Phase 1: `ggg pr ready <pr> --json` with fixtures matching the current Python behavior.
- Add structured `codexComments` JSON output rather than only message strings.
- Convert `10-validate-downstream-focused.sh` into YAML validation profiles.

### Code review instructions
- Start with the design doc sections “Current functionality inventory”, “Data model for a Go package”, and “Implementation phases”.
- Compare references against:
  - `scripts/go-go-golems/01-pr-ready-check.py`
  - `scripts/go-go-golems/05-batch-pr-ready.sh`
  - XGOJA-015 `scripts/10-validate-downstream-focused.sh`

### Technical details

Successful upload command:

```bash
remarquee upload bundle \
  ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/design-doc/01-go-go-golems-open-source-management-cli-design.md \
  ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli/reference/01-investigation-diary.md \
  --name "INFRA 001 Go Go Golems Management CLI Design" \
  --remote-dir "/ai/2026/05/26/INFRA-001" \
  --toc-depth 2 \
  --non-interactive
```

## Step 3: Update the design for YAML PR lists, force-trigger safety, release verbs, and Glazed output

The third step incorporated the requested design changes and started implementation. The important changes are that PR lists are now YAML instead of ad-hoc newline files, `pr codex-trigger` is safe by default and needs `--force` to trigger over an in-progress Codex run, release verbs include patch/minor/major, and the new CLI uses Glazed commands so output is row-oriented and can become structured data.

I then implemented the initial Go CLI scaffold plus the first functional commands. This is not the full tool yet, but it establishes the command architecture, PR-reference parsing, YAML PR-list parsing, Codex trigger safety check, and release tag verbs.

### Prompt Context

**User prompt (verbatim):** "small changes:

- add --force to pr codex-trigger, because by default, check if a codex run is already running (eyes)
- use yaml for prs.txt
- add tag-minor and tag-major as verbs 

Use glazed commands for all verbs, so that per default we have a concise human output, and with structured out --with-structured-output we get rows oriented structured data.

Update the document, then add a detailed set of phases and for each phases a set of tasks for the work. 

Then work on the tasks one by one, committing at appropriate intervals, and keeping a detailed didary"

**Assistant interpretation:** Revise the design guide/task plan to include safer Codex triggering, YAML PR-list config, release tag minor/major verbs, Glazed command requirements, and then begin implementing the phased plan while maintaining the diary.

**Inferred user intent:** The user wants the design to become an executable implementation plan and wants the first implementation work to follow that plan rather than remain purely theoretical.

**Commit (code):** pending — implementation validated and ready to commit.

### What I did
- Updated the design guide to specify:
  - `ggg pr codex-trigger --force`;
  - default Codex trigger behavior that checks for `EYES` reactions and skips already-running reviews;
  - YAML PR-list input instead of newline `prs.txt`;
  - `ggg release tag-patch`, `tag-minor`, and `tag-major`;
  - all verbs as Glazed commands with row-oriented structured output.
- Rewrote the ticket tasks into phase/task sections.
- Initialized a Go module: `github.com/go-go-golems/infra-tooling`.
- Added dependencies: Glazed, Cobra, YAML.
- Added CLI scaffold:
  - `cmd/ggg/main.go`
  - `internal/cli/root.go`
- Added PR input packages:
  - `pkg/prref`
  - `pkg/prlist`
- Added a `gh`-backed Codex status/trigger client:
  - `pkg/ghclient`
- Added first Glazed PR command:
  - `ggg pr codex-trigger`
- Added release tag implementation and Glazed commands:
  - `ggg release tag-patch`
  - `ggg release tag-minor`
  - `ggg release tag-major`
- Added tests for PR parsing and YAML PR-list loading.

### Why
- `--force` is needed because triggering Codex while an `EYES` reaction is present can create duplicate review requests and muddy readiness interpretation.
- YAML input gives the future release-train file room to grow with metadata.
- Release patch/minor/major share one implementation and should be first-class verbs rather than Makefile-only targets.
- Glazed commands make output processing consistent across human tables and structured JSON/YAML outputs.

### What worked
- `go test ./...` passed.
- `go run ./cmd/ggg --help` built the root command.
- `go run ./cmd/ggg pr codex-trigger --help` built the first PR command.
- `go run ./cmd/ggg release tag-patch --repo . --dry-run --output json` emitted a structured row.

### What didn't work
- N/A in this implementation pass.

### What I learned
- The first useful Go CLI slice can be small: typed PR parsing, YAML PR lists, one mutating Codex command, and release tag verbs already cover several script pain points.
- Module path detection from `go.mod` avoids the exact Makefile placeholder failure seen in `goja-git`.

### What was tricky to build
- The Codex trigger command must be both mutating and row-oriented. It therefore emits an action row for every PR (`triggered`, `skipped_running`, or `would_trigger`) instead of only printing a URL.
- The release commands need to support dry-run without creating tags, but still fetch tags and compute the next tag so operators can see what would happen.

### What warrants a second pair of eyes
- The first `ghclient.CodexStatus` implementation only checks the latest Codex-like review/comment and `@codex review` trigger for `EYES`; Phase 3 should replace this with the fuller readiness model from the Python script.
- `release.Tag` currently shells out to `svu`; this matches current Makefile behavior, but tests should later cover fallback/semver behavior if `svu` is missing.
- Root-level `--with-structured-output` is currently a compatibility flag; the real structured output path is Glazed `--output json/yaml/csv`.

### What should be done in the future
- Commit this implementation slice.
- Continue with Phase 3: port PR readiness parity from Python.
- Add direct tests for release tag calculation using temporary git repos.

### Code review instructions
- Start with `internal/cli/pr/codex_trigger.go` and `pkg/ghclient/ghclient.go` for Codex behavior.
- Review `pkg/prlist/prlist.go` for YAML input compatibility.
- Review `pkg/release/release.go` and `internal/cli/release/tag.go` for release guardrails.
- Validate with:

```bash
go test ./...
go run ./cmd/ggg release tag-patch --repo . --dry-run --output json
```

### Technical details

YAML PR list format implemented in this slice:

```yaml
prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2
  - ref: go-go-golems/go-minitrace#11
```

## Step 4: Port the first PR readiness classifier into Go

The fourth step implemented the first Go version of `ggg pr ready`. This ports the core behavior of `01-pr-ready-check.py`: query GitHub GraphQL, decode status checks and Codex review/comment signals, classify readiness, and emit Glazed rows.

This is still a first parity slice rather than the final implementation. It supports the current state names and Codex stale-feedback logic, but future work should add fixture-based golden tests, full exit-code compatibility, and structured Codex comment rows.

### Prompt Context

**User prompt (verbatim):** "go ahead."

**Assistant interpretation:** Continue the phased implementation plan, starting with the next pending phase after the initial CLI scaffold.

**Inferred user intent:** The user wants continued implementation work, committed in focused increments and documented in the diary.

**Commit (code):** pending — PR readiness implementation validated and ready to commit.

### What I did
- Added `pkg/prready` with:
  - readiness states;
  - check/Codex/finding/report types;
  - status-check classification;
  - Codex signal classification;
  - stale reviewed-commit detection;
  - inline review comment detection;
  - unit tests for ready, current-head Codex feedback, and stale Codex feedback.
- Added `pkg/ghclient.Readiness` with a GitHub GraphQL query covering:
  - `statusCheckRollup.contexts`;
  - PR reviews;
  - review inline comments;
  - PR comments;
  - reaction groups.
- Added `ggg pr ready <pr>` as a Glazed command.
- Added `--findings` to emit one row per readiness finding.
- Registered the command under `ggg pr`.

### Why
- The Python readiness checker is now central release-train infrastructure. Porting it to Go is the highest-value next step because every later batch/train command depends on the same readiness model.

### What worked
- `go test ./...` passed.
- Live smoke against the already-merged Discord Bot PR produced a structured JSON summary row with `state=ready`.

### What didn't work
- The first GraphQL query string had one extra closing brace and GitHub rejected it:

```text
gh: Expected one of SCHEMA, SCALAR, TYPE, ENUM, INPUT, UNION, INTERFACE, actual: RCURLY ("}") at [1, 685]
```

I counted braces in the query string, removed the extra closing brace, reran tests, and then the live smoke succeeded.

### What I learned
- The Go implementation should eventually store GraphQL queries as multiline constants or embedded `.graphql` files to make brace balance easier to review.
- The stale reviewed-commit test must use hex strings because the parser intentionally matches only commit-like `[0-9a-fA-F]+` values.

### What was tricky to build
- The Go classifier must preserve the subtle XGOJA-015 behavior: a newer human `@codex review` trigger does not mask current-head Codex-authored feedback, but stale Codex-authored feedback for an older reviewed commit should not block the current head.

### What warrants a second pair of eyes
- Whether `ggg pr ready` should return non-zero when the state is not ready. The current Glazed command emits rows but does not yet force process exit based on readiness.
- Whether `--findings` should be the default for human output or remain opt-in.

### What should be done in the future
- Add golden JSON fixtures copied from real XGOJA-015 states.
- Add structured `codex_comments` output as a separate command or output mode.
- Add batch readiness using the new Go classifier.

### Code review instructions
- Start with `pkg/prready/prready.go` for the state machine.
- Review `pkg/ghclient/readiness.go` for GraphQL field parity with the Python script.
- Review `internal/cli/pr/ready.go` for Glazed row output.
- Validate with:

```bash
go test ./...
go run ./cmd/ggg pr ready https://github.com/go-go-golems/discord-bot/pull/9 --output json
```

### Technical details

Successful live smoke output included:

```json
{
  "ok": true,
  "state": "ready",
  "terminal": true,
  "repository": "go-go-golems/discord-bot"
}
```

## Step 5: Add YAML batch readiness command

The fifth step implemented the first Go replacement for `05-batch-pr-ready.sh`. The new command reads the YAML PR-list format introduced in the design update and emits Glazed rows for every PR plus a summary row.

The command supports watch mode and keeps the important operator semantics from the shell script: all-ready, terminal Codex feedback, failed checks, errors, and partial readiness are all actionable states.

### Prompt Context

**User prompt (verbatim):** (same as Step 4)

**Assistant interpretation:** Continue implementing the next phase after `ggg pr ready`, now moving into batch readiness with YAML input.

**Inferred user intent:** The user wants the new Go CLI to replace the existing scripts incrementally while preserving release-train behavior.

**Commit (code):** pending — batch readiness implementation validated and ready to commit.

### What I did
- Added `internal/cli/batch` command group.
- Replaced the placeholder `batch` root command with a real command group.
- Added `ggg batch ready <prs.yaml>` as a Glazed command.
- Implemented flags:
  - `--watch`
  - `--interval-seconds`
  - `--timeout-seconds`
  - `--trigger-missing-codex`
- The command emits one row per PR and one summary row.
- The command uses the Go `pkg/prlist` YAML loader and `pkg/ghclient.Readiness` classifier.

### Why
- Batch readiness is the operator-facing loop for release trains. Moving it to Go makes YAML input, structured output, and future release-train orchestration possible.

### What worked
- `go test ./...` passed.
- A live YAML smoke file with Discord Bot PR 9 returned a `ready` PR row and a `ready` summary row using `--output json`.

### What didn't work
- N/A in this implementation pass.

### What I learned
- Glazed structured output works well for batch readiness because PR rows and summary rows can share columns while still being machine-readable.
- The command can preserve shell-script semantics internally even before exact process exit-code parity is fully implemented.

### What was tricky to build
- Watch mode needs to emit rows per attempt without losing structured output. The command adds an `attempt` column so repeated checks remain interpretable.

### What warrants a second pair of eyes
- Whether summary rows should be emitted in the same stream as PR rows or separated into a different output/table in a future version.
- Whether non-ready one-shot batch checks should return a non-zero process status even when rows were emitted successfully.

### What should be done in the future
- Add unit tests for batch aggregation without live GitHub calls.
- Add exact exit-code parity for shell-script replacements.
- Add `ggg batch trigger-codex` or make `--trigger-missing-codex` safer with confirmation.

### Code review instructions
- Start with `internal/cli/batch/ready.go`.
- Validate with:

```bash
go test ./...
cat > /tmp/ggg-prs.yaml <<'YAML'
prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
YAML
go run ./cmd/ggg batch ready /tmp/ggg-prs.yaml --output json
```

### Technical details

Successful live smoke emitted a PR row and summary row with `state: ready`.
