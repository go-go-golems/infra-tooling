#!/usr/bin/env bash
set -euo pipefail

# Trigger a Codex review by adding the standard review request comment.
# Usage: ./02-trigger-codex-review.sh https://github.com/OWNER/REPO/pull/NUMBER

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <pr-url-or-number>" >&2
  exit 2
fi

gh pr comment "$1" --body '@codex review'
