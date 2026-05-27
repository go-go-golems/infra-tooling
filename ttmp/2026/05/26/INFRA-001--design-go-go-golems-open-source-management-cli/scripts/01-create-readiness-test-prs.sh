#!/usr/bin/env bash
set -euo pipefail

# Create intentionally varied PRs to exercise ggg readiness tooling:
# - a harmless control PR that should become ready after CI/Codex;
# - an intentionally failing-test PR that should classify as failed_checks;
# - an intentionally bad-code PR that should invite Codex feedback while CI may pass.
#
# The script starts each branch from origin/main and writes the resulting PR URLs
# to scripts/02-readiness-test-prs.yaml.

ROOT="${ROOT:-/home/manuel/code/wesen/go-go-golems/infra-tooling}"
REMOTE="${REMOTE:-origin}"
BASE="${BASE:-main}"
TICKET_DIR="$ROOT/ttmp/2026/05/26/INFRA-001--design-go-go-golems-open-source-management-cli"
PRS_FILE="$TICKET_DIR/scripts/02-readiness-test-prs.yaml"

cd "$ROOT"

git fetch "$REMOTE" "$BASE"
if [[ -n "$(git status --porcelain)" ]]; then
  echo "working tree is dirty; commit or stash before creating test PRs" >&2
  git status --short >&2
  exit 1
fi

mkdir -p "$TICKET_DIR/scripts"
cat > "$PRS_FILE" <<'YAML'
prs: []
YAML

append_pr() {
  local url="$1" kind="$2"
  python3 - "$PRS_FILE" "$url" "$kind" <<'PY'
from pathlib import Path
import sys
path = Path(sys.argv[1])
url = sys.argv[2]
kind = sys.argv[3]
text = path.read_text()
if text.strip() == "prs: []":
    text = "prs:\n"
text += f"  - url: {url}\n    kind: {kind}\n"
path.write_text(text)
PY
}

create_ready_control() {
  local branch="test/infra-001-ready-control"
  git checkout -B "$branch" "$REMOTE/$BASE"
  mkdir -p docs/go-go-golems/experiments
  cat > docs/go-go-golems/experiments/infra-001-ready-control.md <<'MD'
# INFRA-001 ready-control PR

This is a harmless documentation-only PR used to test `ggg pr ready` and
`ggg batch ready` on a PR that should pass CI and receive a benign Codex review.
MD
  git add docs/go-go-golems/experiments/infra-001-ready-control.md
  git commit -m "Test readiness control PR"
  git push -f "$REMOTE" "$branch"
  local url
  url=$(gh pr create --repo go-go-golems/infra-tooling --base "$BASE" --head "$branch" --title "INFRA-001 readiness control PR" --body "Harmless documentation-only PR for testing ggg readiness behavior.")
  append_pr "$url" "ready-control"
}

create_failing_tests() {
  local branch="test/infra-001-failing-tests"
  git checkout -B "$branch" "$REMOTE/$BASE"
  cat > tests/gitops/test_infra001_intentional_failure.py <<'PY'
import unittest


class Infra001IntentionalFailure(unittest.TestCase):
    def test_intentional_failure_for_readiness_tooling(self) -> None:
        self.assertEqual("expected", "intentionally-wrong")
PY
  git add tests/gitops/test_infra001_intentional_failure.py
  git commit -m "Test readiness failed checks PR"
  git push -f "$REMOTE" "$branch"
  local url
  url=$(gh pr create --repo go-go-golems/infra-tooling --base "$BASE" --head "$branch" --title "INFRA-001 intentionally failing checks PR" --body "This PR intentionally adds a failing unittest so ggg readiness can classify failed checks. Do not merge.")
  append_pr "$url" "failed-checks"
}

create_codex_feedback_bait() {
  local branch="test/infra-001-codex-feedback-bait"
  git checkout -B "$branch" "$REMOTE/$BASE"
  cat > scripts/go-go-golems/99-infra001-dangerous-example.py <<'PY'
#!/usr/bin/env python3
"""Intentionally unsafe example for readiness tooling tests. Do not use."""

import os
import subprocess
import sys


def delete_path_from_user_input() -> None:
    # Egregiously wrong on purpose: shell=True with unsanitized user input and rm -rf.
    target = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("TARGET", "/tmp/missing")
    subprocess.run(f"rm -rf {target}", shell=True, check=False)


if __name__ == "__main__":
    delete_path_from_user_input()
PY
  chmod +x scripts/go-go-golems/99-infra001-dangerous-example.py
  git add scripts/go-go-golems/99-infra001-dangerous-example.py
  git commit -m "Test readiness Codex feedback PR"
  git push -f "$REMOTE" "$branch"
  local url
  url=$(gh pr create --repo go-go-golems/infra-tooling --base "$BASE" --head "$branch" --title "INFRA-001 intentionally unsafe Codex feedback PR" --body "This PR intentionally adds unsafe code so Codex/readiness tooling can surface review feedback. Do not merge.")
  append_pr "$url" "codex-feedback-bait"
}

create_ready_control
create_failing_tests
create_codex_feedback_bait

git checkout "$BASE"
git pull --ff-only "$REMOTE" "$BASE"

echo "Wrote $PRS_FILE"
cat "$PRS_FILE"
