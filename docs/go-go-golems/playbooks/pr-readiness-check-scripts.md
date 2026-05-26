---
Title: PR readiness check scripts
Ticket: PR-REVIEW-READY-001
Status: active
Topics:
    - automation
    - github
    - cicd
    - documentation
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../pinocchio/.github/workflows/push.yml
      Note: Live PR test target for intentional bad change and revert
    - Path: ttmp/2026/05/26/PR-REVIEW-READY-001--automate-pr-readiness-checks-for-codex-reviews/scripts/00-pr-ready-check.sh
      Note: Bash entry point for the readiness checker
    - Path: ttmp/2026/05/26/PR-REVIEW-READY-001--automate-pr-readiness-checks-for-codex-reviews/scripts/01-pr-ready-check.py
      Note: Main non-mutating PR readiness checker implementation
    - Path: ttmp/2026/05/26/PR-REVIEW-READY-001--automate-pr-readiness-checks-for-codex-reviews/scripts/02-trigger-codex-review.sh
      Note: |-
        Helper to add the @codex review trigger comment
        Codex trigger helper
    - Path: ttmp/2026/05/26/PR-REVIEW-READY-001--automate-pr-readiness-checks-for-codex-reviews/scripts/03-watch-codex-reactions.py
      Note: |-
        Polling helper for observing Codex reaction transitions
        Codex reaction polling helper
ExternalSources:
    - https://github.com/go-go-golems/pinocchio/pull/158
Summary: Design and usage notes for scripts that decide whether a PR is ready based on completed checks and Codex review reactions.
LastUpdated: 2026-05-26T13:00:00-04:00
WhatFor: Use this when batching PR readiness checks across many repositories.
WhenToUse: Before merging rollout PRs that require both green CI and a satisfied Codex review.
---


# PR readiness check scripts

## Executive summary

This ticket provides a small script set for checking whether a GitHub pull request is ready to merge under the workflow used in the logcopter rollout: all Actions/status checks must have run and succeeded, and Codex must have finished review with a thumbs-up rather than an in-progress eyes reaction or substantive review comments.

The operator entry point is `scripts/00-pr-ready-check.sh`, a Bash wrapper around `scripts/01-pr-ready-check.py`. It is intentionally non-mutating and exits with status `0` only when the PR is ready. It can be run against a single PR URL or wrapped by future batch scripts that iterate through many repositories.

## Problem statement

When rolling changes through multiple go-go-golems repositories, manual PR readiness checks become repetitive and easy to misread. The desired ready state is stricter than "checks are green": Codex review must also be satisfied. In the observed GitHub UI flow, Codex uses reactions on an `@codex review` trigger comment: an eyes reaction indicates review is running, and a thumbs-up indicates successful review. If Codex leaves substantive comments, the PR is not ready even if CI is green.

A reusable script should therefore answer four questions:

1. Did Actions/status checks run?
2. Are all checks completed with acceptable conclusions?
3. Is there a Codex review signal?
4. Does the newest Codex signal indicate done/satisfied rather than running/unsatisfied?

## Proposed solution

### Script layout

```text
scripts/
  00-pr-ready-check.sh          # bash entry point for operators
  01-pr-ready-check.py          # non-mutating readiness check implementation
  02-trigger-codex-review.sh    # posts '@codex review'
  03-watch-codex-reactions.py   # polls until a Codex signal appears
  04-wait-pr-ready.sh           # waits on one PR; stops on terminal feedback/failures
  05-batch-pr-ready.sh          # one-shot or watch-mode readiness table for many PRs
  06-batch-trigger-codex-review.sh # posts '@codex review' to many PRs
```

### Readiness model

A PR is ready when all of the following are true:

- `statusCheckRollup` contains at least one check/status.
- Every check run is `COMPLETED` and has conclusion `SUCCESS`, `SKIPPED`, or `NEUTRAL`.
- Every legacy status context has state `SUCCESS`.
- A Codex signal exists. A signal is either:
  - a Codex-authored review/comment, matched by configurable author regex; or
  - a human comment whose body is exactly `@codex review`, because Codex reacts to that trigger comment.
- The latest Codex signal has at least one `THUMBS_UP` reaction, or a Codex-authored body explicitly says no major issues and includes a thumbs-up token such as `:+1:`.
- The latest Codex signal has no `EYES` reaction.
- If the latest signal is Codex-authored, its body is empty/benign/satisfied; substantive body text with suggestions means the review likely requested changes or left comments.

`04-wait-pr-ready.sh` treats substantive Codex review text as a terminal wait condition. It exits immediately with status `3` instead of polling until timeout, because those comments require an operator/code change before the PR can become ready.

### Why trigger comments are signals

Initial testing on Pinocchio PR 158 showed that after posting `@codex review`, the PR comment received an `EYES` reaction while Codex was running. That means the relevant in-progress state may be attached to the human trigger comment rather than a bot-authored review object. The script treats the latest exact `@codex review` comment as a Codex signal and checks its reactions.

## Usage examples

Check one PR:

```bash
scripts/00-pr-ready-check.sh https://github.com/go-go-golems/pinocchio/pull/158
```

Machine-readable output:

```bash
scripts/00-pr-ready-check.sh https://github.com/go-go-golems/pinocchio/pull/158 --json
```

The JSON output includes automation-friendly fields:

```json
{
  "ok": false,
  "state": "waiting_checks",
  "terminal": false,
  "failedCheckKinds": ["pending_checks"]
}
```

Known `state` values are `ready`, `waiting_checks`, `waiting_codex`, `no_codex`, `failed_checks`, `codex_feedback`, `not_ready`, and `error`. `terminal=true` means a human/code change is required before waiting can succeed.

Trigger Codex review:

```bash
scripts/02-trigger-codex-review.sh https://github.com/go-go-golems/pinocchio/pull/158
```

Trigger Codex review for a batch:

```bash
scripts/06-batch-trigger-codex-review.sh /tmp/prs.txt
```

Watch for a Codex reaction signal:

```bash
scripts/03-watch-codex-reactions.py https://github.com/go-go-golems/pinocchio/pull/158 --interval 30 --timeout 900
```

Check many PRs without blocking on one PR:

```bash
scripts/05-batch-pr-ready.sh /tmp/prs.txt
```

Watch many PRs until there is something for the operator to do:

```bash
scripts/05-batch-pr-ready.sh /tmp/prs.txt --watch --interval 30 --timeout 1800
```

Watch mode keeps polling only while every PR is still waiting. It stops when all PRs are ready, when any PR reaches a terminal operator-action state (`codex_feedback`, `failed_checks`, or `error`), or when some PR becomes `ready` while others are still waiting. The partial-ready stop exits with code `5` so release-train operators can merge/release in dependency order instead of sleeping through an actionable state.

## Implementation notes

The checker uses `gh api graphql` rather than scraping HTML. It queries:

- `statusCheckRollup.contexts.nodes` for check runs and legacy status contexts;
- `reviews` for Codex-authored review objects;
- `comments` for exact `@codex review` trigger comments;
- `reactionGroups` for `THUMBS_UP` and `EYES` reaction counts.

The default Codex author regex is intentionally broad:

```text
(?i)(^|[-_])(codex|openai-codex|chatgpt)([-_]|$)|codex|openai
```

If the organization standardizes on a specific bot login, pass a narrower value:

```bash
scripts/01-pr-ready-check.py OWNER/REPO#123 --codex-author-regex '^openai-codex\[bot\]$'
```

## Observed Pinocchio test output

After temporarily pushing an intentionally bad workflow change and commenting `@codex review`, the checker observed the in-progress state:

```text
READY: no
FAIL: pending checks: Analyze: status=IN_PROGRESS; lint: status=IN_PROGRESS; GoSec Security Scan: status=IN_PROGRESS
OK: latest Codex signal (codex-trigger) by wesen: https://github.com/go-go-golems/pinocchio/pull/158#issuecomment-4546486328
FAIL: latest Codex signal has no thumbs-up reaction
FAIL: latest Codex signal has 1 eyes reaction(s), review may still be running
OK: latest signal is a human @codex review trigger; body comments are not treated as review findings
```

That verifies the script can detect the eyes/in-progress state. A subsequent Codex review on the intentionally bad commit produced a substantive Codex-authored body, which the checker also treated as not ready. After reverting the bad commit, a later Codex comment said it did not find major issues and included `:+1:`; the checker now treats that body form as a satisfied thumbs-up signal even when GitHub reaction counts do not include `THUMBS_UP`.

## Risks and open questions

- GitHub's GraphQL schema can vary by host/version. The script already had to avoid unsupported `statusCheckRollup(first:)` and unsupported `CheckRun.workflowName` fields.
- Reaction identity is currently summarized by count. If unrelated users add thumbs-up or eyes reactions to the trigger comment, the script could misread the Codex state. The query already fetches reaction user logins for future tightening, but the first version checks counts.
- Codex may sometimes leave review findings as inline review comments rather than a review body. Future versions should query review threads/comments if this becomes common.
- The exact satisfied signal should be revalidated after a full successful Codex review cycle, because this pass verified the running `EYES` state but not yet a completed `THUMBS_UP` state.

## Implementation plan for future batch mode

1. Store a list of PR URLs in a file.
2. Loop over the URLs and run `01-pr-ready-check.py --json`.
3. Emit a table with ready/not-ready status and failed criteria.
4. Only merge PRs where the script exits `0`.
5. For not-ready PRs with no Codex signal, optionally run `02-trigger-codex-review.sh`.
6. For PRs with eyes reactions, wait and rerun later.
