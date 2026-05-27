#!/usr/bin/env bash
set -euo pipefail

# Push each target repo's infra-002/glazed-lint branch and open a PR.
# Does not merge anything.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TARGETS="${1:-${SCRIPT_DIR}/02-active-workspace-targets.txt}"
BRANCH="${BRANCH:-infra-002/glazed-lint}"
OUT="${OUT:-${SCRIPT_DIR}/10-glazed-lint-prs.yaml}"
TITLE="Run Glazed CLI policy linting"
BODY_FILE="${SCRIPT_DIR}/10-pr-body.md"

cat > "$BODY_FILE" <<'BODY'
## Summary

- add or normalize `make glazed-lint`
- wire Glazed CLI policy linting into local lint targets and CI lint workflow where needed
- keep legacy allow paths narrow for existing command bridge/tool code

## Validation

- `make glazed-lint`

This PR is part of INFRA-002. Please do not merge until the rollout batch is reviewed.
BODY

printf 'prs:\n' > "$OUT"

while IFS= read -r repo; do
  [[ -z "$repo" || "$repo" =~ ^# ]] && continue
  name="$(basename "$repo")"
  echo "=== ${name} ==="
  git -C "$repo" fetch origin main --quiet
  current="$(git -C "$repo" branch --show-current)"
  if [[ "$current" != "$BRANCH" ]]; then
    echo "expected branch $BRANCH, got ${current:-<detached>}" >&2
    exit 2
  fi
  ahead="$(git -C "$repo" rev-list --count origin/main..HEAD)"
  if [[ "$ahead" != "1" ]]; then
    echo "expected exactly one commit ahead of origin/main, got $ahead" >&2
    exit 2
  fi
  if [[ -n "$(git -C "$repo" status --short --untracked-files=no)" ]]; then
    echo "tracked working tree is dirty in $repo" >&2
    git -C "$repo" status --short >&2
    exit 2
  fi
  git -C "$repo" push -u origin "$BRANCH"
  url="$(cd "$repo" && gh pr view "$BRANCH" --json url --jq .url 2>/dev/null || true)"
  if [[ -z "$url" ]]; then
    url="$(cd "$repo" && gh pr create --base main --head "$BRANCH" --title "$TITLE" --body-file "$BODY_FILE")"
  fi
  echo "PR: $url"
  printf '  - %s\n' "$url" >> "$OUT"
done < "$TARGETS"

echo "wrote $OUT"
