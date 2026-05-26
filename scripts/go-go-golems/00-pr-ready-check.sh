#!/usr/bin/env bash
set -euo pipefail

# Bash entry point for PR readiness checks. Delegates the GitHub GraphQL
# parsing to the adjacent Python implementation while keeping the operator
# interface shell-friendly.
#
# Usage:
#   ./00-pr-ready-check.sh https://github.com/OWNER/REPO/pull/NUMBER
#   ./00-pr-ready-check.sh OWNER/REPO#NUMBER --json

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
exec python3 "$SCRIPT_DIR/01-pr-ready-check.py" "$@"
