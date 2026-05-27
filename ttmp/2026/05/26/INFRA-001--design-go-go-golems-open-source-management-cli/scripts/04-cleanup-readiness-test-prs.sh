#!/usr/bin/env bash
set -euo pipefail

# Close the live INFRA-001 readiness test PRs and delete their remote branches.
# These PRs were intentionally opened to validate ggg readiness behavior and are
# not meant to be merged.
#
# Usage:
#   04-cleanup-readiness-test-prs.sh

REPO="${REPO:-go-go-golems/infra-tooling}"
PRS=(5 6 7)

for pr in "${PRS[@]}"; do
  state="$(gh pr view "$pr" --repo "$REPO" --json state --jq .state 2>/dev/null || true)"
  if [[ -z "$state" ]]; then
    echo "PR #$pr not found; skipping"
    continue
  fi
  if [[ "$state" != "OPEN" ]]; then
    echo "PR #$pr is $state; skipping close"
    continue
  fi
  echo "closing $REPO PR #$pr"
  gh pr close "$pr" \
    --repo "$REPO" \
    --delete-branch \
    --comment "Closing INFRA-001 live readiness-tooling test PR after validation. This PR was intentionally created as a disposable fixture and is not meant to merge."
done
