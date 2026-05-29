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
from dataclasses import dataclass
from pathlib import Path
from typing import Any


REQUIRED_TARGET_KEYS = {
    "name",
    "gitops_repo",
    "gitops_branch",
    "manifest_path",
}

OPTIONAL_TARGET_KEYS = {
    "container_name",
    "patch_strategy",
    "images",
}

REQUIRED_TARGET_IMAGE_KEYS = {
    "container_name",
}

OPTIONAL_TARGET_IMAGE_KEYS = {
    "image_name",
    "image",
    "patch_strategy",
}

SUPPORTED_PATCH_STRATEGIES = {
    "container-image",
    "static-publisher-job",
}

SHA_RE = re.compile(r"^(?:sha-)?[0-9a-fA-F]{7,40}$")
RELEASE_RE = re.compile(r"sha-[0-9a-fA-F]{7,40}")


@dataclass
class TargetResult:
    target_name: str
    changed: bool
    branch_name: str = ""
    pr_number: str = ""
    manifest_path: str = ""


def run(
    cmd: list[str],
    cwd: Path | None = None,
    capture_output: bool = False,
) -> subprocess.CompletedProcess[str]:
    kwargs: dict[str, Any] = {
        "cwd": cwd,
        "check": True,
        "text": True,
    }
    if capture_output:
        kwargs["stdout"] = subprocess.PIPE
        kwargs["stderr"] = subprocess.PIPE
    return subprocess.run(cmd, **kwargs)


def normalize_target_image(raw_image: Any, raw_target: dict[str, Any]) -> dict[str, str]:
    if not isinstance(raw_image, dict):
        raise ValueError(f"target {raw_target!r} has non-object image entry {raw_image!r}")

    missing = sorted(REQUIRED_TARGET_IMAGE_KEYS - set(raw_image))
    if missing:
        raise ValueError(
            f"target {raw_target!r} image entry {raw_image!r} is missing keys: {', '.join(missing)}"
        )

    image = {key: str(raw_image[key]).strip() for key in REQUIRED_TARGET_IMAGE_KEYS}
    for key in OPTIONAL_TARGET_IMAGE_KEYS:
        if key in raw_image:
            image[key] = str(raw_image[key]).strip()

    empty_values = sorted(key for key, value in image.items() if not value)
    if empty_values:
        raise ValueError(
            f"target {raw_target!r} image entry {raw_image!r} has empty values for keys: "
            f"{', '.join(empty_values)}"
        )

    if "image" in image and "image_name" in image:
        raise ValueError(
            f"target {raw_target!r} image entry {raw_image!r} must use either image or image_name, not both"
        )

    patch_strategy = image.get("patch_strategy", raw_target.get("patch_strategy", "container-image"))
    if patch_strategy not in SUPPORTED_PATCH_STRATEGIES:
        supported = ", ".join(sorted(SUPPORTED_PATCH_STRATEGIES))
        raise ValueError(
            f"target {raw_target!r} image entry {raw_image!r} has unsupported patch_strategy "
            f"{patch_strategy!r}; supported values: {supported}"
        )
    image["patch_strategy"] = patch_strategy
    return image


def load_targets(config_path: Path) -> list[dict[str, Any]]:
    payload = json.loads(config_path.read_text(encoding="utf-8"))
    targets = payload.get("targets", [])
    if not isinstance(targets, list) or not targets:
        raise ValueError(f"no targets defined in {config_path}")

    normalized: list[dict[str, Any]] = []
    seen_names: set[str] = set()
    for raw_target in targets:
        if not isinstance(raw_target, dict):
            raise ValueError(f"target {raw_target!r} is not an object")
        missing = sorted(REQUIRED_TARGET_KEYS - set(raw_target))
        if missing:
            raise ValueError(f"target {raw_target!r} is missing keys: {', '.join(missing)}")

        target: dict[str, Any] = {key: str(raw_target[key]).strip() for key in REQUIRED_TARGET_KEYS}
        for key in OPTIONAL_TARGET_KEYS:
            if key in raw_target and key != "images":
                target[key] = str(raw_target[key]).strip()

        empty_values = sorted(key for key, value in target.items() if isinstance(value, str) and not value)
        if empty_values:
            raise ValueError(
                f"target {raw_target!r} has empty values for keys: {', '.join(empty_values)}"
            )

        has_legacy_container = "container_name" in target
        has_images = "images" in raw_target
        if has_legacy_container and has_images:
            raise ValueError(f"target {raw_target!r} must use either container_name or images, not both")
        if not has_legacy_container and not has_images:
            raise ValueError(f"target {raw_target!r} must define container_name or images")

        patch_strategy = target.get("patch_strategy", "container-image")
        if patch_strategy not in SUPPORTED_PATCH_STRATEGIES:
            supported = ", ".join(sorted(SUPPORTED_PATCH_STRATEGIES))
            raise ValueError(
                f"target {raw_target!r} has unsupported patch_strategy {patch_strategy!r}; "
                f"supported values: {supported}"
            )
        target["patch_strategy"] = patch_strategy

        if has_images:
            raw_images = raw_target["images"]
            if not isinstance(raw_images, list) or not raw_images:
                raise ValueError(f"target {raw_target!r} images must be a non-empty list")
            target_images = [normalize_target_image(raw_image, raw_target) for raw_image in raw_images]
            seen_containers: set[str] = set()
            for image_entry in target_images:
                container_name = image_entry["container_name"]
                if container_name in seen_containers:
                    raise ValueError(
                        f"target {raw_target!r} has duplicate image container_name {container_name!r}"
                    )
                seen_containers.add(container_name)
            target["images"] = target_images

        if target["name"] in seen_names:
            raise ValueError(f"duplicate target name {target['name']!r} in {config_path}")
        seen_names.add(target["name"])
        normalized.append(target)

    return normalized


def validate_targets(config_path: Path) -> list[dict[str, Any]]:
    return load_targets(config_path)


def select_targets(
    targets: list[dict[str, Any]],
    target_name: str | None,
    all_targets: bool,
) -> list[dict[str, Any]]:
    if target_name and all_targets:
        raise ValueError("use either --target or --all-targets, not both")
    if all_targets:
        return targets
    if target_name:
        for target in targets:
            if target["name"] == target_name:
                return [target]
        raise ValueError(f"target {target_name!r} not found")
    raise ValueError("select a target with --target or pass --all-targets")


def sanitize_ref_fragment(value: str) -> str:
    return re.sub(r"[^a-zA-Z0-9._/-]+", "-", value).strip("-")


def image_tag_fragment(image: str) -> str:
    if ":" in image:
        return sanitize_ref_fragment(image.split(":", 1)[1])
    return sanitize_ref_fragment(image.rsplit("/", 1)[-1])


def image_repo_fragment(image: str) -> str:
    repo = image.rsplit("/", 1)[-1]
    return sanitize_ref_fragment(repo.split(":", 1)[0])


def build_branch_name(image: str, target_name: str) -> str:
    app_name = image_repo_fragment(image)
    tag_fragment = image_tag_fragment(image)
    return f"automation/{app_name}-{target_name}-{tag_fragment}"


def build_commit_title(target_name: str, image: str) -> str:
    return f"Deploy {target_name} using {image}"


def replace_manifest_image_text(
    manifest_path: Path,
    original: str,
    container_name: str,
    image: str,
) -> tuple[bool, str]:
    lines = original.splitlines(keepends=True)

    containers_indent = None
    current_container = None
    target_container_indent = None
    replaced = False

    for idx, line in enumerate(lines):
        stripped = line.strip()
        if not stripped or stripped.startswith("#"):
            continue

        indent = len(line) - len(line.lstrip(" "))

        if stripped == "containers:":
            containers_indent = indent
            current_container = None
            target_container_indent = None
            continue

        if containers_indent is not None and indent <= containers_indent:
            containers_indent = None
            current_container = None
            target_container_indent = None

        if containers_indent is None:
            continue

        if stripped.startswith("- name:") and indent == containers_indent + 2:
            current_container = stripped.split(":", 1)[1].strip()
            target_container_indent = indent
            continue

        if current_container != container_name or target_container_indent is None:
            continue

        if indent == target_container_indent + 2 and stripped.startswith("image:"):
            prefix = line[: len(line) - len(line.lstrip(" "))]
            suffix = "\n" if line.endswith("\n") else ""
            existing = stripped.split(":", 1)[1].strip()
            if existing == image:
                return False, original
            lines[idx] = f"{prefix}image: {image}{suffix}"
            replaced = True
            break

    if not replaced:
        raise ValueError(
            f"could not find image field for container {container_name!r} in {manifest_path}"
        )

    return True, "".join(lines)


def patch_manifest_image(
    manifest_path: Path,
    container_name: str,
    image: str,
) -> tuple[bool, str, str]:
    original = manifest_path.read_text(encoding="utf-8")
    image_changed, updated = replace_manifest_image_text(
        manifest_path=manifest_path,
        original=original,
        container_name=container_name,
        image=image,
    )
    if not image_changed:
        return False, original, original
    manifest_path.write_text(updated, encoding="utf-8")
    return True, original, updated


def normalize_release(value: str) -> str:
    if not SHA_RE.match(value):
        raise ValueError(
            f"invalid static publisher release {value!r}; expected a git SHA or sha-<git-sha> "
            "with 7-40 hexadecimal characters"
        )
    suffix = value[4:] if value.lower().startswith("sha-") else value
    return f"sha-{suffix.lower()}"


def release_from_image(image: str) -> str:
    if ":" not in image:
        raise ValueError(
            f"static-publisher-job strategy requires an image tag in {image!r}; "
            "expected an immutable sha-<git-sha> tag"
        )
    return normalize_release(image.rsplit(":", 1)[1])


def patch_static_publisher_job(
    manifest_path: Path,
    container_name: str,
    image: str,
) -> tuple[bool, str, str]:
    original = manifest_path.read_text(encoding="utf-8")
    new_release = release_from_image(image)
    _, with_image = replace_manifest_image_text(
        manifest_path=manifest_path,
        original=original,
        container_name=container_name,
        image=image,
    )
    releases = sorted(set(RELEASE_RE.findall(with_image)))
    if not releases:
        raise ValueError(f"no sha-* release token found in static publisher manifest {manifest_path}")
    updated = RELEASE_RE.sub(new_release, with_image)
    if updated == original:
        return False, original, original
    manifest_path.write_text(updated, encoding="utf-8")
    return True, original, updated


def image_with_tag(image_name: str, source_image: str) -> str:
    if ":" not in source_image:
        raise ValueError(
            f"multi-image targets using image_name require --image to include a tag; got {source_image!r}"
        )
    return f"{image_name}:{source_image.rsplit(':', 1)[1]}"


def resolve_target_image(image_entry: dict[str, str], source_image: str) -> str:
    if "image" in image_entry:
        return image_entry["image"]
    if "image_name" in image_entry:
        return image_with_tag(image_entry["image_name"], source_image)
    return source_image


def patch_one_image(
    manifest_path: Path,
    original: str,
    container_name: str,
    image: str,
    strategy: str,
) -> tuple[bool, str]:
    if strategy == "container-image":
        return replace_manifest_image_text(
            manifest_path=manifest_path,
            original=original,
            container_name=container_name,
            image=image,
        )
    if strategy == "static-publisher-job":
        new_release = release_from_image(image)
        _, with_image = replace_manifest_image_text(
            manifest_path=manifest_path,
            original=original,
            container_name=container_name,
            image=image,
        )
        releases = sorted(set(RELEASE_RE.findall(with_image)))
        if not releases:
            raise ValueError(f"no sha-* release token found in static publisher manifest {manifest_path}")
        updated = RELEASE_RE.sub(new_release, with_image)
        return updated != original, updated
    raise ValueError(f"unsupported patch_strategy {strategy!r}")


def patch_target_manifest(
    manifest_path: Path,
    target: dict[str, Any],
    image: str,
) -> tuple[bool, str, str]:
    original = manifest_path.read_text(encoding="utf-8")
    updated = original
    changed_any = False

    if "images" in target:
        for image_entry in target["images"]:
            resolved_image = resolve_target_image(image_entry, image)
            changed, updated = patch_one_image(
                manifest_path=manifest_path,
                original=updated,
                container_name=image_entry["container_name"],
                image=resolved_image,
                strategy=image_entry.get("patch_strategy", target.get("patch_strategy", "container-image")),
            )
            changed_any = changed_any or changed
    else:
        changed_any, updated = patch_one_image(
            manifest_path=manifest_path,
            original=updated,
            container_name=target["container_name"],
            image=image,
            strategy=target.get("patch_strategy", "container-image"),
        )

    if not changed_any:
        return False, original, original
    manifest_path.write_text(updated, encoding="utf-8")
    return True, original, updated


def ensure_git_identity(repo_dir: Path) -> None:
    author_name = os.environ.get("GITOPS_PR_GIT_AUTHOR_NAME", "github-actions[bot]")
    author_email = os.environ.get(
        "GITOPS_PR_GIT_AUTHOR_EMAIL",
        "41898282+github-actions[bot]@users.noreply.github.com",
    )
    run(["git", "config", "user.name", author_name], cwd=repo_dir)
    run(["git", "config", "user.email", author_email], cwd=repo_dir)


def get_existing_pr_number(repo_dir: Path, target: dict[str, Any], branch_name: str) -> str:
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


def open_or_update_pr(
    repo_dir: Path,
    target: dict[str, Any],
    branch_name: str,
    title: str,
    body: str,
) -> str:
    existing = get_existing_pr_number(repo_dir, target, branch_name)
    if existing:
        print(f"PR already exists for {target['name']}: #{existing}")
        return existing

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
    return get_existing_pr_number(repo_dir, target, branch_name)


def clone_repo(target: dict[str, Any], token: str) -> Path:
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


def render_diff(path: Path, original: str, updated: str) -> str:
    diff = difflib.unified_diff(
        original.splitlines(),
        updated.splitlines(),
        fromfile=f"{path} (before)",
        tofile=f"{path} (after)",
        lineterm="",
    )
    return "\n".join(diff)


def target_image_lines(target: dict[str, Any], image: str) -> list[str]:
    if "images" not in target:
        return [f"- Image: `{image}`"]
    lines = ["- Images:"]
    for image_entry in target["images"]:
        resolved_image = resolve_target_image(image_entry, image)
        lines.append(f"  - `{image_entry['container_name']}`: `{resolved_image}`")
    return lines


def rollback_line(target: dict[str, Any]) -> str:
    if "images" in target:
        containers = ", ".join(f"`{entry['container_name']}`" for entry in target["images"])
        return f"- revert this PR merge, or open a new PR that resets {containers} to the previous immutable tags"
    return f"- revert this PR merge, or open a new PR that resets `{target['container_name']}` to the previous immutable tag"


def build_pr_body(target: dict[str, Any], image: str) -> str:
    source_repo = os.environ.get("GITHUB_REPOSITORY", "")
    source_sha = os.environ.get("GITHUB_SHA", "")
    workflow_run = os.environ.get("GITHUB_SERVER_URL", "https://github.com")
    run_id = os.environ.get("GITHUB_RUN_ID", "")

    lines = [
        f"Automated image bump for `{target['name']}`.",
        "",
        *target_image_lines(target, image),
        f"- Target manifest: `{target['manifest_path']}`",
    ]
    if source_repo and source_sha:
        lines.append(f"- Source commit: `{source_repo}@{source_sha}`")
    if source_repo and run_id:
        lines.append(f"- Workflow run: {workflow_run}/{source_repo}/actions/runs/{run_id}")
    lines.extend(
        [
            "",
            "Rollback:",
            rollback_line(target),
        ]
    )
    return "\n".join(lines)


def append_github_outputs(
    output_path: str | None,
    results: list[TargetResult],
) -> None:
    if not output_path:
        return

    changed_targets = [result.target_name for result in results if result.changed]
    branch_names = [result.branch_name for result in results if result.branch_name]
    pr_numbers = [result.pr_number for result in results if result.pr_number]
    with Path(output_path).open("a", encoding="utf-8") as handle:
        handle.write(f"changed={'true' if bool(changed_targets) else 'false'}\n")
        handle.write(f"changed_targets={','.join(changed_targets)}\n")
        handle.write(f"branch_names={','.join(branch_names)}\n")
        handle.write(f"pr_numbers={','.join(pr_numbers)}\n")


def process_target(
    target: dict[str, Any],
    image: str,
    gitops_repo_dir: Path | None,
    dry_run: bool,
    push: bool,
    open_pr: bool,
    token: str,
) -> TargetResult:
    repo_dir = gitops_repo_dir.resolve() if gitops_repo_dir else clone_repo(target, token)
    created_temp_dir = gitops_repo_dir is None
    result = TargetResult(target_name=target["name"], changed=False, manifest_path=target["manifest_path"])
    try:
        ensure_git_identity(repo_dir)
        manifest_path = repo_dir / target["manifest_path"]
        if not manifest_path.exists():
            raise FileNotFoundError(f"manifest not found: {manifest_path}")

        changed, original, updated = patch_target_manifest(
            manifest_path=manifest_path,
            target=target,
            image=image,
        )
        diff = render_diff(manifest_path, original, updated)
        print(diff or f"No change needed for {target['name']}")
        if not changed:
            return result

        result.changed = True
        branch_name = build_branch_name(image, target["name"])
        title = build_commit_title(target["name"], image)
        body = build_pr_body(target, image)

        if dry_run and not push:
            manifest_path.write_text(original, encoding="utf-8")
            result.branch_name = branch_name
            return result

        run(["git", "checkout", "-b", branch_name], cwd=repo_dir)
        run(["git", "add", target["manifest_path"]], cwd=repo_dir)
        run(["git", "commit", "-m", title], cwd=repo_dir)

        if push:
            run(["git", "push", "--set-upstream", "origin", branch_name], cwd=repo_dir)
        pr_number = ""
        if open_pr:
            pr_number = open_or_update_pr(repo_dir, target, branch_name, title, body)

        result.branch_name = branch_name
        result.pr_number = pr_number
        return result
    finally:
        if created_temp_dir:
            shutil.rmtree(repo_dir, ignore_errors=True)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Open GitOps PRs for published images")
    parser.add_argument("--config", default="deploy/gitops-targets.json")
    parser.add_argument("--target")
    parser.add_argument("--all-targets", action="store_true")
    parser.add_argument("--image", required=True)
    parser.add_argument("--gitops-repo-dir", help="Use an existing local GitOps checkout for dry-run validation")
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument("--push", action="store_true")
    parser.add_argument("--open-pr", action="store_true")
    parser.add_argument(
        "--github-output",
        help="Path to a GitHub Actions output file to append machine-readable results to",
    )
    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)

    config_path = Path(args.config).resolve()
    targets = select_targets(load_targets(config_path), args.target, args.all_targets)

    gitops_repo_dir = Path(args.gitops_repo_dir).resolve() if args.gitops_repo_dir else None
    if gitops_repo_dir and len(targets) != 1:
        raise ValueError("--gitops-repo-dir currently supports exactly one target")

    token = os.environ.get("GH_TOKEN", "")
    if (args.push or args.open_pr) and not token and not gitops_repo_dir:
        raise ValueError("GH_TOKEN is required when --push or --open-pr is used")
    if args.open_pr and not args.push:
        raise ValueError("--open-pr requires --push")

    results: list[TargetResult] = []
    for target in targets:
        results.append(
            process_target(
                target=target,
                image=args.image,
                gitops_repo_dir=gitops_repo_dir,
                dry_run=args.dry_run,
                push=args.push,
                open_pr=args.open_pr,
                token=token,
            )
        )

    output_path = args.github_output or os.environ.get("GITHUB_OUTPUT")
    append_github_outputs(output_path, results)
    return 0


def cli() -> int:
    try:
        return main()
    except Exception as exc:  # noqa: BLE001
        print(f"error: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(cli())
