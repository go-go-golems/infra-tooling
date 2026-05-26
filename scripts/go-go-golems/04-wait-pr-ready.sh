#!/usr/bin/env bash
set -euo pipefail

# Poll a PR until 00-pr-ready-check.sh reports ready, until timeout, or until
# Codex has posted substantive review comments that require operator action.
# Usage:
#   ./04-wait-pr-ready.sh <pr-url-or-owner/repo#number> [interval-seconds] [timeout-seconds]

if [[ $# -lt 1 || $# -gt 3 ]]; then
  echo "usage: $0 <pr-url-or-owner/repo#number> [interval-seconds] [timeout-seconds]" >&2
  exit 2
fi

PR="$1"
INTERVAL="${2:-30}"
TIMEOUT="${3:-1800}"
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
CHECK="$SCRIPT_DIR/00-pr-ready-check.sh"
START="$(date +%s)"
ATTEMPT=1

json_field() {
  python3 -c 'import json,sys; data=json.load(open(sys.argv[1])); print(data.get(sys.argv[2], ""))' "$1" "$2"
}

while true; do
  NOW="$(date +%s)"
  ELAPSED=$((NOW - START))
  echo "--- attempt ${ATTEMPT} elapsed=${ELAPSED}s $(date -Is) ---"

  CHECK_JSON="$(mktemp)"
  if "$CHECK" "$PR" --json >"$CHECK_JSON"; then
    "$CHECK" "$PR"
    rm -f "$CHECK_JSON"
    echo "PR ready: $PR"
    exit 0
  fi

  "$CHECK" "$PR" || true
  STATE="$(json_field "$CHECK_JSON" state)"
  rm -f "$CHECK_JSON"

  if [[ "$STATE" == "codex_feedback" ]]; then
    echo "Codex posted substantive review comments; stopping wait for operator action: $PR" >&2
    exit 3
  fi
  if [[ "$STATE" == "failed_checks" ]]; then
    echo "Status checks failed; stopping wait for operator action: $PR" >&2
    exit 4
  fi

  NOW="$(date +%s)"
  ELAPSED=$((NOW - START))
  if (( ELAPSED >= TIMEOUT )); then
    echo "timed out after ${ELAPSED}s waiting for PR readiness: $PR" >&2
    exit 1
  fi
  ATTEMPT=$((ATTEMPT + 1))
  sleep "$INTERVAL"
done
