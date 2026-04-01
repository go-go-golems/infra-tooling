#!/usr/bin/env python3

from __future__ import annotations

import argparse
import difflib
import json
import subprocess
import sys
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Dry-run or validate a federation-manifest GitOps target update against a local GitOps checkout.",
    )
    parser.add_argument("--config", required=True)
    parser.add_argument("--target")
    parser.add_argument("--manifest-url", required=True)
    parser.add_argument("--gitops-repo-dir", required=True)
    return parser.parse_args()


def load_targets(config_path: Path) -> list[dict]:
    payload = json.loads(config_path.read_text(encoding="utf-8"))
    targets = payload.get("targets")
    if not isinstance(targets, list) or not targets:
        raise ValueError(f"no targets defined in {config_path}")
    required = {
        "name",
        "kind",
        "remote_id",
        "gitops_repo",
        "gitops_branch",
        "target_file",
    }
    for target in targets:
        missing = sorted(required - set(target))
        if missing:
            raise ValueError(f"target {target!r} is missing keys: {', '.join(missing)}")
        if target["kind"] != "federation-manifest":
            raise ValueError(f"unsupported target kind: {target['kind']!r}")
    return targets


def select_target(targets: list[dict], name: str | None) -> dict:
    if name is None:
        if len(targets) != 1:
            raise ValueError("select a target with --target when multiple targets are defined")
        return targets[0]
    for target in targets:
        if target["name"] == name:
            return target
    raise ValueError(f"target {name!r} not found")


def run(cmd: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, check=True, text=True, capture_output=True)


def render_diff(path: Path, before: str, after: str) -> str:
    return "\n".join(
        difflib.unified_diff(
            before.splitlines(),
            after.splitlines(),
            fromfile=f"{path} (before)",
            tofile=f"{path} (after)",
            lineterm="",
        )
    )


def main() -> int:
    args = parse_args()
    config_path = Path(args.config).resolve()
    gitops_repo_dir = Path(args.gitops_repo_dir).resolve()
    target = select_target(load_targets(config_path), args.target)
    target_file = gitops_repo_dir / target["target_file"]

    before = target_file.read_text(encoding="utf-8")
    patch_script = Path(__file__).resolve().parent / "patch_federation_registry_target.py"
    cmd = [
        sys.executable,
        str(patch_script),
        "--target-file",
        str(target_file),
        "--remote-id",
        target["remote_id"],
        "--manifest-url",
        args.manifest_url,
        "--enabled",
        "true",
    ]
    if target.get("config_key"):
        cmd.extend(["--config-key", target["config_key"]])
    run(cmd)

    after = target_file.read_text(encoding="utf-8")
    target_file.write_text(before, encoding="utf-8")
    diff = render_diff(target_file, before, after)
    print(diff or f"No change needed for {target['name']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
