#!/usr/bin/env bash
set -euo pipefail

# Check readiness for multiple PRs without blocking on a single PR.
# Usage:
#   05-batch-pr-ready.sh <prs-file> [--trigger-missing-codex]
#
# The PR file contains one PR URL or owner/repo#number per line. Blank lines and
# lines beginning with # are ignored.

if [[ $# -lt 1 || $# -gt 2 ]]; then
  echo "usage: $0 <prs-file> [--trigger-missing-codex]" >&2
  exit 2
fi

PRS_FILE="$1"
TRIGGER_MISSING="false"
if [[ $# -eq 2 ]]; then
  case "$2" in
    --trigger-missing-codex) TRIGGER_MISSING="true" ;;
    *) echo "unknown option: $2" >&2; exit 2 ;;
  esac
fi

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/00-pr-ready-check.sh"
TRIGGER="$SCRIPT_DIR/02-trigger-codex-review.sh"

if [[ ! -f "$PRS_FILE" ]]; then
  echo "PR file not found: $PRS_FILE" >&2
  exit 2
fi

status_for_output() {
  local output="$1"
  if grep -q '^READY: yes' "$output"; then
    echo READY
  elif grep -q 'latest Codex-authored body contains substantive comments' "$output"; then
    echo CODEX_FEEDBACK
  elif grep -q 'pending checks:' "$output"; then
    echo WAITING_CHECKS
  elif grep -q 'failing/non-success checks:' "$output"; then
    echo FAILED_CHECKS
  elif grep -q 'no Codex-authored review/comment signal found' "$output"; then
    echo NO_CODEX
  elif grep -q 'eyes reaction' "$output"; then
    echo WAITING_CODEX
  elif grep -q 'no thumbs-up reaction or satisfied thumbs-up body' "$output"; then
    echo WAITING_CODEX
  else
    echo NOT_READY
  fi
}

ready=0
not_ready=0
codex_feedback=0
failed_checks=0
errors=0

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

printf '%-16s %s\n' STATUS PR
printf '%-16s %s\n' '------' '--'

while IFS= read -r pr || [[ -n "$pr" ]]; do
  pr="${pr%%#*}"
  pr="$(printf '%s' "$pr" | xargs)"
  [[ -z "$pr" ]] && continue

  out="$tmpdir/out.$RANDOM.txt"
  if "$CHECK" "$pr" >"$out" 2>&1; then
    status=READY
    ready=$((ready + 1))
  else
    status="$(status_for_output "$out")"
    not_ready=$((not_ready + 1))
    case "$status" in
      CODEX_FEEDBACK) codex_feedback=$((codex_feedback + 1)) ;;
      FAILED_CHECKS) failed_checks=$((failed_checks + 1)) ;;
      NO_CODEX)
        if [[ "$TRIGGER_MISSING" == "true" ]]; then
          if "$TRIGGER" "$pr" >/dev/null 2>&1; then
            status=CODEX_TRIGGERED
          else
            status=CODEX_TRIGGER_FAILED
            errors=$((errors + 1))
          fi
        fi
        ;;
    esac
  fi
  printf '%-16s %s\n' "$status" "$pr"
done < "$PRS_FILE"

printf '\nsummary: ready=%d not_ready=%d codex_feedback=%d failed_checks=%d errors=%d\n' \
  "$ready" "$not_ready" "$codex_feedback" "$failed_checks" "$errors"

if (( errors > 0 )); then
  exit 2
fi
if (( codex_feedback > 0 )); then
  exit 3
fi
if (( failed_checks > 0 )); then
  exit 4
fi
if (( not_ready > 0 )); then
  exit 1
fi
exit 0
