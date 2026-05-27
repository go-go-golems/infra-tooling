---
Title: Glazed linting rollout plan
Ticket: INFRA-002
Status: active
Topics:
    - cli
    - automation
    - release
    - github
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: docs/go-go-golems/glazed-linting-rollout-playbook.md
      Note: |-
        Source playbook for Makefile, CI, and validation policy
        Source playbook for Glazed lint rollout policy
    - Path: ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/01-inventory-glazed-repos.sh
      Note: Inventory script created for this rollout
ExternalSources: []
Summary: Plan for applying Glazed CLI policy linting across go-go-golems repositories and releasing the resulting changes.
LastUpdated: 2026-05-27T11:20:00-04:00
WhatFor: Use as the rollout control document for INFRA-002.
WhenToUse: Before changing repositories, opening PRs, or releasing Glazed lint rollout patches.
---


# Glazed linting rollout plan

## Executive summary

This ticket rolls out the Glazed CLI policy linter to go-go-golems repositories that depend on `github.com/go-go-golems/glazed`. The source policy is documented in `docs/go-go-golems/glazed-linting-rollout-playbook.md`. The rollout will add `glazed-lint-build` and `glazed-lint` Makefile targets, wire the analyzer into existing lint targets and CI where appropriate, fix or narrowly allow existing diagnostics, then use `ggg` to manage PR readiness and releases.

## Problem statement

Glazed command policy is currently enforced inconsistently. Repositories can define raw Cobra flags, read raw environment variables in CLI paths, or expose output commands without `RunIntoGlazeProcessor`. The analyzer in `github.com/go-go-golems/glazed/cmd/tools/glazed-lint` makes those policies executable, but each repository must build and run it consistently in local lint and CI.

## Rollout invariants

- Only target repositories that actually depend on `github.com/go-go-golems/glazed`, unless a repo should add Glazed as part of this rollout.
- Keep helper scripts in this ticket's `scripts/` directory.
- Keep repository changes focused: Makefile/CI lint wiring plus necessary policy fixes or narrow allow paths.
- Prefer fixing real user-facing CLI violations over allow-listing them.
- Use narrow allow paths for intentional legacy bridge code, internal generator tools, or non-user-facing command helpers.
- Validate each repo locally before opening a PR.
- Use `ggg pr codex-trigger`, `ggg pr ready`, `ggg pr codex-comments`, and `ggg batch ready` for PR operations.
- Use `ggg release tag-patch --dry-run` before any release tag push.

## Standard Makefile target

The default target comes from the playbook and will be adapted per repository package layout:

```make
GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell GOWORK=off go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/

.PHONY: glazed-lint-build glazed-lint

glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./cmd/... ./pkg/...
```

## Target repository inventory

The first inventory pass is stored at:

```text
ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/sources/01-glazed-repo-inventory.tsv
```

The scan found many local repositories that depend on Glazed under both `/home/manuel/code/wesen/go-go-golems` and the recent xgoja workspace `/home/manuel/workspaces/2026-05-24/add-js-providers`. Several repositories already have `glazed-lint-build` / `glazed-lint` targets, including `glazed`, `clay`, `geppetto`, and `pinocchio`.

The rollout should prioritize canonical repositories under `/home/manuel/code/wesen/go-go-golems` and avoid duplicate workspace checkouts unless the workspace checkout is the intended active branch for a repo. Dirty repositories require inspection before edits.

Initial observations:

- `infra-tooling` has no Makefile yet, so it needs a first Makefile if we want local `make glazed-lint` there.
- `almanach` is dirty and should be inspected before modification.
- `css-visual-diff` appears dirty in the xgoja workspace but clean in the canonical go-go-golems checkout.
- `corporate-headquarters/workspace-manager.backup` is a backup checkout and should be excluded.
- `markdown-quizz` still has module path `github.com/go-go-golems/XXX`; exclude until the module identity is fixed.
- Nested or experimental repos without Makefiles (`dmeta`, `js-analyzer`, `vm-system`, ESP32 subrepos) should be handled separately rather than mixed into the first automated rollout.

## Implementation plan

1. Inventory repositories and choose target set.
2. Add Makefile targets per repo.
3. Wire into lint/lintmax and CI where appropriate.
4. Run `make glazed-lint` and triage diagnostics.
5. Validate each repo.
6. Commit, push, and open PRs.
7. Store PRs in ticket YAML and use `ggg` for readiness.
8. Merge and release patch versions after readiness.

## Open questions

- Whether to target only recently active xgoja rollout repositories or every local go-go-golems repository that depends on Glazed.
- Whether release tags should be pushed immediately after each merge or paused for manual confirmation.
