---
Title: ggg rollout generated status report
Ticket: INFRA-002
Status: active
Topics:
  - cli
  - automation
DocType: source
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Generated Markdown report from ggg rollout report for the INFRA-002 rollout config.
LastUpdated: 2026-05-27T14:20:00-04:00
WhatFor: Preserve generated rollout status/report output as validation evidence.
WhenToUse: Use while reviewing the first ggg rollout implementation and INFRA-002 PR state.
---

# Rollout report: glazed-lint-rollout

- ID: `INFRA-002`
- Workspace: `/home/manuel/workspaces/2026-05-24/add-js-providers`
- Branch: `infra-002/glazed-lint`
- Base: `origin/main`

## Targets

| Repo | Module | Glazed | Branch | Ahead | Dirty | Makefile targets |
| --- | --- | --- | --- | ---: | --- | --- |
| `css-visual-diff` | `github.com/go-go-golems/css-visual-diff` | `v1.3.0` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `discord-bot` | `github.com/go-go-golems/discord-bot` | `v1.2.6` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `geppetto` | `github.com/go-go-golems/geppetto` | `v1.3.0` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `glazed` | `github.com/go-go-golems/glazed` | `` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `go-go-goja` | `github.com/go-go-golems/go-go-goja` | `v1.2.5` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `goja-git` | `github.com/go-go-golems/goja-git` | `v1.2.5` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `go-minitrace` | `github.com/go-go-golems/go-minitrace` | `v1.2.5` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `loupedeck` | `github.com/go-go-golems/loupedeck` | `v1.2.5` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `pinocchio` | `github.com/go-go-golems/pinocchio` | `v1.3.0` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |
| `workspace-manager` | `github.com/go-go-golems/workspace-manager` | `v1.3.4` | `infra-002/glazed-lint` | 1 | clean | `lint, lintmax, glazed-lint, glazed-lint-build` |

## Branch checks

- ✅ `css-visual-diff`: branch state matches rollout expectations
- ✅ `discord-bot`: branch state matches rollout expectations
- ✅ `geppetto`: branch state matches rollout expectations
- ✅ `glazed`: branch state matches rollout expectations
- ✅ `go-go-goja`: branch state matches rollout expectations
- ✅ `goja-git`: branch state matches rollout expectations
- ✅ `go-minitrace`: branch state matches rollout expectations
- ✅ `loupedeck`: branch state matches rollout expectations
- ✅ `pinocchio`: branch state matches rollout expectations
- ✅ `workspace-manager`: branch state matches rollout expectations

## Validation commands

- `glazed-lint`: `make glazed-lint`

Logs: `/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/sources/ggg-rollout-logs`

## Pull requests

PR list from `/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-002--roll-out-glazed-cli-policy-linting-across-go-go-golems-repositories/scripts/10-glazed-lint-prs.yaml`:

```yaml
prs:
  - https://github.com/go-go-golems/css-visual-diff/pull/9
  - https://github.com/go-go-golems/discord-bot/pull/10
  - https://github.com/go-go-golems/geppetto/pull/363
  - https://github.com/go-go-golems/glazed/pull/582
  - https://github.com/go-go-golems/go-go-goja/pull/42
  - https://github.com/go-go-golems/goja-git/pull/3
  - https://github.com/go-go-golems/go-minitrace/pull/12
  - https://github.com/go-go-golems/loupedeck/pull/4
  - https://github.com/go-go-golems/pinocchio/pull/161
  - https://github.com/go-go-golems/workspace-manager/pull/21
```

## Next steps

1. Run validation and inspect failed logs.
2. Verify each branch is focused and based on `origin/main`.
3. Push/open PRs and run `ggg batch ready` on the generated PR list.
