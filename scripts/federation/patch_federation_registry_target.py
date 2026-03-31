#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Patch one remote entry inside an embedded federation.registry.json YAML literal block.",
    )
    parser.add_argument("--target-file", required=True)
    parser.add_argument("--remote-id", required=True)
    parser.add_argument("--manifest-url", required=True)
    parser.add_argument("--config-key", default="federation.registry.json")
    parser.add_argument("--enabled", choices=("true", "false"), default="true")
    parser.add_argument("--mode", default=None)
    parser.add_argument("--dry-run", action="store_true")
    return parser.parse_args()


def find_literal_block(lines: list[str], config_key: str) -> tuple[int, int, int]:
    key_line_index = -1
    key_indent = -1
    block_indent = -1
    needle = f"{config_key}: |"
    for idx, line in enumerate(lines):
        if line.strip() == needle:
            key_line_index = idx
            key_indent = len(line) - len(line.lstrip(" "))
            break
    if key_line_index == -1:
        raise ValueError(f"could not find literal block for {config_key!r}")

    for idx in range(key_line_index + 1, len(lines)):
        line = lines[idx]
        if not line.strip():
            continue
        current_indent = len(line) - len(line.lstrip(" "))
        if current_indent <= key_indent:
            raise ValueError(f"literal block for {config_key!r} is empty")
        block_indent = current_indent
        break
    if block_indent == -1:
        raise ValueError(f"literal block for {config_key!r} has no content")

    block_end = len(lines)
    for idx in range(key_line_index + 1, len(lines)):
        line = lines[idx]
        if not line.strip():
            continue
        current_indent = len(line) - len(line.lstrip(" "))
        if current_indent <= key_indent:
            block_end = idx
            break
    return key_line_index, block_end, block_indent


def extract_json_block(lines: list[str], start: int, end: int, block_indent: int) -> str:
    payload_lines: list[str] = []
    prefix = " " * block_indent
    for line in lines[start:end]:
        if line.startswith(prefix):
            payload_lines.append(line[len(prefix) :])
        elif not line.strip():
            payload_lines.append("\n")
        else:
            raise ValueError(f"literal block indentation mismatch near line: {line.rstrip()}")
    return "".join(payload_lines)


def render_json_block(payload: str, block_indent: int) -> list[str]:
    prefix = " " * block_indent
    return [f"{prefix}{line}\n" for line in payload.splitlines()]


def patch_registry(
    registry: dict,
    remote_id: str,
    manifest_url: str,
    enabled: bool,
    mode: str | None,
) -> dict:
    remotes = registry.get("remotes")
    if not isinstance(remotes, list):
        raise ValueError("registry.remotes is missing or not a list")

    matched = False
    for remote in remotes:
        if not isinstance(remote, dict):
            continue
        if remote.get("remoteId") != remote_id:
            continue
        remote["manifestUrl"] = manifest_url
        remote["enabled"] = enabled
        if mode is not None:
            remote["mode"] = mode
        matched = True
        break

    if not matched:
        raise ValueError(f"remoteId {remote_id!r} not found in registry")
    return registry


def main() -> int:
    args = parse_args()
    target_file = Path(args.target_file).resolve()
    original_lines = target_file.read_text(encoding="utf-8").splitlines(keepends=True)

    key_line_index, block_end, block_indent = find_literal_block(original_lines, args.config_key)
    json_start = key_line_index + 1
    raw_json = extract_json_block(original_lines, json_start, block_end, block_indent)
    registry = json.loads(raw_json)

    updated_registry = patch_registry(
        registry=registry,
        remote_id=args.remote_id,
        manifest_url=args.manifest_url,
        enabled=args.enabled == "true",
        mode=args.mode,
    )

    updated_json = json.dumps(updated_registry, indent=2)
    updated_lines = (
        original_lines[:json_start]
        + render_json_block(updated_json, block_indent)
        + original_lines[block_end:]
    )
    updated_text = "".join(updated_lines)

    if args.dry_run:
        print(updated_text)
        return 0

    target_file.write_text(updated_text, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
