#!/usr/bin/env bash
set -euo pipefail

# Check readiness for multiple PRs without blocking on a single PR.
# Usage:
#   05-batch-pr-ready.sh <prs-file> [--trigger-missing-codex] [--watch] [--interval SECONDS] [--timeout SECONDS]
#
# The PR file contains one PR URL or owner/repo#number per line. Blank lines and
# lines beginning with # are ignored.

if [[ $# -lt 1 ]]; then
  echo "usage: $0 <prs-file> [--trigger-missing-codex] [--watch] [--interval SECONDS] [--timeout SECONDS]" >&2
  exit 2
fi

PRS_FILE="$1"
shift
TRIGGER_MISSING="false"
WATCH="false"
INTERVAL=30
TIMEOUT=1800

while [[ $# -gt 0 ]]; do
  case "$1" in
    --trigger-missing-codex) TRIGGER_MISSING="true"; shift ;;
    --watch) WATCH="true"; shift ;;
    --interval) INTERVAL="$2"; shift 2 ;;
    --timeout) TIMEOUT="$2"; shift 2 ;;
    *) echo "unknown option: $1" >&2; exit 2 ;;
  esac
done

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/00-pr-ready-check.sh"
TRIGGER="$SCRIPT_DIR/02-trigger-codex-review.sh"

if [[ ! -f "$PRS_FILE" ]]; then
  echo "PR file not found: $PRS_FILE" >&2
  exit 2
fi

json_get() {
  python3 -c 'import json,sys; data=json.load(open(sys.argv[1])); v=data.get(sys.argv[2], ""); print(",".join(v) if isinstance(v, list) else v)' "$1" "$2"
}

read_prs() {
  while IFS= read -r pr || [[ -n "$pr" ]]; do
    pr="${pr%%#*}"
    pr="$(printf '%s' "$pr" | xargs)"
    [[ -z "$pr" ]] && continue
    printf '%s\n' "$pr"
  done < "$PRS_FILE"
}

run_once() {
  local tmpdir ready not_ready codex_feedback failed_checks errors terminal_not_ready
  tmpdir="$(mktemp -d)"
  ready=0; not_ready=0; codex_feedback=0; failed_checks=0; errors=0; terminal_not_ready=0

  printf '%-18s %-20s %s\n' STATUS FAILED_CHECKS PR
  printf '%-18s %-20s %s\n' '------' '-------------' '--'

  while IFS= read -r pr; do
    out="$tmpdir/out.$RANDOM.json"
    err="$tmpdir/err.$RANDOM.txt"
    if "$CHECK" "$pr" --json >"$out" 2>"$err"; then
      status="ready"
    elif [[ -s "$out" ]]; then
      status="$(json_get "$out" state)"
    else
      status="error"
      errors=$((errors + 1))
    fi

    failed_kinds=""
    if [[ -s "$out" ]]; then
      failed_kinds="$(json_get "$out" failedCheckKinds)"
    fi

    if [[ "$status" == "no_codex" && "$TRIGGER_MISSING" == "true" ]]; then
      if "$TRIGGER" "$pr" >/dev/null 2>&1; then
        status="codex_triggered"
      else
        status="codex_trigger_failed"
        errors=$((errors + 1))
      fi
    fi

    case "$status" in
      ready) ready=$((ready + 1)) ;;
      codex_feedback) codex_feedback=$((codex_feedback + 1)); not_ready=$((not_ready + 1)); terminal_not_ready=$((terminal_not_ready + 1)) ;;
      failed_checks) failed_checks=$((failed_checks + 1)); not_ready=$((not_ready + 1)); terminal_not_ready=$((terminal_not_ready + 1)) ;;
      error|codex_trigger_failed) not_ready=$((not_ready + 1)); terminal_not_ready=$((terminal_not_ready + 1)) ;;
      *) not_ready=$((not_ready + 1)) ;;
    esac

    printf '%-18s %-20s %s\n' "$(echo "$status" | tr '[:lower:]' '[:upper:]')" "${failed_kinds:--}" "$pr"
  done < <(read_prs)

  rm -rf "$tmpdir"
  printf '\nsummary: ready=%d not_ready=%d codex_feedback=%d failed_checks=%d errors=%d\n' \
    "$ready" "$not_ready" "$codex_feedback" "$failed_checks" "$errors"

  if (( errors > 0 )); then return 2; fi
  if (( codex_feedback > 0 )); then return 3; fi
  if (( failed_checks > 0 )); then return 4; fi
  if (( not_ready > 0 )); then return 1; fi
  return 0
}

START="$(date +%s)"
ATTEMPT=1
while true; do
  if [[ "$WATCH" == "true" ]]; then
    echo "--- batch attempt ${ATTEMPT} elapsed=$(( $(date +%s) - START ))s $(date -Is) ---"
  fi
  set +e
  run_once
  code=$?
  set -e

  if [[ "$WATCH" != "true" ]]; then
    exit "$code"
  fi
  if [[ "$code" == "0" || "$code" == "2" || "$code" == "3" || "$code" == "4" ]]; then
    exit "$code"
  fi
  ELAPSED=$(( $(date +%s) - START ))
  if (( ELAPSED >= TIMEOUT )); then
    echo "timed out after ${ELAPSED}s waiting for batch readiness" >&2
    exit 1
  fi
  ATTEMPT=$((ATTEMPT + 1))
  sleep "$INTERVAL"
done
