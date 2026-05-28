---
Title: Repository follow-up inventory for logcopter docsctl Glazed linting and xgoja
Ticket: INFRA-003
Status: active
Topics:
    - cli
    - automation
    - release
    - github
    - docsctl
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/scripts/14-inventory-followup-repos.py
      Note: Scanner used to classify Go repositories for rollout follow-up tracks
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/41-repository-follow-up-inventory.json
      Note: Full machine-readable repository inventory
    - Path: ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/42-repository-follow-up-inventory.tsv
      Note: Compact follow-up table
ExternalSources: []
Summary: Repository-wide heuristic inventory of Go-Go-Golems repositories that may need generated logcopter package loggers, docsctl CI/CD publishing, Glazed CLI linting, or xgoja provider bindings.
LastUpdated: 2026-05-28T00:00:00-04:00
WhatFor: Use this inventory to plan follow-up rollout PRs across the Go-Go-Golems repository set.
WhenToUse: When deciding which repositories still need logcopter, docsctl publishing, Glazed linting, or xgoja binding work after INFRA-003.
---


# Repository follow-up inventory for logcopter, docsctl, Glazed linting, and xgoja

## Purpose

This document stores a repository-wide inventory for follow-up rollout work across `/home/manuel/code/wesen/go-go-golems`. It answers which Go repositories appear to need any of four rollout tracks: generated logcopter package loggers, docsctl release publishing, Glazed CLI policy linting, or xgoja bindings/provider adapters.

The inventory is heuristic and intended as a planning list, not as an implementation order. Each row should still be confirmed before opening a PR, especially for docsctl and xgoja where repository intent matters.

## Methodology

- Scanned first-level repositories under `/home/manuel/code/wesen/go-go-golems` that contain `go.mod`.
- `logcopter addition` means the repo lacks at least one of: `logcopter_generate.go`, checked-in generated `**/logcopter.go`, or `make logcopter-check`.
- `docsctl CI/CD + push` means the repo appears to be a Glazed command/help repository (`cmd`, Glazed dependency, help/export markers) but has no release workflow calling `publish-docsctl`/`docsctl publish`.
- `Glazed linting added` means the repo imports `github.com/go-go-golems/glazed` but its Makefile does not expose `glazed-lint`.
- `xgoja bindings` means the repo imports `go-go-goja`, contains module/runtime registration markers, and does not have a recognized provider directory such as `pkg/xgoja/provider` or `pkg/js/modules/<name>/provider`.

Source artifacts:

- `scripts/14-inventory-followup-repos.py` — scanner script.
- `sources/41-repository-follow-up-inventory.json` — full machine-readable scan.
- `sources/42-repository-follow-up-inventory.tsv` — compact table of repos with at least one follow-up flag.

## Summary counts

- Go repositories scanned: **77**
- Repositories with at least one follow-up flag: **70**
- Logcopter addition candidates: **65**
- Docsctl CI/CD + publish candidates: **39**
- Glazed linting candidates: **49**
- xgoja binding candidates: **11**

## Combined follow-up list

|repo|needs_logcopter_addition|needs_docsctl_cicd_push|needs_glazed_linting_added|needs_xgoja_bindings|module|
|---|---|---|---|---|---|
|`ai-in-action-app`|yes||||`github.com/go-go-golems/ai-in-action-app`|
|`almanach`|yes|yes|yes||`github.com/go-go-golems/almanach`|
|`barbar`|yes||||`github.com/go-go-golems/barbar`|
|`biberon`|yes|yes|yes||`github.com/go-go-golems/biberon`|
|`bobatea`|yes||||`github.com/go-go-golems/bobatea`|
|`bubble-table`|yes||||`github.com/evertras/bubble-table`|
|`bucheron`|yes|yes|yes||`github.com/go-go-golems/bucheron`|
|`cliopatra`|yes|yes|yes||`github.com/go-go-golems/cliopatra`|
|`codex-sessions`|yes|yes|yes||`github.com/go-go-golems/codex-session`|
|`common-sense`|yes||yes||`github.com/go-go-golems/common-sense`|
|`cozodb-goja`|yes||yes|yes|`github.com/go-go-golems/cozodb-goja`|
|`devctl`|yes|yes|yes||`github.com/go-go-golems/devctl`|
|`dmeta`|yes||yes||`github.com/go-go-golems/dmeta`|
|`docmgr`|yes|yes|yes||`github.com/go-go-golems/docmgr`|
|`ecrivain`|yes|yes|yes||`github.com/go-go-golems/ecrivain`|
|`escuse-me`|yes|yes|yes||`github.com/go-go-golems/escuse-me`|
|`esper`|yes||yes||`github.com/go-go-golems/esper`|
|`font-util`|yes|yes|yes||`github.com/go-go-golems/font-util`|
|`form-generator`|yes|yes|yes||`github.com/go-go-golems/form-generator`|
|`geppetto`||yes|||`github.com/go-go-golems/geppetto`|
|`gitcommit`|yes|yes|yes||`github.com/go-go-golems/gitcommit`|
|`go-emrichen`|yes|yes|yes||`github.com/go-go-golems/go-emrichen`|
|`go-go-agent`|yes|yes|yes||`github.com/go-go-golems/go-go-agent`|
|`go-go-agent-action`|yes||||`github.com/go-go-golems/go-go-agent-action`|
|`go-go-app-arc-agi`|yes||||`github.com/go-go-golems/go-go-app-arc-agi`|
|`go-go-app-inventory`|yes||yes||`github.com/go-go-golems/go-go-app-inventory`|
|`go-go-app-sqlite`|yes||||`github.com/go-go-golems/go-go-app-sqlite`|
|`go-go-gepa`|yes|yes|yes|yes|`github.com/go-go-golems/go-go-gepa`|
|`go-go-goja`||||yes|`github.com/go-go-golems/go-go-goja`|
|`go-go-host`|yes||||`github.com/go-go-golems/XXX`|
|`go-go-mcp`|yes|yes|yes||`github.com/go-go-golems/go-go-mcp`|
|`go-go-os-backend`|yes||||`github.com/go-go-golems/go-go-os-backend`|
|`go-minitrace`||||yes|`github.com/go-go-golems/go-minitrace`|
|`go-sqlite-regexp`|yes||||`github.com/go-go-golems/go-sqlite-regexp`|
|`goja-github-actions`|yes|yes|yes|yes|`github.com/go-go-golems/goja-github-actions`|
|`harkonnen`|yes|yes|yes||`github.com/go-go-golems/harkonnen`|
|`infra-tooling`|yes||yes||`github.com/go-go-golems/infra-tooling`|
|`jesus`|yes|yes|yes||`github.com/go-go-golems/jesus`|
|`js-analyzer`|yes|yes|yes||`github.com/go-go-golems/js-analyzer`|
|`logcopter`|yes||||`github.com/go-go-golems/logcopter`|
|`markdown-quizz`|yes||yes||`github.com/go-go-golems/XXX`|
|`mastoid`|yes|yes|yes||`github.com/go-go-golems/mastoid`|
|`oak`|yes|yes|yes||`github.com/go-go-golems/oak`|
|`oak-git-db`|yes||||`github.com/go-go-golems/oak-git-db`|
|`openai-app-server`|yes||yes|yes|`github.com/go-go-golems/openai-app-server`|
|`openai-mock-server`|yes|yes|yes||`mock-openai-server`|
|`parka`|yes|yes|yes||`github.com/go-go-golems/parka`|
|`pinocchio`||||yes|`github.com/go-go-golems/pinocchio`|
|`plunger`|yes||yes||`github.com/go-go-golems/plunger`|
|`plz-confirm`|yes|yes|yes||`github.com/go-go-golems/plz-confirm`|
|`prescribe`|yes|yes|yes||`github.com/go-go-golems/prescribe`|
|`prompto`|yes|yes|yes||`github.com/go-go-golems/prompto`|
|`raza`|yes||||`github.com/wesen/raza`|
|`refactorio`|yes||yes||`github.com/go-go-golems/refactorio`|
|`remarquee`|yes|yes|yes||`github.com/go-go-golems/remarquee`|
|`salad`|yes||||`github.com/go-go-golems/salad`|
|`sanitize`|yes||yes||`github.com/go-go-golems/sanitize`|
|`scraper`|yes|yes|yes|yes|`github.com/go-go-golems/scraper`|
|`sessionstream`|yes|yes|yes||`github.com/go-go-golems/sessionstream`|
|`smailnail`|yes|yes|yes|yes|`github.com/go-go-golems/smailnail`|
|`sqleton`|yes|yes|yes||`github.com/go-go-golems/sqleton`|
|`tactician`|yes|yes|yes||`github.com/go-go-golems/tactician`|
|`terraform-provider-stytch-b2b`|yes||||`github.com/mento/terraform-provider-stytch-b2b`|
|`uhoh`|yes|yes|yes||`github.com/go-go-golems/uhoh`|
|`vault-envrc-generator`|yes|yes|yes||`github.com/go-go-golems/vault-envrc-generator`|
|`vm-system`|yes|yes|yes|yes|`github.com/go-go-golems/vm-system`|
|`voyage`|yes||||`github.com/go-go-golems/voyage`|
|`web-agent-example`|yes|yes|yes||`github.com/go-go-golems/web-agent-example`|
|`workspace-manager`||||yes|`github.com/go-go-golems/workspace-manager`|
|`zine-layout`|yes|yes|yes||`github.com/go-go-golems/zine-layout`|

## Logcopter addition candidates

|repo|module|notes|
|---|---|---|
|`ai-in-action-app`|`github.com/go-go-golems/ai-in-action-app`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`almanach`|`github.com/go-go-golems/almanach`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`barbar`|`github.com/go-go-golems/barbar`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`biberon`|`github.com/go-go-golems/biberon`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`bobatea`|`github.com/go-go-golems/bobatea`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`bubble-table`|`github.com/evertras/bubble-table`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`bucheron`|`github.com/go-go-golems/bucheron`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`cliopatra`|`github.com/go-go-golems/cliopatra`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`codex-sessions`|`github.com/go-go-golems/codex-session`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`common-sense`|`github.com/go-go-golems/common-sense`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`cozodb-goja`|`github.com/go-go-golems/cozodb-goja`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`devctl`|`github.com/go-go-golems/devctl`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`dmeta`|`github.com/go-go-golems/dmeta`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`docmgr`|`github.com/go-go-golems/docmgr`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`ecrivain`|`github.com/go-go-golems/ecrivain`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`escuse-me`|`github.com/go-go-golems/escuse-me`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`esper`|`github.com/go-go-golems/esper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`font-util`|`github.com/go-go-golems/font-util`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`form-generator`|`github.com/go-go-golems/form-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`gitcommit`|`github.com/go-go-golems/gitcommit`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-emrichen`|`github.com/go-go-golems/go-emrichen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-agent`|`github.com/go-go-golems/go-go-agent`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-agent-action`|`github.com/go-go-golems/go-go-agent-action`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`go-go-app-arc-agi`|`github.com/go-go-golems/go-go-app-arc-agi`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`go-go-app-inventory`|`github.com/go-go-golems/go-go-app-inventory`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`go-go-app-sqlite`|`github.com/go-go-golems/go-go-app-sqlite`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`go-go-gepa`|`github.com/go-go-golems/go-go-gepa`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`go-go-host`|`github.com/go-go-golems/XXX`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`go-go-mcp`|`github.com/go-go-golems/go-go-mcp`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-os-backend`|`github.com/go-go-golems/go-go-os-backend`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`go-sqlite-regexp`|`github.com/go-go-golems/go-sqlite-regexp`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`goja-github-actions`|`github.com/go-go-golems/goja-github-actions`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`harkonnen`|`github.com/go-go-golems/harkonnen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`infra-tooling`|`github.com/go-go-golems/infra-tooling`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`jesus`|`github.com/go-go-golems/jesus`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`js-analyzer`|`github.com/go-go-golems/js-analyzer`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`logcopter`|`github.com/go-go-golems/logcopter`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`|
|`markdown-quizz`|`github.com/go-go-golems/XXX`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`mastoid`|`github.com/go-go-golems/mastoid`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`oak`|`github.com/go-go-golems/oak`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`oak-git-db`|`github.com/go-go-golems/oak-git-db`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`openai-app-server`|`github.com/go-go-golems/openai-app-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`openai-mock-server`|`mock-openai-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`parka`|`github.com/go-go-golems/parka`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`plunger`|`github.com/go-go-golems/plunger`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`plz-confirm`|`github.com/go-go-golems/plz-confirm`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prescribe`|`github.com/go-go-golems/prescribe`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prompto`|`github.com/go-go-golems/prompto`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`raza`|`github.com/wesen/raza`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`refactorio`|`github.com/go-go-golems/refactorio`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`remarquee`|`github.com/go-go-golems/remarquee`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`salad`|`github.com/go-go-golems/salad`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`sanitize`|`github.com/go-go-golems/sanitize`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`scraper`|`github.com/go-go-golems/scraper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sessionstream`|`github.com/go-go-golems/sessionstream`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`smailnail`|`github.com/go-go-golems/smailnail`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sqleton`|`github.com/go-go-golems/sqleton`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`tactician`|`github.com/go-go-golems/tactician`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`terraform-provider-stytch-b2b`|`github.com/mento/terraform-provider-stytch-b2b`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`uhoh`|`github.com/go-go-golems/uhoh`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vault-envrc-generator`|`github.com/go-go-golems/vault-envrc-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vm-system`|`github.com/go-go-golems/vm-system`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`voyage`|`github.com/go-go-golems/voyage`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`|
|`web-agent-example`|`github.com/go-go-golems/web-agent-example`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`zine-layout`|`github.com/go-go-golems/zine-layout`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|

## Docsctl CI/CD + publish candidates

|repo|module|notes|
|---|---|---|
|`almanach`|`github.com/go-go-golems/almanach`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`biberon`|`github.com/go-go-golems/biberon`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`bucheron`|`github.com/go-go-golems/bucheron`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`cliopatra`|`github.com/go-go-golems/cliopatra`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`codex-sessions`|`github.com/go-go-golems/codex-session`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`devctl`|`github.com/go-go-golems/devctl`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`docmgr`|`github.com/go-go-golems/docmgr`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`ecrivain`|`github.com/go-go-golems/ecrivain`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`escuse-me`|`github.com/go-go-golems/escuse-me`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`font-util`|`github.com/go-go-golems/font-util`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`form-generator`|`github.com/go-go-golems/form-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`geppetto`|`github.com/go-go-golems/geppetto`|docsctl candidate: command/help markers but no publish workflow|
|`gitcommit`|`github.com/go-go-golems/gitcommit`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-emrichen`|`github.com/go-go-golems/go-emrichen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-agent`|`github.com/go-go-golems/go-go-agent`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-gepa`|`github.com/go-go-golems/go-go-gepa`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`go-go-mcp`|`github.com/go-go-golems/go-go-mcp`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`goja-github-actions`|`github.com/go-go-golems/goja-github-actions`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`harkonnen`|`github.com/go-go-golems/harkonnen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`jesus`|`github.com/go-go-golems/jesus`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`js-analyzer`|`github.com/go-go-golems/js-analyzer`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`mastoid`|`github.com/go-go-golems/mastoid`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`oak`|`github.com/go-go-golems/oak`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`openai-mock-server`|`mock-openai-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`parka`|`github.com/go-go-golems/parka`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`plz-confirm`|`github.com/go-go-golems/plz-confirm`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prescribe`|`github.com/go-go-golems/prescribe`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prompto`|`github.com/go-go-golems/prompto`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`remarquee`|`github.com/go-go-golems/remarquee`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`scraper`|`github.com/go-go-golems/scraper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sessionstream`|`github.com/go-go-golems/sessionstream`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`smailnail`|`github.com/go-go-golems/smailnail`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sqleton`|`github.com/go-go-golems/sqleton`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`tactician`|`github.com/go-go-golems/tactician`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`uhoh`|`github.com/go-go-golems/uhoh`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vault-envrc-generator`|`github.com/go-go-golems/vault-envrc-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vm-system`|`github.com/go-go-golems/vm-system`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`web-agent-example`|`github.com/go-go-golems/web-agent-example`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`zine-layout`|`github.com/go-go-golems/zine-layout`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|

## Glazed linting candidates

|repo|module|notes|
|---|---|---|
|`almanach`|`github.com/go-go-golems/almanach`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`biberon`|`github.com/go-go-golems/biberon`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`bucheron`|`github.com/go-go-golems/bucheron`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`cliopatra`|`github.com/go-go-golems/cliopatra`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`codex-sessions`|`github.com/go-go-golems/codex-session`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`common-sense`|`github.com/go-go-golems/common-sense`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`cozodb-goja`|`github.com/go-go-golems/cozodb-goja`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`devctl`|`github.com/go-go-golems/devctl`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`dmeta`|`github.com/go-go-golems/dmeta`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`docmgr`|`github.com/go-go-golems/docmgr`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`ecrivain`|`github.com/go-go-golems/ecrivain`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`escuse-me`|`github.com/go-go-golems/escuse-me`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`esper`|`github.com/go-go-golems/esper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`font-util`|`github.com/go-go-golems/font-util`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`form-generator`|`github.com/go-go-golems/form-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`gitcommit`|`github.com/go-go-golems/gitcommit`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-emrichen`|`github.com/go-go-golems/go-emrichen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-agent`|`github.com/go-go-golems/go-go-agent`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`go-go-app-inventory`|`github.com/go-go-golems/go-go-app-inventory`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`go-go-gepa`|`github.com/go-go-golems/go-go-gepa`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`go-go-mcp`|`github.com/go-go-golems/go-go-mcp`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`goja-github-actions`|`github.com/go-go-golems/goja-github-actions`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`harkonnen`|`github.com/go-go-golems/harkonnen`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`infra-tooling`|`github.com/go-go-golems/infra-tooling`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`jesus`|`github.com/go-go-golems/jesus`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`js-analyzer`|`github.com/go-go-golems/js-analyzer`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`markdown-quizz`|`github.com/go-go-golems/XXX`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`mastoid`|`github.com/go-go-golems/mastoid`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`oak`|`github.com/go-go-golems/oak`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`openai-app-server`|`github.com/go-go-golems/openai-app-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`openai-mock-server`|`mock-openai-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`parka`|`github.com/go-go-golems/parka`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`plunger`|`github.com/go-go-golems/plunger`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`plz-confirm`|`github.com/go-go-golems/plz-confirm`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prescribe`|`github.com/go-go-golems/prescribe`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`prompto`|`github.com/go-go-golems/prompto`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`refactorio`|`github.com/go-go-golems/refactorio`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`remarquee`|`github.com/go-go-golems/remarquee`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`sanitize`|`github.com/go-go-golems/sanitize`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`|
|`scraper`|`github.com/go-go-golems/scraper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sessionstream`|`github.com/go-go-golems/sessionstream`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`smailnail`|`github.com/go-go-golems/smailnail`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`sqleton`|`github.com/go-go-golems/sqleton`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`tactician`|`github.com/go-go-golems/tactician`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`uhoh`|`github.com/go-go-golems/uhoh`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vault-envrc-generator`|`github.com/go-go-golems/vault-envrc-generator`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`vm-system`|`github.com/go-go-golems/vm-system`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`web-agent-example`|`github.com/go-go-golems/web-agent-example`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|
|`zine-layout`|`github.com/go-go-golems/zine-layout`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`|

## xgoja binding candidates

|repo|module|notes|
|---|---|---|
|`cozodb-goja`|`github.com/go-go-golems/cozodb-goja`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`go-go-gepa`|`github.com/go-go-golems/go-go-gepa`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`go-go-goja`|`github.com/go-go-golems/go-go-goja`|uses go-go-goja module registration/runtime markers but no provider dir|
|`go-minitrace`|`github.com/go-go-golems/go-minitrace`|uses go-go-goja module registration/runtime markers but no provider dir|
|`goja-github-actions`|`github.com/go-go-golems/goja-github-actions`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`openai-app-server`|`github.com/go-go-golems/openai-app-server`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`pinocchio`|`github.com/go-go-golems/pinocchio`|uses go-go-goja module registration/runtime markers but no provider dir|
|`scraper`|`github.com/go-go-golems/scraper`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`smailnail`|`github.com/go-go-golems/smailnail`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`vm-system`|`github.com/go-go-golems/vm-system`|logcopter: no `logcopter_generate.go`, no `make logcopter-check`, no generated `logcopter.go`<br>docsctl candidate: command/help markers but no publish workflow<br>uses Glazed but no `make glazed-lint`<br>uses go-go-goja module registration/runtime markers but no provider dir|
|`workspace-manager`|`github.com/go-go-golems/workspace-manager`|uses go-go-goja module registration/runtime markers but no provider dir|

## Repositories that did not trigger follow-up flags

|repo|module|
|---|---|
|`clay`|`github.com/go-go-golems/clay`|
|`css-visual-diff`|`github.com/go-go-golems/css-visual-diff`|
|`discord-bot`|`github.com/go-go-golems/discord-bot`|
|`glazed`|`github.com/go-go-golems/glazed`|
|`go-template`|`github.com/go-go-golems/XXX`|
|`goja-git`|`github.com/go-go-golems/goja-git`|
|`loupedeck`|`github.com/go-go-golems/loupedeck`|

## Caveats

- This scan intentionally does not inspect non-Go repositories.
- Logcopter is broad: many libraries/tools may not need generated package loggers immediately, but they lack the rollout baseline.
- Docsctl candidates require manual confirmation that the CLI has useful help sections and that Terraform/Vault roles should be created.
- xgoja candidates require manual API design; some repos may only consume go-go-goja internally and not need a public provider.
