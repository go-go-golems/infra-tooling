#!/usr/bin/env bash
set -euo pipefail

# Trigger Codex review for multiple PRs listed one per line.
# Usage:
#   06-batch-trigger-codex-review.sh <prs-file>

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <prs-file>" >&2
  exit 2
fi

PRS_FILE="$1"
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
TRIGGER="$SCRIPT_DIR/02-trigger-codex-review.sh"

if [[ ! -f "$PRS_FILE" ]]; then
  echo "PR file not found: $PRS_FILE" >&2
  exit 2
fi

ok=0
failed=0

while IFS= read -r pr || [[ -n "$pr" ]]; do
  pr="${pr%%#*}"
  pr="$(printf '%s' "$pr" | xargs)"
  [[ -z "$pr" ]] && continue

  printf 'triggering Codex: %s\n' "$pr"
  if "$TRIGGER" "$pr"; then
    ok=$((ok + 1))
  else
    failed=$((failed + 1))
  fi
done < "$PRS_FILE"

printf 'summary: triggered=%d failed=%d\n' "$ok" "$failed"
if (( failed > 0 )); then
  exit 1
fi
