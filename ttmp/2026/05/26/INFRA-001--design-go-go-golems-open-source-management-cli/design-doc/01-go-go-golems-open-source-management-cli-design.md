---
Title: Go-Go-Golems Open Source Management CLI Design
Ticket: INFRA-001
Status: active
Topics:
    - cli
    - github
    - release
    - automation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../../../../workspaces/2026-05-24/add-js-providers/go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/10-validate-downstream-focused.sh
      Note: XGOJA-015 focused downstream validation profile evidence
    - Path: docs/go-go-golems/package-publishing-release-train.md
      Note: Current release-train policy and operator workflow
    - Path: examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk
      Note: Current dependency bump algorithm with GOWORK=off
    - Path: scripts/go-go-golems/01-pr-ready-check.py
      Note: Current GitHub GraphQL PR readiness and Codex signal classifier
    - Path: scripts/go-go-golems/05-batch-pr-ready.sh
      Note: Current batch readiness/watch state machine and exit codes
ExternalSources:
    - https://docs.github.com/en/graphql
    - https://cli.github.com/manual/
Summary: Analysis and implementation guide for a future Go package and CLI that manages go-go-golems open-source PR readiness, release trains, dependency bumps, validation, and publication.
LastUpdated: 2026-05-26T23:20:00-04:00
WhatFor: Use as the intern-oriented technical blueprint for replacing ad-hoc go-go-golems management scripts with a typed Go CLI/toolbox.
WhenToUse: Use before implementing an infra-tooling Go CLI, extending release-train automation, or debugging current script behavior.
---


# Go-Go-Golems Open Source Management CLI Design

## Executive summary

The current go-go-golems open-source management workflow is implemented as a set of Bash and Python scripts plus Markdown playbooks. The scripts already encode valuable policy: a pull request is not ready just because CI is green; Codex must also be finished and satisfied; downstream repositories must validate with `GOWORK=off`; release trains must follow dependency order; and tag publication must be verified through the public Go module proxy.

The scripts worked during the XGOJA-015 release train, but their responsibilities have outgrown ad-hoc shell wrappers. They now span GitHub GraphQL, `gh` CLI calls, status-check state machines, Codex review parsing, release-train batch state, per-repository validation commands, tag publication, Go module proxy verification, and documentation bookkeeping. A typed Go package plus CLI would make those concepts explicit, testable, reusable, and safer for future multi-repository maintenance.

This document proposes a future `ggg` or `go-go-golems` management CLI with packages for:

- GitHub pull request inspection and mutation.
- Codex signal parsing and readiness classification.
- Batch release-train state and dependency-order operations.
- Go module dependency bump and proxy visibility checks.
- Repository validation profiles.
- Release/tag publication wrappers.
- Human-readable tables, JSON output, and resumable runbooks.

The design intentionally starts by documenting the existing system. An intern should be able to read this guide, understand the current scripts, and then implement the Go CLI in small phases without guessing at hidden workflow rules.

## Problem statement and scope

### Problem

The current management workflow is spread across multiple locations:

1. Reusable infra scripts in `infra-tooling/scripts/go-go-golems/`.
2. Playbooks in `infra-tooling/docs/go-go-golems/`.
3. Makefile snippets in `infra-tooling/examples/go-go-golems/`.
4. Ticket-local historical scripts in the XGOJA-015 workspace.
5. Operator knowledge from the recent release train.

This works for an expert operator, but it is hard for a new engineer to answer basic questions:

- Which APIs are being used?
- What data is read from GitHub?
- What does `READY` mean?
- Why does a newer `@codex review` comment not erase older Codex feedback?
- When is it safe to merge?
- When is it safe to tag?
- Which validation command belongs to which repository?
- What should happen if a Makefile has a placeholder module path?
- Which parts are reusable policy and which parts are release-train-specific configuration?

### Scope

This design covers a future Go CLI/toolbox for go-go-golems open-source repository operations:

- PR readiness checks.
- Codex review triggering and parsing.
- Batch PR watch/summary behavior.
- Release-train dependency ordering and active PR lists.
- Go module dependency bumping.
- `GOWORK=off` validation.
- Tag-patch release publication.
- Go module proxy verification.
- Per-repository validation profiles.
- Evidence and diagnostics collection.

This design does not implement the CLI yet. It is a research/design guide that records current behavior and proposes building blocks.

## Current repository layout

The `infra-tooling` repository is already the neutral home for shared mechanics. The root README states that the repo should contain shared release/GitOps helper scripts, reusable workflow templates, target metadata examples, extracted platform docs, and generic source-repo-to-GitOps PR automation. It explicitly excludes app-specific business logic and live GitOps state.

Relevant current files:

```text
infra-tooling/
  README.md
  scripts/go-go-golems/
    00-pr-ready-check.sh
    01-pr-ready-check.py
    02-trigger-codex-review.sh
    03-watch-codex-reactions.py
    04-wait-pr-ready.sh
    05-batch-pr-ready.sh
    06-batch-trigger-codex-review.sh
  docs/go-go-golems/
    package-publishing-release-train.md
    glazed-linting-rollout-playbook.md
    logcopter-rollout-colleague-instructions.md
    playbooks/pr-readiness-check-scripts.md
    playbooks/logcopter-package-rollout-playbook.md
  examples/go-go-golems/
    Makefile.bump-go-go-golems.snippet.mk
    Makefile.bump-go-go-golems-gowork-off.snippet.mk
```

The XGOJA-015 ticket adds historical scripts that should influence the design:

```text
go-go-goja/ttmp/.../XGOJA-015.../scripts/
  08-extract-codex-review-comments.sh
  09-pr-check-summary.sh
  10-validate-downstream-focused.sh
```

The `10-validate-downstream-focused.sh` script is especially important because it shows that validation is not a single global command. It is a mapping from repository name to focused package tests, lints, and smoke commands.

## Current functionality inventory

### PR readiness check

The main implementation is `scripts/go-go-golems/01-pr-ready-check.py`. It performs four high-level tasks:

1. Parse a PR reference.
2. Query GitHub GraphQL.
3. Convert checks and Codex signals into findings.
4. Classify findings into a state.

Evidence:

- The script accepts either a full PR URL or an `owner/repo#number` reference through `parse_pr` (`sources/01-line-anchored-evidence.txt`, lines 119 onward).
- The GraphQL query starts at `QUERY = r"""` and includes `statusCheckRollup`, `reviews`, review `comments`, and issue `comments` (`sources/01-line-anchored-evidence.txt`, lines 37-87).
- The checker reads `statusCheckRollup.contexts.nodes` to inspect check runs and status contexts (`sources/01-line-anchored-evidence.txt`, line 175).
- It extracts inline review comments with `code_review_comments` (`sources/01-line-anchored-evidence.txt`, line 159).

Current readiness states:

```text
ready
waiting_checks
waiting_codex
no_codex
failed_checks
codex_feedback
not_ready
error
```

The state machine matters. Operators used it to decide whether to wait, fix code, merge, or inspect CI.

### Codex signal model

Codex review information is not represented by one GitHub object. The current scripts combine several sources:

- Codex-authored PR reviews.
- Codex-authored issue comments.
- Human `@codex review` trigger comments.
- Reaction groups on those comments/reviews.
- Inline review comments.
- Reviewed commit information embedded in Codex review body text.

The key behavior learned during XGOJA-015 is that a newer human `@codex review` trigger must not mask older current-head Codex feedback. The readiness checker therefore separates:

- latest overall Codex signal; and
- latest Codex-authored signal.

If Codex-authored comments are for the current head commit, they block readiness even if a newer trigger exists. If they are for an older reviewed commit, they are stale and do not block the new head.

### Batch PR readiness

`05-batch-pr-ready.sh` reads a file of PR references and prints a readiness table. It wraps the single-PR checker and returns meaningful exit codes.

Evidence:

- `run_once` starts the batch table and counters (`sources/01-line-anchored-evidence.txt`, line 280).
- The package-publishing playbook says batch watch mode stops on terminal failures, Codex feedback, all-ready state, or partial readiness (`sources/01-line-anchored-evidence.txt`, line 480).

Exit codes from the current behavior:

```text
0 = all ready
1 = all still waiting / timeout in watch mode
2 = checker or script error
3 = codex_feedback
4 = failed_checks
5 = partial readiness: at least one PR ready while others still wait
```

The `5` case is important. In a release train, partial readiness is actionable because the next dependency-order PR may be mergeable even while leaf PRs are still waiting.

### Codex triggering

`02-trigger-codex-review.sh` posts the standard trigger comment:

```bash
gh pr comment "$1" --body '@codex review'
```

`06-batch-trigger-codex-review.sh` applies the same operation to a PR list.

The future Go CLI should not hide that this is a mutating GitHub operation. It should support dry-run, confirmation, and clear output.

### Waiting on one PR

`04-wait-pr-ready.sh` polls one PR until ready, timeout, or terminal operator-action state. It exits immediately on:

- `codex_feedback` with exit `3`;
- `failed_checks` with exit `4`.

This is better than a blind timeout loop because it stops when a human needs to fix code.

### Codex inline review extraction

XGOJA-015 added `08-extract-codex-review-comments.sh` because the first readiness messages were too generic. This script queries PR reviews and prints Codex-authored review bodies plus inline review comments.

Evidence:

- The script usage supports individual PRs or `--file` (`sources/02-xgoja-015-script-evidence.txt`, lines 9-10).
- Its GraphQL query reads review `comments(first:100)` with `path`, `line`, `body`, `url`, and `author.login` (`sources/02-xgoja-015-script-evidence.txt`, line 46).

This should become a first-class command:

```bash
ggg pr codex-comments <pr>
ggg pr codex-comments --file prs.txt --format markdown
```

### PR check summaries

XGOJA-015 added `09-pr-check-summary.sh`, which loops a PR list and runs `gh pr checks`. This is not as structured as the GraphQL checker, but it is useful for operator-facing diagnostics. The future CLI should have both structured readiness output and raw check summaries.

### Focused downstream validation

XGOJA-015 added `10-validate-downstream-focused.sh`. It maps repository names to targeted validation commands.

Evidence:

- The script documents the key invariant: run one repo at a time with `GOWORK=off` so local workspace replacements cannot hide unpublished upstream tags (`sources/02-xgoja-015-script-evidence.txt`, line 121).
- It validates `discord-bot` with `GOWORK=off go test ./pkg/xgoja/provider ./internal/jsdiscord ./pkg/botcli -count=1` (`sources/02-xgoja-015-script-evidence.txt`, line 137).
- It validates `loupedeck` with focused tests plus generated xgoja smoke (`sources/02-xgoja-015-script-evidence.txt`, lines 145-146).
- It validates `workspace-manager` and `goja-git` with focused tests plus golangci-lint (`sources/02-xgoja-015-script-evidence.txt`, lines 150-156).

This should become configuration, not hard-coded shell:

```yaml
repositories:
  discord-bot:
    root: /home/manuel/workspaces/2026-05-24/add-js-providers/discord-bot
    validation:
      - name: focused-tests
        env: { GOWORK: "off" }
        command: go
        args: [test, ./pkg/xgoja/provider, ./internal/jsdiscord, ./pkg/botcli, -count=1]
  loupedeck:
    validation:
      - name: focused-tests
        env: { GOWORK: "off" }
        command: go
        args: [test, ./runtime/js, ./runtime/js/provider, ./cmd/loupedeck/cmds/verbs, ./pkg/jsmetrics, -count=1]
      - name: generated-smoke
        command: make
        args: [-C, examples/xgoja/loupedeck-command-provider, smoke]
```

### Release-train playbook

The release playbook defines the central operational policy:

- Rollouts must follow dependency order.
- Early downstream PRs are allowed for CI/Codex feedback.
- Downstream PRs must not merge until required upstream tags are visible and `GOWORK=off` validation passes.
- Batch watch stops on operator-action states.
- Published module visibility is checked with `go list -m -versions` or `go list -m module@version`.

Evidence:

- The playbook states not to merge downstream PRs until upstream tags are visible and `GOWORK=off` validation passes (`sources/01-line-anchored-evidence.txt`, line 471).
- It gives upstream visibility commands including `go list -m -versions` (`sources/01-line-anchored-evidence.txt`, lines 493 onward).
- It provides the `GOWORK=off go get` dependency bump loop (`sources/01-line-anchored-evidence.txt`, lines 508-513).
- It says to prefer `GOWORK=off go test ./...` for published dependency graph smoke tests (`sources/01-line-anchored-evidence.txt`, lines 525-528).
- It documents `gh pr merge` after readiness (`sources/01-line-anchored-evidence.txt`, line 586).

### Dependency bump snippets

`examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk` is the core reusable dependency bump algorithm. It scans `go.mod` for `github.com/go-go-golems/*` dependencies, runs `GOWORK=off go get dep@latest`, then runs `GOWORK=off go mod tidy`.

Evidence:

- The snippet states that forcing `GOWORK=off` prevents local `go.work` from hiding missing upstream releases (`sources/01-line-anchored-evidence.txt`, lines 813 onward).
- It uses the awk-based dependency extraction and `GOWORK=off go get` loop (`sources/01-line-anchored-evidence.txt`, line 824).

The future CLI should implement this as a typed module graph operation, not a Makefile snippet.

## Data model for a Go package

### Core types

```go
type PRRef struct {
    Owner  string
    Repo   string
    Number int
}

type CheckState string
const (
    CheckSuccess CheckState = "success"
    CheckPending CheckState = "pending"
    CheckFailure CheckState = "failure"
)

type CheckRun struct {
    Name       string
    Status     string
    Conclusion string
    URL        string
}

type CodexSignalKind string
const (
    CodexReview  CodexSignalKind = "review"
    CodexComment CodexSignalKind = "comment"
    CodexTrigger CodexSignalKind = "trigger"
)

type CodexSignal struct {
    Kind          CodexSignalKind
    Author        string
    URL           string
    SubmittedAt   time.Time
    Body          string
    ReviewedCommit string
    ThumbsUp      int
    Eyes          int
    AuthoredByCodex bool
    InlineComments []ReviewComment
}

type ReviewComment struct {
    Path string
    Line int
    Body string
    URL  string
}

type ReadinessState string
const (
    Ready          ReadinessState = "ready"
    WaitingChecks  ReadinessState = "waiting_checks"
    WaitingCodex   ReadinessState = "waiting_codex"
    NoCodex        ReadinessState = "no_codex"
    FailedChecks   ReadinessState = "failed_checks"
    CodexFeedback  ReadinessState = "codex_feedback"
    NotReady       ReadinessState = "not_ready"
    ErrorState     ReadinessState = "error"
)

type Finding struct {
    OK      bool
    Kind    string
    Message string
    URL     string
}

type ReadinessReport struct {
    PR                PRRef
    URL               string
    MergeStateStatus  string
    ReviewDecision    string
    HeadSHA           string
    State             ReadinessState
    Terminal          bool
    FailedCheckKinds  []string
    Findings          []Finding
}
```

### Release-train types

```go
type Repository struct {
    Name       string
    ModulePath string
    Root       string
    Remote     string
    MainBranch string
    Dependencies []string
    Validation []ValidationStep
}

type ValidationStep struct {
    Name    string
    WorkDir string
    Env     map[string]string
    Command string
    Args    []string
    Timeout time.Duration
}

type ReleasePlan struct {
    Name         string
    Repositories []Repository
    Edges        []DependencyEdge
    PRs          []PullRequestPlan
}

type DependencyEdge struct {
    FromRepo string // dependent
    ToRepo   string // dependency
}

type PullRequestPlan struct {
    Repository string
    PR         PRRef
    RequiredTags []ModuleVersion
}

type ModuleVersion struct {
    Module string
    Version string
}
```

## Proposed CLI command groups

All verbs should be implemented as Glazed commands. The default output should be concise and useful to a human operator, usually a small table or one-row summary. When the operator asks for structured output, the command should emit row-oriented data through Glazed output plumbing. In practice this means every verb should implement `RunIntoGlazeProcessor`, emit `types.Row` values, and support Glazed output flags such as JSON/YAML/CSV selection. The CLI should also provide a compatibility alias or explicit flag named `--with-structured-output` if the final root command needs a binary switch between concise prose and structured rows.


### `ggg pr ready`

Checks one PR.

```bash
ggg pr ready https://github.com/go-go-golems/discord-bot/pull/9
ggg pr ready go-go-golems/discord-bot#9 --json
ggg pr ready --codex-author-regex 'chatgpt-codex-connector' <pr>
```

Output should include:

- state;
- terminal flag;
- failed check kinds;
- findings;
- inline Codex comments if present;
- merge state;
- head SHA.

### `ggg pr codex-trigger`

Posts `@codex review`.

```bash
ggg pr codex-trigger <pr>
ggg pr codex-trigger --file prs.yaml --dry-run
ggg pr codex-trigger --file prs.yaml --yes
ggg pr codex-trigger --file prs.yaml --force
```

This is mutating, so default behavior should show what will happen and require confirmation unless `--yes` is passed. By default it should also check whether the latest Codex signal has an `EYES` reaction and skip triggering if a Codex run already appears to be in progress. `--force` overrides that safety check and always posts a fresh `@codex review` comment.

All PR-list input should use YAML rather than ad-hoc newline text. The minimum supported form is:

```yaml
prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2
```

The YAML form gives the future CLI room to attach metadata such as dependency order, target release version, validation profile, and required upstream tags.

### `ggg pr codex-comments`

Prints Codex-authored review bodies and inline review comments.

```bash
ggg pr codex-comments <pr>
ggg pr codex-comments --file prs.txt --format markdown
ggg pr codex-comments <pr> --current-head-only
```

### `ggg pr checks`

Equivalent to structured `gh pr checks` summaries, but with JSON support.

```bash
ggg pr checks <pr>
ggg pr checks --file prs.txt --json
```

### `ggg batch ready`

Replaces `05-batch-pr-ready.sh`.

```bash
ggg batch ready prs.yaml
ggg batch ready prs.yaml --watch --interval 30s --timeout 30m
ggg batch ready prs.yaml --trigger-missing-codex
```

Exit codes should preserve current semantics for script compatibility.

### `ggg repo deps`

Scans `go.mod` for go-go-golems dependencies.

```bash
ggg repo deps --repo /path/to/repo
ggg repo deps --workspace /home/manuel/workspaces/2026-05-24/add-js-providers --json
```

### `ggg repo bump-go-go-golems`

Replaces the Makefile snippets.

```bash
ggg repo bump-go-go-golems --repo /path/to/repo --gowork off
ggg repo bump-go-go-golems --repo /path/to/repo --module github.com/go-go-golems/go-go-goja@v0.6.0
```

### `ggg repo validate`

Runs configured validation profiles.

```bash
ggg repo validate discord-bot --profile xgoja-focused
ggg repo validate loupedeck --profile xgoja-focused --dry-run
ggg repo validate --all --profile release-train
```

### `ggg release tag-patch`, `tag-minor`, and `tag-major`

Wrap existing Makefile behavior but add guardrails. Patch, minor, and major release verbs should share one implementation with a version-bump mode.

```bash
ggg release tag-patch --repo go-minitrace
ggg release tag-minor --repo goja-git --verify-proxy
ggg release tag-major --repo discord-bot --from origin/main
```

Guardrails:

- fetch tags first;
- ensure target commit is expected;
- check clean worktree or detached `origin/main`;
- compute next tag with `svu` or Go semver helper;
- push only the new tag;
- verify with `GOPROXY=proxy.golang.org go list -m module@version`.

### `ggg train status`

Reads a release-train file and prints dependency/pr/release state.

```bash
ggg train status xgoja-release.yaml
ggg train watch xgoja-release.yaml --merge-ready=false
ggg train next xgoja-release.yaml
```

## API and implementation architecture

### Package layout

```text
internal/cli/                 Cobra/Glazed command wiring
pkg/githubx/                  GitHub GraphQL and gh-compatible operations
pkg/prready/                  readiness classifier and Codex parser
pkg/releasetrain/             release train model, dependency order, batch state
pkg/gomodx/                   go.mod scanning, go get/tidy, module proxy checks
pkg/validate/                 validation profile runner
pkg/release/                  tag computation, push, module proxy verification
pkg/config/                   YAML config loading and defaults
pkg/run/                      command execution, env handling, logging, dry-run
```

### GitHub client interface

```go
type GitHubClient interface {
    PullRequest(ctx context.Context, ref PRRef) (*PullRequestSnapshot, error)
    CommentPR(ctx context.Context, ref PRRef, body string) (*Comment, error)
    MergePR(ctx context.Context, ref PRRef, opts MergeOptions) (*MergeResult, error)
    PRChecks(ctx context.Context, ref PRRef) ([]CheckRun, error)
}
```

Implementation options:

1. Use GitHub GraphQL directly through `githubv4` or raw HTTP.
2. Shell out to `gh api graphql` initially for compatibility.
3. Provide both: direct client for long-term, `gh` client for quick parity.

Recommendation: start with a `gh`-backed client because the scripts already depend on authenticated `gh`. Hide it behind `GitHubClient` so a direct GraphQL client can replace it later.

### Readiness classifier pseudocode

```text
function classify(snapshot):
    findings = []
    findings += checkFindings(snapshot.statusChecks)
    codexSignals = collectCodexSignals(snapshot.reviews, snapshot.comments)

    if codexSignals is empty:
        findings += fail("no Codex signal")
        return report(no_codex, terminal=false, findings)

    latest = newest(codexSignals)
    latestAuthored = newest(signal where signal.AuthoredByCodex)

    if latestAuthored exists:
        if latestAuthored.reviewedCommit exists and latestAuthored.reviewedCommit != snapshot.headSHA:
            findings += ok("Codex findings are stale")
        else if latestAuthored.inlineComments not empty:
            findings += fail("Codex inline comments", comments)
        else if bodyIsSubstantive(latestAuthored.body):
            findings += fail("Codex body contains feedback")
        else:
            findings += ok("Codex-authored review is benign")

    if latest.eyes > 0:
        findings += fail("Codex still running")
    if latest.thumbsUp == 0 and !bodySatisfied(latest.body):
        findings += fail("Codex not satisfied")

    return classifyFindings(findings)
```

### Batch watch pseudocode

```text
function watchBatch(prs, interval, timeout):
    start = now
    loop:
        reports = checkAll(prs)
        printTable(reports)

        if all reports ready:
            return 0
        if any report state == codex_feedback:
            return 3
        if any report state == failed_checks:
            return 4
        if any report state == error:
            return 2
        if any report ready:
            return 5
        if now - start > timeout:
            return 1
        sleep(interval)
```

### Release tag pseudocode

```text
function tagPatch(repo):
    git.fetch(repo, "origin", "main", tags=true)
    target = git.revParse(repo, "origin/main")
    requireCleanWorktree(repo)
    checkoutDetached(repo, target)

    current = semver.highestTag(git.tags(repo))
    next = current.patchPlusOne()

    git.tag(repo, next, target)
    git.pushTag(repo, "origin", next)

    module = readModulePath(repo/go.mod)
    gomod.verifyProxy(module, next)

    return ReleaseResult{Tag: next, Commit: target}
```

The XGOJA-015 goja-git Makefile bug demonstrates why `module = readModulePath(go.mod)` is safer than hard-coded module strings. The Makefile used `github.com/go-go-golems/XXX` and failed proxy verification even though the tag was pushed.

## Diagrams

### PR readiness data flow

```text
         PR URL / owner/repo#n
                  |
                  v
          parse PR reference
                  |
                  v
        GitHub GraphQL query
                  |
      +-----------+------------+
      |                        |
 statusCheckRollup        reviews/comments
      |                        |
      v                        v
 check findings          Codex signal parser
      |                        |
      +-----------+------------+
                  |
                  v
          readiness classifier
                  |
        +---------+----------+
        |                    |
 human table          JSON / exit code
```

### Release train loop

```text
             dependency graph
                   |
                   v
           choose next repo/PR
                   |
                   v
        validate with GOWORK=off
                   |
                   v
          open/update PR + Codex
                   |
                   v
          batch readiness watch
                   |
     +-------------+--------------+
     |                            |
 feedback/failure              ready
     |                            |
 fix and push                 merge PR
                                  |
                                  v
                         tag-patch release
                                  |
                                  v
                         verify Go proxy
                                  |
                                  v
                         update downstreams
```

## Configuration model

A future release-train config should be explicit and checked into a ticket or repo:

```yaml
name: xgoja-runtime-api-release
workspace: /home/manuel/workspaces/2026-05-24/add-js-providers
codexAuthorRegex: "chatgpt-codex-connector|codex|openai"
repositories:
  go-go-goja:
    module: github.com/go-go-golems/go-go-goja
    root: go-go-goja
  geppetto:
    module: github.com/go-go-golems/geppetto
    root: geppetto
    dependsOn: [go-go-goja]
  discord-bot:
    module: github.com/go-go-golems/discord-bot
    root: discord-bot
    dependsOn: [go-go-goja]
    validationProfile: xgoja-focused
validationProfiles:
  xgoja-focused:
    discord-bot:
      - name: focused-tests
        env: { GOWORK: "off" }
        command: go
        args: [test, ./pkg/xgoja/provider, ./internal/jsdiscord, ./pkg/botcli, -count=1]
```

## Missing functionality and gaps

### Script-level gaps

- No typed tests around GraphQL response parsing.
- `ggg pr codex-comments` now emits structured Codex comment rows, but readiness JSON can still expose richer nested comment data in a future pass.
- Current implementation reports GraphQL truncation for reviews/comments and treats current-head truncated Codex review comments conservatively, but it still does not fetch additional pages beyond the first 100 review comments.
- Batch scripts are shell-based and hard to unit-test.
- Release/tag scripts rely on Makefiles that can contain stale placeholders.
- Focused validation profiles are hard-coded in ticket-local shell.
- No durable release-train state file with PRs, tags, merge commits, and validation history.
- No safe dry-run for merge/release operations.
- No unified report generation for docmgr/reMarkable handoff.

### Workflow gaps

- Dependency ordering is manual and based on operator inspection.
- Early downstream PRs are allowed, but there is no tool-enforced dependency gate before merge.
- `GOWORK=off` is a convention, not enforced by a central runner.
- Repository-specific CI failure remedies are in prose, not encoded as diagnostics.
- Release publication verification is inconsistent across Makefiles.

### API gaps

- Codex reviewed-commit parsing depends on body text format.
- Reaction counts do not verify that the reaction came from Codex rather than another user.
- `gh` CLI output is convenient but not a stable API for all commands.

## Implementation phases and task plan

### Phase 1: CLI scaffold and Glazed command foundation

Goal: create the Go module, root command, command groups, and output conventions before porting business logic.

Tasks:

1. Initialize a Go module in `infra-tooling` if one does not exist.
2. Add Glazed, Cobra, and YAML dependencies.
3. Create `cmd/ggg/main.go` as the initial binary entry point.
4. Create `internal/cli` root wiring with command groups: `pr`, `batch`, `repo`, `release`, and `train`.
5. Add a helper for building Glazed commands with concise table defaults and structured output support.
6. Add a root-level `--with-structured-output` compatibility flag if needed, but keep row-oriented Glazed output as the implementation mechanism.
7. Add a smoke test or `go test ./...` validation that proves the command tree builds.

### Phase 2: PR references, YAML PR lists, and Codex trigger safety

Goal: implement the first mutating command while introducing typed PR input handling.

Tasks:

1. Implement `pkg/prref.Parse` for URL and `owner/repo#number` formats.
2. Implement YAML PR list loading for:
   - `prs: ["https://github.com/.../pull/1"]`
   - `prs: [{repo: "go-go-golems/repo", number: 1}]`
3. Implement a `GitHubClient` interface with an initial `gh`-backed implementation.
4. Implement `CodexRunInProgress(ctx, pr)` by querying latest Codex signals and checking for `EYES` reactions.
5. Implement `ggg pr codex-trigger <pr|--file prs.yaml>` as a Glazed command.
6. Add `--force`; default behavior skips trigger when a Codex run is already in progress.
7. Add `--dry-run` and emit one row per PR with action `triggered`, `skipped_running`, or `would_trigger`.
8. Add tests for PR parsing and YAML list loading.

### Phase 3: PR readiness parity

Goal: port the current Python readiness classifier into Go.

Tasks:

1. Port the GraphQL query fields from `01-pr-ready-check.py`.
2. Decode status checks, reviews, issue comments, reaction groups, and inline review comments into typed structs.
3. Implement check-run classification.
4. Implement Codex signal collection.
5. Implement stale reviewed-commit detection.
6. Implement inline comment extraction and structured `codex_comments` rows.
7. Implement `ggg pr ready <pr>` with Glazed row output and JSON/YAML output support.
8. Preserve current state names and exit-code semantics.
9. Add golden fixtures from observed XGOJA-015 states.

### Phase 4: Batch readiness with YAML input

Goal: replace `05-batch-pr-ready.sh` without losing operational behavior.

Tasks:

1. Implement `ggg batch ready prs.yaml`.
2. Support `--watch`, `--interval`, `--timeout`, and `--trigger-missing-codex`.
3. Preserve exit codes `0`, `1`, `2`, `3`, `4`, and `5`.
4. Emit one row per PR plus a summary row/table.
5. Add tests for batch aggregation and partial readiness.

### Phase 5: Release verbs and Go module verification

Goal: provide safe tag/release commands that do not rely on fragile Makefile placeholders.

Tasks:

1. Implement module-path detection from `go.mod`.
2. Implement highest semver tag discovery.
3. Implement next patch, minor, and major version calculation.
4. Implement `ggg release tag-patch`, `ggg release tag-minor`, and `ggg release tag-major`.
5. Add guardrails: fetch tags, verify clean worktree, target `origin/main` or explicit commit, push only the new tag.
6. Verify publication with `GOPROXY=proxy.golang.org go list -m module@version`.
7. Emit structured rows with tag, commit, module, and verification status.
8. Add tests using temporary local Git repositories.

### Phase 6: Validation profiles

Goal: convert `10-validate-downstream-focused.sh` into reusable YAML-driven profiles.

Tasks:

1. Define validation profile YAML schema.
2. Implement command runner with environment, working directory, timeout, dry-run, and log capture.
3. Implement `ggg repo validate <repo> --profile <name>`.
4. Port XGOJA-015 focused validations into a sample profile.
5. Add tests for dry-run and command expansion.

### Phase 7: Release-train orchestration

Goal: make multi-repo release trains explicit and resumable.

Tasks:

1. Define release-train YAML schema with repositories, dependencies, PRs, validation profiles, and required tags.
2. Implement dependency graph loading and topological sort.
3. Implement `ggg train status`.
4. Implement `ggg train next` to recommend the next safe operator action.
5. Implement merge gates that require readiness and visible upstream tags.
6. Record merge commits, tags, and verification results in a run-state file.

### Phase 8: Reporting and docmgr integration

Goal: generate durable handoff artifacts after release operations.

Tasks:

1. Generate Markdown release reports from run-state files.
2. Generate changelog snippets for docmgr tickets.
3. Optionally add `ggg docmgr changelog` helpers that shell out to `docmgr`.
4. Keep reMarkable upload as documented workflow unless it proves stable enough to automate.

## Testing strategy

### Unit tests

- PR reference parsing.
- GraphQL JSON decoding.
- Check-run classification.
- Codex signal collection.
- Codex body benign/substantive/satisfied parsing.
- Reviewed-commit stale/current-head logic.
- Batch exit-code classification.
- Semver/tag increment logic.
- `go.mod` module path and dependency scanning.

### Integration tests

- Run `ggg pr ready --json` against a known public PR fixture, but keep this optional because live GitHub tests are flaky.
- Run command runner in dry-run mode against a temporary repo.
- Create a local Git repo with tags and verify patch-tag computation.

### Golden tests

The current Python JSON outputs and XGOJA-015 examples should become golden fixtures. This gives confidence that the Go rewrite preserves behavior before adding new features.

## Risks and alternatives

### Alternative: keep Bash/Python scripts

This is lowest cost, but the scripts are already accumulating state-machine complexity. Adding dependency graphs, validation profiles, release history, and structured reports will be brittle in shell.

### Alternative: build only a Python CLI

Python would reuse existing code faster. The drawback is that go-go-golems repositories are Go-first, and a Go CLI can more easily become a distributed binary with typed Go module helpers.

### Alternative: implement as GitHub Actions only

Actions are useful for CI, but this workflow is operator-driven and spans local worktrees, release tags, module proxy verification, docmgr, and reMarkable. A local CLI remains useful.

## Recommended next implementation step

Build a minimal Go module in `infra-tooling` with one command:

```bash
ggg pr ready <pr> --json
```

Keep the command behavior compatible with `01-pr-ready-check.py`. Once the classifier has fixtures and tests, add `batch ready`, then move to release/tag and validation profile commands.

## References

Primary current implementation:

- `/home/manuel/code/wesen/go-go-golems/infra-tooling/scripts/go-go-golems/01-pr-ready-check.py`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/scripts/go-go-golems/05-batch-pr-ready.sh`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/scripts/go-go-golems/04-wait-pr-ready.sh`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/scripts/go-go-golems/02-trigger-codex-review.sh`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/scripts/go-go-golems/06-batch-trigger-codex-review.sh`

Primary playbooks:

- `/home/manuel/code/wesen/go-go-golems/infra-tooling/docs/go-go-golems/package-publishing-release-train.md`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/docs/go-go-golems/playbooks/pr-readiness-check-scripts.md`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/docs/go-go-golems/logcopter-rollout-colleague-instructions.md`
- `/home/manuel/code/wesen/go-go-golems/infra-tooling/docs/go-go-golems/glazed-linting-rollout-playbook.md`

XGOJA-015 historical scripts:

- `/home/manuel/workspaces/2026-05-24/add-js-providers/go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/08-extract-codex-review-comments.sh`
- `/home/manuel/workspaces/2026-05-24/add-js-providers/go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/09-pr-check-summary.sh`
- `/home/manuel/workspaces/2026-05-24/add-js-providers/go-go-goja/ttmp/2026/05/26/XGOJA-015--release-xgoja-runtime-api-and-bump-downstream-repositories/scripts/10-validate-downstream-focused.sh`

Evidence snapshots in this ticket:

- `sources/01-line-anchored-evidence.txt`
- `sources/02-xgoja-015-script-evidence.txt`
