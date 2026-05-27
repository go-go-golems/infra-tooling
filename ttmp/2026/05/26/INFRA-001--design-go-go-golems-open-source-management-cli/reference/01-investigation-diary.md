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

## Step 6: Add readiness exit-code parity

The sixth step added the compatibility layer needed for the Go CLI to replace the shell scripts in automation. The commands already emitted Glazed rows, but they still needed to return meaningful process exit codes for waiting, Codex feedback, failed checks, and partial batch readiness.

I added a typed exit-code error and taught the root `ggg` entry point to translate it into `os.Exit(code)`. The PR and batch readiness commands now emit their rows first, then return the same class of exit code as the scripts they replace.

### Prompt Context

**User prompt (verbatim):** "go ahead"

**Assistant interpretation:** Continue implementation by addressing the previously discussed exit-code parity gap.

**Inferred user intent:** The user wants the Go CLI to become script-compatible, not just produce human/structured rows.

**Commit (code):** pending — exit-code parity implementation validated and ready to commit.

### What I did
- Added `internal/exitcode.Error` with a numeric `Code` and message.
- Updated `cmd/ggg/main.go` to detect `exitcode.Error` and exit with its code.
- Updated `ggg pr ready` so non-ready states return:
  - `1` for waiting/no Codex/not-ready states;
  - `3` for `codex_feedback`;
  - `4` for `failed_checks`.
- Updated `ggg batch ready` so batch summary states return:
  - `0` all ready;
  - `1` waiting;
  - `2` tool/API errors;
  - `3` Codex feedback;
  - `4` failed checks;
  - `5` partial readiness.

### Why
- Existing wait and batch scripts use exit codes as control flow. Without parity, existing automation could not safely replace the scripts with `ggg`.

### What worked
- `go test ./...` passed.
- A live non-go-go-golems PR smoke returned a non-ready row and exited non-zero for `no_codex`.
- Glazed still emitted rows before the typed error reached Cobra/root handling.

### What didn't work
- Using `go run` masks exact non-zero exit codes as a generic `go run` failure in the shell output. Exact code checks should use a built binary when testing codes `3`, `4`, or `5`.

### What I learned
- Returning an error from `RunIntoGlazeProcessor` after `gp.AddRow` still lets Glazed emit the row, which is what we need for structured diagnostics plus script control flow.

### What was tricky to build
- The command must both emit structured data and signal failure. The ordering matters: add rows first, then return the typed exit-code error.

### What warrants a second pair of eyes
- Whether Cobra should suppress the `Error: ...` line for expected non-ready states in a future UX pass.
- Whether batch watch should emit an explicit final summary row before exiting with code `5` in partial-ready cases; it currently does.

### What should be done in the future
- Add tests that build the binary and assert exact process exit codes for fixture-backed readiness states.
- Add fixture-backed batch aggregation tests.

### Code review instructions
- Start with `internal/exitcode/exitcode.go` and `cmd/ggg/main.go`.
- Then review the post-row error returns in `internal/cli/pr/ready.go` and `internal/cli/batch/ready.go`.
- Validate with:

```bash
go test ./...
go build -o /tmp/ggg ./cmd/ggg
/tmp/ggg pr ready <non-ready-pr> --output json; echo $?
```

### Technical details

Observed non-ready smoke with `go run`:

```text
state: no_codex
Error: PR not ready: state=no_codex
exit status 1
```

## Step 7: Harden Codex comment handling and release tagging

The seventh step implemented the first Codex/release hardening pass. The Codex trigger path now reuses the same GitHub snapshot and Codex signal model as readiness, which removes the earlier duplicate simplified parser. I also added a `codex-comments` command so inline Codex review comments can be inspected as structured rows.

On the release side, the tag commands now have the main safety rails that the shell/Makefile flow lacked: dirty-worktree checks, explicit tag target selection, non-dry-run confirmation, existing-tag collision checks, narrow tag push, proxy verification retry, and richer dry-run plan output.

### Prompt Context

**User prompt (verbatim):** "ok do it. Add detailed tasks to the ticket so we can keep track."

**Assistant interpretation:** Add detailed hardening tasks to INFRA-001, then implement the Codex comment/trigger and release hardening work discussed in the previous answers.

**Inferred user intent:** The user wants concrete tracked implementation, not just explanation of hardening concepts.

**Commit (code):** pending — hardening implementation validated and ready to commit.

### What I did
- Added detailed Phase 9 Codex hardening tasks and Phase 10 release hardening tasks to `tasks.md`.
- Refactored `ghclient.Readiness` to build a shared `prready.Snapshot` through `ghclient.Snapshot`.
- Changed `ghclient.CodexStatus` to derive from the shared snapshot instead of its own simplified GraphQL parser.
- Added `pkg/prready/codex_helpers.go` with helpers for sorted signals, latest signal, latest authored signal, current-head matching, and current authored feedback detection.
- Updated `ggg pr codex-trigger` to use the shared snapshot and skip on current Codex feedback unless `--force` is used.
- Added `ggg pr codex-comments` with structured rows for Codex-authored review bodies and inline comments.
- Reworked `pkg/release` to support `TagWithOptions` with dirty checks, target selection, explicit confirmation, existing-tag collision handling, proxy retry, and dry-run plans.
- Added release command flags:
  - `--allow-dirty`
  - `--from`
  - `--commit`
  - `--yes`
- Expanded release rows with `current_tag`, `target`, `dirty`, `existing_tag`, and `plan`.

### Why
- Codex readiness and Codex trigger safety must not diverge. A duplicate simplified parser could skip or trigger differently than the readiness classifier.
- Release commands mutate Git state and publish tags. They need guardrails before they are used as a Makefile replacement.

### What worked
- `go test ./...` passed.
- `ggg pr codex-comments` against Discord Bot PR 9 emitted structured rows with stale/current-head status, paths, lines, bodies, and URLs.
- `ggg release tag-patch --repo . --dry-run --allow-dirty --output json` emitted a rich plan without mutating the repo.

### What didn't work
- Piping the large `codex-comments --output json` output through `head` caused a normal broken-pipe signal after enough rows were printed. This is a shell pipeline artifact, not a command failure in the underlying model.

### What I learned
- The historical Discord Bot PR is a good live smoke for `codex-comments` because it contains multiple stale Codex review comments across reviewed commits.
- The release dry-run row is much more useful once it includes the exact planned commands and target commit.

### What was tricky to build
- Release hardening needs to preserve the simple operator path while adding safety. The command now requires `--yes` for non-dry-run tag pushes, but dry-run remains easy and informative.
- Codex comments are attached to review nodes, while trigger state may be on issue comments. The shared snapshot is the right internal boundary.

### What warrants a second pair of eyes
- `skipped_recent_trigger` is still listed as a task and not implemented yet; the current skip reasons are `skipped_running` and `skipped_current_feedback`.
- GraphQL pagination is still not implemented; review comments are limited to the first 100 per review.
- Release tests using temporary git repositories are still pending.

### What should be done in the future
- Add fixture tests for `codex-comments` and trigger skip decisions.
- Implement `skipped_recent_trigger` if duplicate non-eyes trigger comments become a problem.
- Add release command tests with temporary repositories and fake/proxied verification.

### Code review instructions
- Review Codex changes in:
  - `pkg/ghclient/readiness.go`
  - `pkg/ghclient/ghclient.go`
  - `pkg/prready/codex_helpers.go`
  - `internal/cli/pr/codex_trigger.go`
  - `internal/cli/pr/codex_comments.go`
- Review release changes in:
  - `pkg/release/release.go`
  - `internal/cli/release/tag.go`
- Validate with:

```bash
go test ./...
go run ./cmd/ggg pr codex-comments https://github.com/go-go-golems/discord-bot/pull/9 --output json
go run ./cmd/ggg release tag-patch --repo . --dry-run --allow-dirty --output json
```

### Technical details

The release dry-run now includes a row with fields like:

```json
{
  "module": "github.com/go-go-golems/infra-tooling",
  "current_tag": "v0.0.0",
  "tag": "v0.0.1",
  "target": "origin/main",
  "dirty": true,
  "plan": ["git fetch origin main --tags", "git checkout --detach origin/main", "git tag v0.0.1"]
}
```

## Step 8: Add recent-trigger cooldown and pagination/truncation reporting

The eighth step completed the Codex hardening policy we defined after the first hardening pass. The trigger command now has a configurable recent-trigger cooldown, and the GraphQL snapshot records whether review/comment connections were truncated.

This does not yet implement full GraphQL pagination. Instead, it makes truncation visible and conservative: if the latest current-head Codex review has truncated comments, readiness treats that as actionable Codex feedback rather than silently ignoring possible comments.

### Prompt Context

**User prompt (verbatim):** "do it"

**Assistant interpretation:** Implement the recent-trigger cooldown and pagination/truncation reporting policy that was just defined.

**Inferred user intent:** The user wants the remaining Codex hardening behavior implemented, not just documented.

**Commit (code):** pending — recent-trigger/truncation implementation validated and ready to commit.

### What I did
- Added `CodexSignal.CommentsTruncated`, `Snapshot.ReviewsTruncated`, and `Snapshot.CommentsTruncated`.
- Updated the GitHub GraphQL query to fetch `pageInfo.hasPreviousPage` for review/comment windows and `pageInfo.hasNextPage` for review inline comments.
- Added conservative readiness behavior: current-head truncated Codex review comments produce `codex_feedback`.
- Added `prready.RecentTrigger` helper.
- Added `--recent-trigger-window`, default `10m`, to `ggg pr codex-trigger`.
- Added `skipped_recent_trigger` behavior unless `--force` is used.
- Added trigger row fields:
  - `recent_trigger`
  - `trigger_age_seconds`
  - `recent_trigger_window`
- Added truncation field to `ggg pr codex-comments` rows.
- Added tests for recent trigger detection and truncated current-head review classification.

### Why
- Recent trigger cooldown avoids duplicate `@codex review` spam before Codex has had time to react.
- Truncation reporting avoids silently missing Codex review comments when GitHub GraphQL returns only the first page.

### What worked
- `go test ./...` passed.
- `ggg pr codex-trigger ... --dry-run --output json` emitted the new recent-trigger fields.

### What didn't work
- Full pagination is still not implemented; this pass reports and reacts to truncation instead of fetching additional pages.

### What I learned
- For readiness, conservative truncation handling is safer than silent partial data. For historical reporting, `codex-comments` can expose `truncated=true` so the operator knows to inspect GitHub manually or wait for a future pagination implementation.

### What was tricky to build
- The GraphQL query needs pageInfo at three levels: PR reviews, PR issue comments, and per-review inline comments. Only the per-review inline comment truncation can directly affect current-head Codex feedback completeness.

### What warrants a second pair of eyes
- Whether `comments(last:100).pageInfo.hasPreviousPage` should influence readiness or only diagnostics. It is currently diagnostic unless the latest authored review comment page is truncated.
- Whether `--recent-trigger-window` should default to 10 minutes or a shorter value.

### What should be done in the future
- Implement actual pagination for `ggg pr codex-comments --all`.
- Add live or fixture tests for `skipped_recent_trigger` command rows.

### Code review instructions
- Review `pkg/ghclient/readiness.go` for pageInfo fields.
- Review `pkg/prready/prready.go` and `pkg/prready/codex_helpers.go` for conservative truncation and recent-trigger logic.
- Review `internal/cli/pr/codex_trigger.go` for skip ordering.
- Validate with:

```bash
go test ./...
go run ./cmd/ggg pr codex-trigger https://github.com/go-go-golems/discord-bot/pull/9 --dry-run --output json
```

### Technical details

Skip priority remains:

1. `--dry-run` -> `would_trigger`
2. running Codex `EYES` -> `skipped_running`
3. current-head feedback -> `skipped_current_feedback`
4. recent trigger -> `skipped_recent_trigger`
5. otherwise -> `triggered`
