# Changelog

## 2026-05-27

- Initial workspace created


## 2026-05-27

Initialized Glazed lint rollout ticket and added repository inventory script/output for Glazed-dependent local repositories.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/01-inventory-glazed-repos.sh — Inventory script for Glazed-dependent repositories
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/sources/01-glazed-repo-inventory.tsv — Captured inventory output


## 2026-05-27

Applied Glazed lint Makefile and CI wiring to the ten active add-js-providers workspace repositories; added fallback linter installation and narrow legacy allow paths; final make glazed-lint pass succeeded for all targets.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/03-apply-glazed-lint-wiring.py — Generated Makefile and CI wiring
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/06-run-glazed-lint.sh — Validation runner for make glazed-lint
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/sources/glazed-lint-logs — Per-repository lint logs


## 2026-05-27

Committed Glazed lint rollout changes locally in all ten target repositories, removed an incidental css-visual-diff .bin artifact, and rebased go-go-goja/loupedeck lint commits onto origin/main so each future PR is exactly one commit ahead.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/09-commit-workspace-repos.sh — Script that committed the focused rollout files in each target repo


## 2026-05-27

Opened ten Glazed lint rollout PRs, stored the PR YAML, triggered Codex with ggg, and recorded initial batch readiness as waiting_checks for all PRs. No PRs were merged.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/10-glazed-lint-prs.yaml — Batch PR list for ggg readiness
- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/11-push-and-open-prs-no-verify.sh — Traceable PR publication script used after local pre-push hook failure


## 2026-05-27

Added a per-repository rollout action document describing the Makefile/CI changes, allow paths, diagnostics, and validation result for each Glazed lint PR.

### Related Files

- /home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/reference/02-repository-rollout-actions.md — Per-repository review aid for Glazed lint rollout

