#!/usr/bin/env python3

from __future__ import annotations

import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTION_SRC = REPO_ROOT / "actions" / "open-gitops-pr" / "src"
sys.path.insert(0, str(ACTION_SRC))

from gitops_pr_action.open_gitops_pr import cli  # noqa: E402


if __name__ == "__main__":
    raise SystemExit(cli())
