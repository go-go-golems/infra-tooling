#!/usr/bin/env bash
set -euo pipefail

# Add synthetic commit statuses to INFRA-001 readiness test PRs. The repo's
# current workflows may not report checks for these test branches, so this script
# creates StatusContext entries that exercise ggg readiness classification.
#
# Usage:
#   03-set-readiness-test-statuses.sh

REPO="${REPO:-go-go-golems/infra-tooling}"

declare -A STATES=(
  [5]=success
  [6]=failure
  [7]=success
)

declare -A DESCRIPTIONS=(
  [5]="synthetic readiness control status"
  [6]="synthetic failing status for readiness testing"
  [7]="synthetic success status; Codex feedback should still block"
)

for pr in 5 6 7; do
  sha="$(gh pr view "$pr" --repo "$REPO" --json headRefOid --jq .headRefOid)"
  state="${STATES[$pr]}"
  description="${DESCRIPTIONS[$pr]}"
  echo "setting $REPO PR #$pr $sha -> $state"
  gh api "repos/$REPO/statuses/$sha" \
    -f state="$state" \
    -f context="infra-001/synthetic-readiness" \
    -f description="$description" \
    -f target_url="https://github.com/$REPO/pull/$pr" >/dev/null
done
