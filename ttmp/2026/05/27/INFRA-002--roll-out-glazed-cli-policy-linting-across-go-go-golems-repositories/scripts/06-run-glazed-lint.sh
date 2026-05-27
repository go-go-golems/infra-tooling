#!/usr/bin/env bash
set -euo pipefail

# Run `make glazed-lint` for each active workspace target and collect logs.
# This script does not stop on the first repository failure; it writes one log
# per repository and exits non-zero if any target fails.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGETS="${1:-${SCRIPT_DIR}/02-active-workspace-targets.txt}"
LOG_DIR="${SCRIPT_DIR}/../sources/glazed-lint-logs"
mkdir -p "$LOG_DIR"

failures=0
while IFS= read -r repo; do
  [[ -z "$repo" || "$repo" =~ ^# ]] && continue
  name="$(basename "$repo")"
  log="${LOG_DIR}/${name}.log"
  echo "=== ${name} ==="
  if (cd "$repo" && make glazed-lint) >"$log" 2>&1; then
    echo "ok ${name} log=${log}"
  else
    code=$?
    failures=$((failures + 1))
    echo "FAIL ${name} exit=${code} log=${log}"
    tail -80 "$log" || true
  fi
done < "$TARGETS"

if (( failures > 0 )); then
  echo "glazed-lint failures: ${failures}" >&2
  exit 1
fi
