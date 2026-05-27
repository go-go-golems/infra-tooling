#!/usr/bin/env bash
set -euo pipefail

# Create/update an INFRA-002 branch in each active workspace repo and commit only
# the Glazed lint rollout files. This intentionally does not stage incidental
# build artifacts such as .bin/.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGETS="${1:-${SCRIPT_DIR}/02-active-workspace-targets.txt}"
BRANCH="${BRANCH:-infra-002/glazed-lint}"
MESSAGE="${MESSAGE:-Run Glazed CLI policy linting}"

while IFS= read -r repo; do
  [[ -z "$repo" || "$repo" =~ ^# ]] && continue
  name="$(basename "$repo")"
  echo "=== ${name} ==="
  git -C "$repo" checkout -B "$BRANCH"
  paths=(Makefile .github/workflows/lint.yml)
  existing=()
  for p in "${paths[@]}"; do
    [[ -e "$repo/$p" ]] && existing+=("$p")
  done
  if [[ ${#existing[@]} -eq 0 ]]; then
    echo "no rollout files found; skipping"
    continue
  fi
  git -C "$repo" add "${existing[@]}"
  if git -C "$repo" diff --cached --quiet; then
    echo "nothing staged; skipping commit"
    continue
  fi
  git -C "$repo" commit -m "$MESSAGE"
  git -C "$repo" status --short --branch
  git -C "$repo" rev-parse HEAD
done < "$TARGETS"
