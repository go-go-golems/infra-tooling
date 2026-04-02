#!/usr/bin/env python3

from __future__ import annotations

import argparse
import sys
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTION_SRC = REPO_ROOT / "actions" / "open-gitops-pr" / "src"
sys.path.insert(0, str(ACTION_SRC))

from gitops_pr_action.open_gitops_pr import validate_targets  # noqa: E402


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Validate deploy/gitops-targets.json")
    parser.add_argument("config", nargs="?", default="deploy/gitops-targets.json")
    args = parser.parse_args(argv)
    config_path = Path(args.config).resolve()
    targets = validate_targets(config_path)
    print(f"{config_path}: OK ({len(targets)} target(s))")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
