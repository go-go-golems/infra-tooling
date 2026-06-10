#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-/home/manuel/workspaces/2026-06-10/add-docs-deploy}"
OUT_DIR="${2:-$ROOT/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/sources/help-export-inventory}"
mkdir -p "$OUT_DIR"
SUMMARY="$OUT_DIR/summary.tsv"
: > "$SUMMARY"
printf 'repo\tmodule\tstatus\tpackage\tcmd_dir\texport_command\tnote\n' >> "$SUMMARY"

run_candidate() {
  local repo="$1" package="$2" cmd_dir="$3" export_cmd="$4" note="${5:-}"
  local repo_dir="$ROOT/$repo"
  local log="$OUT_DIR/${repo}.log"
  rm -rf "$repo_dir/.docsctl"
  mkdir -p "$repo_dir/.docsctl"
  set +e
  (cd "$repo_dir" && timeout 180 bash -lc "$export_cmd" && test -s .docsctl/help.sqlite && docsctl validate --file .docsctl/help.sqlite --package "$package" --version v0.0.0-local) >"$log" 2>&1
  local code=$?
  set -e
  local status="ready"
  if [[ $code -ne 0 ]]; then
    status="needs-work(code=$code)"
  fi
  local module=""
  [[ -f "$repo_dir/go.mod" ]] && module=$(awk '/^module /{print $2; exit}' "$repo_dir/go.mod")
  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$repo" "$module" "$status" "$package" "$cmd_dir" "$export_cmd" "$note" >> "$SUMMARY"
}

# Workspace packages; excludes glazed (already set up) and infra-tooling (not part of this rollout target).
run_candidate devctl devctl ./cmd/devctl 'GOWORK=off go run ./cmd/devctl help export --format sqlite --output-path .docsctl/help.sqlite'
run_candidate docmgr docmgr ./cmd/docmgr 'GOWORK=off go run ./cmd/docmgr help export --format sqlite --output-path .docsctl/help.sqlite' 'already has release-coupled docs job; keep as reference/baseline'
run_candidate goja-bleve goja-bleve ./cmd/goja-bleve 'mkdir -p .docsctl && (cd cmd/goja-bleve && GOWORK=off go run . help export --format sqlite --output-path ../../.docsctl/help.sqlite)' 'already has separate publish-docs workflow; nested command module'
run_candidate llm-proxy llm-proxy ./cmd/llm-proxy-server 'GOWORK=off go run ./cmd/llm-proxy-server help export --format sqlite --output-path .docsctl/help.sqlite' 'currently stdlib flag/http server, expected to need CLI/help wiring first'
run_candidate logcopter logcopter ./cmd/logcopter-gen 'GOWORK=off go run ./cmd/logcopter-gen help export --format sqlite --output-path .docsctl/help.sqlite' 'currently stdlib flag generator, expected to need CLI/help wiring first'
run_candidate react-chat chat-overlay ./cmd/chat-overlay 'GOWORK=off go run ./cmd/chat-overlay help export --format sqlite --output-path .docsctl/help.sqlite'
run_candidate remarquee remarquee ./cmd/remarquee 'GOWORK=off go run ./cmd/remarquee help export --format sqlite --output-path .docsctl/help.sqlite' 'help command is cgo-gated; CI default must keep cgo or use build tags deliberately'
run_candidate scraper scraper ./cmd/scraper 'GOWORK=off go run ./cmd/scraper help export --format sqlite --output-path .docsctl/help.sqlite'
run_candidate sessionstream sessionstream ./cmd/sessionstream-systemlab 'GOWORK=off go run ./cmd/sessionstream-systemlab help export --format sqlite --output-path .docsctl/help.sqlite' 'published package can be sessionstream while binary is sessionstream-systemlab'
run_candidate vm-system vm-system ./cmd/vm-system 'GOWORK=off go run ./cmd/vm-system help export --format sqlite --output-path .docsctl/help.sqlite'

column -t -s $'\t' "$SUMMARY" > "$OUT_DIR/summary.txt"
cat "$OUT_DIR/summary.txt"
