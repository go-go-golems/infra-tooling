#!/usr/bin/env python3

from __future__ import annotations

import argparse
import difflib
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path
from urllib.parse import urlparse


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Open or update GitOps PRs for federation manifest targets."
    )
    parser.add_argument("--config", required=True)
    parser.add_argument("--target")
    parser.add_argument("--manifest-url", required=True)
    parser.add_argument(
        "--gitops-repo-dir",
        help="Use an existing local GitOps checkout for dry-run validation.",
    )
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--push", action="store_true")
    parser.add_argument("--open-pr", action="store_true")
    return parser.parse_args()


def run(cmd: list[str], cwd: Path | None = None, capture_output: bool = False) -> subprocess.CompletedProcess[str]:
    kwargs = {
        "cwd": cwd,
        "check": True,
        "text": True,
    }
    if capture_output:
        kwargs["stdout"] = subprocess.PIPE
        kwargs["stderr"] = subprocess.PIPE
    return subprocess.run(cmd, **kwargs)


def load_targets(config_path: Path) -> list[dict]:
    payload = json.loads(config_path.read_text(encoding="utf-8"))
    targets = payload.get("targets", [])
    if not isinstance(targets, list) or not targets:
        raise ValueError(f"no targets defined in {config_path}")
    required = {
        "name",
        "kind",
        "remote_id",
        "gitops_repo",
        "gitops_branch",
        "target_file",
        "config_key",
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


def sanitize_ref_fragment(value: str) -> str:
    return re.sub(r"[^a-zA-Z0-9._/-]+", "-", value).strip("-")


def manifest_version_fragment(manifest_url: str) -> str:
    path_parts = [part for part in urlparse(manifest_url).path.split("/") if part]
    if len(path_parts) >= 2 and path_parts[-1] == "mf-manifest.json":
        return sanitize_ref_fragment(path_parts[-2])
    return sanitize_ref_fragment(Path(manifest_url).name.replace(".", "-"))


def clone_repo(target: dict, token: str) -> Path:
    temp_dir = Path(tempfile.mkdtemp(prefix=f"gitops-{target['name']}-"))
    remote_url = f"https://x-access-token:{token}@github.com/{target['gitops_repo']}.git"
    run(
        [
            "git",
            "clone",
            "--depth",
            "1",
            "--branch",
            target["gitops_branch"],
            remote_url,
            str(temp_dir),
        ]
    )
    return temp_dir


def ensure_git_identity(repo_dir: Path) -> None:
    author_name = os.environ.get("GITOPS_PR_GIT_AUTHOR_NAME", "github-actions[bot]")
    author_email = os.environ.get(
        "GITOPS_PR_GIT_AUTHOR_EMAIL",
        "41898282+github-actions[bot]@users.noreply.github.com",
    )
    run(["git", "config", "user.name", author_name], cwd=repo_dir)
    run(["git", "config", "user.email", author_email], cwd=repo_dir)


def patch_target_file(target_file: Path, remote_id: str, config_key: str, manifest_url: str) -> tuple[str, str]:
    before = target_file.read_text(encoding="utf-8")
    patch_script = Path(__file__).resolve().parent / "patch_federation_registry_target.py"
    run(
        [
            sys.executable,
            str(patch_script),
            "--target-file",
            str(target_file),
            "--remote-id",
            remote_id,
            "--config-key",
            config_key,
            "--manifest-url",
            manifest_url,
            "--enabled",
            "true",
        ]
    )
    after = target_file.read_text(encoding="utf-8")
    return before, after


def get_existing_pr_number(repo_dir: Path, target: dict, branch_name: str) -> str:
    return run(
        [
            "gh",
            "pr",
            "list",
            "--repo",
            target["gitops_repo"],
            "--head",
            branch_name,
            "--state",
            "open",
            "--json",
            "number",
            "--jq",
            ".[0].number // empty",
        ],
        cwd=repo_dir,
        capture_output=True,
    ).stdout.strip()


def open_or_update_pr(repo_dir: Path, target: dict, branch_name: str, title: str, body: str) -> None:
    existing = get_existing_pr_number(repo_dir, target, branch_name)
    if existing:
        print(f"PR already exists for {target['name']}: #{existing}")
        return
    run(
        [
            "gh",
            "pr",
            "create",
            "--repo",
            target["gitops_repo"],
            "--base",
            target["gitops_branch"],
            "--head",
            branch_name,
            "--title",
            title,
            "--body",
            body,
        ],
        cwd=repo_dir,
    )


def build_pr_body(target: dict, manifest_url: str) -> str:
    source_repo = os.environ.get("GITHUB_REPOSITORY", "")
    source_sha = os.environ.get("GITHUB_SHA", "")
    workflow_run = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
    run_id = os.environ.get("GITHUB_RUN_ID", "")

    lines = [
        f"Automated federation manifest bump for `{target['name']}`.",
        "",
        f"- Remote id: `{target['remote_id']}`",
        f"- Manifest URL: `{manifest_url}`",
        f"- Target file: `{target['target_file']}`",
        f"- Config key: `{target['config_key']}`",
    ]
    if source_repo and source_sha:
        lines.append(f"- Source commit: `{source_repo}@{source_sha}`")
    if source_repo and run_id:
        lines.append(f"- Workflow run: {workflow_run}/{source_repo}/actions/runs/{run_id}")
    lines.extend(
        [
            "",
            "Rollback:",
            f"- revert this PR merge, or open a new PR that resets `{target['remote_id']}` to the previous immutable manifest URL",
        ]
    )
    return "\n".join(lines)


def main() -> int:
    args = parse_args()
    config_path = Path(args.config).resolve()
    target = select_target(load_targets(config_path), args.target)

    if args.open_pr and not args.push:
        raise ValueError("--open-pr requires --push")

    token = os.environ.get("GH_TOKEN", "")
    if (args.push or args.open_pr) and not token and not args.gitops_repo_dir:
        raise ValueError("GH_TOKEN is required when --push or --open-pr is used")

    if args.gitops_repo_dir:
        repo_dir = Path(args.gitops_repo_dir).resolve()
        created_temp_dir = False
    else:
        repo_dir = clone_repo(target, token)
        created_temp_dir = True

    try:
        ensure_git_identity(repo_dir)
        target_file = repo_dir / target["target_file"]
        before, after = patch_target_file(
            target_file=target_file,
            remote_id=target["remote_id"],
            config_key=target["config_key"],
            manifest_url=args.manifest_url,
        )
        diff = render_diff(target_file, before, after)
        print(diff or f"No change needed for {target['name']}")
        changed = before != after
        if not changed:
            return 0

        if args.dry_run and not args.push:
            target_file.write_text(before, encoding="utf-8")
            return 0

        version_fragment = manifest_version_fragment(args.manifest_url)
        branch_name = f"automation/federation-{target['remote_id']}-{target['name']}-{version_fragment}"
        title = f"Deploy {target['remote_id']} for {target['name']} using {version_fragment}"
        body = build_pr_body(target, args.manifest_url)

        if args.open_pr:
            existing_pr = get_existing_pr_number(repo_dir, target, branch_name)
            if existing_pr:
                print(f"PR already exists for {target['name']}: #{existing_pr}")
                return 0

        run(["git", "checkout", "-b", branch_name], cwd=repo_dir)
        run(["git", "add", target["target_file"]], cwd=repo_dir)
        run(["git", "commit", "-m", title], cwd=repo_dir)

        if args.push:
            run(["git", "push", "origin", branch_name], cwd=repo_dir)
        if args.open_pr:
            open_or_update_pr(repo_dir, target, branch_name, title, body)
        return 0
    finally:
        if created_temp_dir:
            shutil.rmtree(repo_dir, ignore_errors=True)


if __name__ == "__main__":
    raise SystemExit(main())
