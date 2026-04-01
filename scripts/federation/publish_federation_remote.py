#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Upload a built federation remote to S3-compatible object storage.",
    )
    parser.add_argument("--source-dir", required=True)
    parser.add_argument("--remote-id", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--bucket", required=True)
    parser.add_argument("--endpoint", required=True)
    parser.add_argument("--region", required=True)
    parser.add_argument("--public-base-url", required=True)
    parser.add_argument(
        "--prefix-template",
        default="remotes/{remote_id}/versions/{version}",
        help="Object-storage key prefix template.",
    )
    parser.add_argument("--dry-run", action="store_true")
    parser.add_argument(
        "--github-output",
        default=os.environ.get("GITHUB_OUTPUT", ""),
        help="Optional GitHub Actions output file.",
    )
    return parser.parse_args()


def require_file(path: Path, label: str) -> None:
    if not path.is_file():
        raise SystemExit(f"missing {label}: {path}")


def write_github_outputs(path: str, values: dict[str, str]) -> None:
    if not path:
        return
    output_path = Path(path)
    with output_path.open("a", encoding="utf-8") as handle:
        for key, value in values.items():
            handle.write(f"{key}={value}\n")


def main() -> int:
    args = parse_args()
    source_dir = Path(args.source_dir).resolve()
    if not source_dir.is_dir():
        raise SystemExit(f"source dir does not exist: {source_dir}")

    manifest_path = source_dir / "mf-manifest.json"
    contract_path = source_dir / f"{args.remote_id}-host-contract.js"
    require_file(manifest_path, "manifest")
    require_file(contract_path, "contract bundle")

    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    if manifest.get("remoteId") != args.remote_id:
        raise SystemExit(
            f"manifest remoteId mismatch: expected {args.remote_id}, got {manifest.get('remoteId')}",
        )

    entry = manifest.get("contract", {}).get("entry")
    if entry != f"./{args.remote_id}-host-contract.js":
        raise SystemExit(
            "manifest contract entry mismatch: "
            f"expected ./{args.remote_id}-host-contract.js, got {entry}",
        )

    prefix = args.prefix_template.format(remote_id=args.remote_id, version=args.version).strip("/")
    base_url = args.public_base_url.rstrip("/")
    manifest_url = f"{base_url}/{prefix}/mf-manifest.json"
    destination = f"s3://{args.bucket}/{prefix}/"

    print("Federation publish plan")
    print(f"- source_dir: {source_dir}")
    print(f"- remote_id: {args.remote_id}")
    print(f"- version: {args.version}")
    print(f"- bucket: {args.bucket}")
    print(f"- endpoint: {args.endpoint}")
    print(f"- region: {args.region}")
    print(f"- destination: {destination}")
    print(f"- manifest_url: {manifest_url}")
    print(f"- dry_run: {'true' if args.dry_run else 'false'}")

    outputs = {
        "remote_prefix": prefix,
        "remote_version": args.version,
        "manifest_url": manifest_url,
    }
    write_github_outputs(args.github_output, outputs)

    if args.dry_run:
        return 0

    aws_env = os.environ.copy()
    if (
        "AWS_ACCESS_KEY_ID" not in aws_env
        and "HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID" in aws_env
    ):
        aws_env["AWS_ACCESS_KEY_ID"] = aws_env["HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID"]
    if (
        "AWS_SECRET_ACCESS_KEY" not in aws_env
        and "HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY" in aws_env
    ):
        aws_env["AWS_SECRET_ACCESS_KEY"] = aws_env["HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY"]

    command = [
        "aws",
        "--endpoint-url",
        args.endpoint,
        "--region",
        args.region,
        "s3",
        "cp",
        str(source_dir),
        destination,
        "--recursive",
        "--cache-control",
        "public,max-age=31536000,immutable",
    ]
    subprocess.run(command, check=True, env=aws_env)
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except subprocess.CalledProcessError as error:
        print(f"command failed with exit code {error.returncode}: {error.cmd}", file=sys.stderr)
        raise
