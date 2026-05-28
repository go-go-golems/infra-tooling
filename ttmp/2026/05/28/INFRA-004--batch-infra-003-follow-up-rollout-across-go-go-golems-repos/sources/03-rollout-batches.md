---
Title: INFRA-004 rollout batches
Ticket: INFRA-004
Status: active
Topics:
  - automation
  - cli
  - release
  - docsctl
  - logcopter
  - github
DocType: reference
Intent: short-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Generated human-readable rollout batch plan derived from the INFRA-003 inventory.
LastUpdated: 2026-05-28T00:00:00-04:00
WhatFor: Use as a quick scan of INFRA-004 repository batches and rollout tracks.
WhenToUse: When selecting the next rollout PR wave or explaining batch membership.
---

# INFRA-004 rollout batches

Inventory: `/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/41-repository-follow-up-inventory.json`

## B1 — foundation and upstream libraries

Run first where needed as dependencies; avoid downstream PRs until these are merged/released.

|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|
|---|---:|---:|---:|---:|---|
|`logcopter`|yes|||||
|`bobatea`|yes||||geppetto|
|`common-sense`|yes||yes||glazed|
|`dmeta`|yes||yes||glazed, logcopter|
|`esper`|yes||yes||glazed, logcopter|
|`go-go-os-backend`|yes||||go-go-goja|
|`go-sqlite-regexp`|yes|||||
|`infra-tooling`|yes||yes||glazed, logcopter|
|`sanitize`|yes||yes||glazed|

## B2 — leaf logcopter-only repositories

Low API risk; one baseline PR per repo, run validation, merge before release train tagging.

|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|
|---|---:|---:|---:|---:|---|
|`ai-in-action-app`|yes||||logcopter|
|`barbar`|yes|||||
|`bubble-table`|yes|||||
|`go-go-agent-action`|yes||||logcopter|
|`go-go-app-arc-agi`|yes||||go-go-os-backend, logcopter|
|`go-go-app-sqlite`|yes||||logcopter|
|`go-go-host`|yes||||glazed, go-go-goja, logcopter|
|`oak-git-db`|yes||||logcopter|
|`raza`|yes|||||
|`salad`|yes||||logcopter|
|`terraform-provider-stytch-b2b`|yes|||||
|`voyage`|yes|||||

## B3 — Glazed linting without docsctl

Add glazed-lint with logcopter where safe; no docs publish/Vault work.

|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|
|---|---:|---:|---:|---:|---|
|`markdown-quizz`|yes||yes||glazed|
|`plunger`|yes||yes||clay, glazed|
|`refactorio`|yes||yes||glazed, oak|
|`go-go-app-inventory`|yes||yes||geppetto, glazed, go-go-goja, go-go-os-backend, pinocchio, plz-confirm|

## B4 — docsctl + Glazed CLI leaf packages

Add docsctl release job only after local help export/validate succeeds and Vault role is ready.

|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|
|---|---:|---:|---:|---:|---|
|`almanach`|yes|yes|yes||glazed|
|`biberon`|yes|yes|yes||glazed|
|`bucheron`|yes|yes|yes||glazed|
|`codex-sessions`|yes|yes|yes||glazed|
|`devctl`|yes|yes|yes||glazed|
|`docmgr`|yes|yes|yes||glazed|
|`font-util`|yes|yes|yes||glazed|
|`gitcommit`|yes|yes|yes||glazed|
|`go-emrichen`|yes|yes|yes||glazed|
|`harkonnen`|yes|yes|yes||glazed|
|`openai-mock-server`|yes|yes|yes||glazed|
|`sessionstream`|yes|yes|yes||glazed|
|`tactician`|yes|yes|yes||glazed|
|`cliopatra`|yes|yes|yes||clay, glazed|
|`ecrivain`|yes|yes|yes||clay, glazed|
|`js-analyzer`|yes|yes|yes||glazed, go-go-goja|
|`mastoid`|yes|yes|yes||clay, glazed|
|`parka`|yes|yes|yes||clay, glazed|
|`plz-confirm`|yes|yes|yes||glazed, go-go-goja|
|`prescribe`|yes|yes|yes||geppetto, glazed|
|`prompto`|yes|yes|yes||clay, glazed|
|`remarquee`|yes|yes|yes||geppetto, glazed|
|`uhoh`|yes|yes|yes||clay, glazed|
|`vault-envrc-generator`|yes|yes|yes||clay, glazed|
|`zine-layout`|yes|yes|yes||glazed, go-emrichen|
|`go-go-mcp`|yes|yes|yes||clay, geppetto, glazed|
|`oak`|yes|yes|yes||bobatea, clay, glazed|
|`sqleton`|yes|yes|yes||clay, glazed, parka|
|`form-generator`|yes|yes|yes||clay, glazed, sqleton, uhoh|
|`escuse-me`|yes|yes|yes||clay, geppetto, glazed, go-emrichen, parka|
|`web-agent-example`|yes|yes|yes||clay, geppetto, glazed, go-go-goja, pinocchio|
|`go-go-agent`|yes|yes|yes||bobatea, clay, geppetto, glazed, go-emrichen, pinocchio|
|`jesus`|yes|yes|yes||clay, geppetto, glazed, go-go-goja, go-go-mcp, pinocchio|
|`geppetto`||yes|||glazed, go-emrichen, go-go-goja, logcopter, sanitize|

## B5 — xgoja provider/API-intent candidates

Do not implement provider bindings until API intent is confirmed; may still do logcopter/glazed baseline separately.

|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|
|---|---:|---:|---:|---:|---|
|`cozodb-goja`|yes||yes|yes|glazed, go-go-goja|
|`go-go-gepa`|yes|yes|yes|yes|clay, geppetto, glazed, go-go-goja, go-go-os-backend|
|`go-go-goja`||||yes|bobatea, geppetto, glazed, logcopter|
|`go-minitrace`||||yes|clay, glazed, go-go-goja, logcopter|
|`goja-github-actions`|yes|yes|yes|yes|glazed, go-go-goja|
|`openai-app-server`|yes||yes|yes|glazed, go-go-goja|
|`pinocchio`||||yes|bobatea, clay, geppetto, glazed, go-go-goja, logcopter, sanitize, sessionstream, uhoh|
|`scraper`|yes|yes|yes|yes|glazed, go-go-goja, sessionstream|
|`smailnail`|yes|yes|yes|yes|clay, geppetto, glazed, go-go-goja, go-go-mcp|
|`vm-system`|yes|yes|yes|yes|glazed, go-go-goja|
|`workspace-manager`||||yes|clay, glazed, go-go-goja, logcopter|
