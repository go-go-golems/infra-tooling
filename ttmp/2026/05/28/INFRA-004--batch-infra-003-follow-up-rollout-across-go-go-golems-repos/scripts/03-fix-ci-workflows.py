#!/usr/bin/env python3
"""
Update CI workflows in a repo to match the canonical go-template patterns.

For each workflow file, this reads the repo's existing file, identifies key fields
that need updating (setup-go version, go-version source, golangci-lint-action version, etc.),
and rewrites only those fields while preserving repo-specific content.

Usage:
    python3 03-fix-ci-workflows.py <repo-dir> [--dry-run] [--write-golangci-version]
"""
import argparse
import os
import re
import sys
import yaml


def read_file(path):
    with open(path) as f:
        return f.read()


def write_file(path, content):
    with open(path, "w") as f:
        f.write(content)


def fix_setup_go_blocks(content, filename):
    """Replace setup-go steps to use v6 + go-version-file: go.mod."""
    # Pattern: old setup-go actions
    replacements = [
        # setup-go@v3 or v4 or v5 -> v6
        (r"uses: actions/setup-go@v[3-5]", "uses: actions/setup-go@v6"),
        # Hardcoded go-version: '1.XX' -> go-version-file: go.mod
        (r"go-version:\s*['\"]?(?:>=)?[\d.]+['\"]?", "go-version-file: go.mod"),
        # go-version-file already correct -> skip
    ]
    for pattern, replacement in replacements:
        content = re.sub(pattern, replacement, content)
    return content


def fix_golangci_lint_action(content):
    """Replace golangci-lint-action with v9 + version-file."""
    # Old: uses: golangci/golangci-lint-action@v3.1.0 or @v3 or @v4 etc
    content = re.sub(
        r"uses: golangci/golangci-lint-action@v[0-9.]+",
        "uses: golangci/golangci-lint-action@v9",
        content,
    )
    # If there's a "version:" key (not "version-file:"), update to version-file
    # But only if there's no version-file already
    if "version-file:" not in content and re.search(r"version:\s*v?[\d.]+", content):
        content = re.sub(r"version:\s*v?[\d.]+", "version-file: .golangci-lint-version", content)
    # If there's no version or version-file at all after the action, add it
    if "golangci-lint-action@v9" in content and "version-file:" not in content and "version:" not in content:
        # Add version-file after the "with:" block start
        content = content.replace(
            "golangci-lint-action@v9\n        with:",
            "golangci-lint-action@v9\n        with:\n          version-file: .golangci-lint-version",
        )
    return content


def fix_actions_checkout(content):
    """Replace old checkout actions with v6."""
    content = re.sub(r"uses: actions/checkout@v[3-5]", "uses: actions/checkout@v6", content)
    return content


def add_logcopter_check_to_push(content):
    """Add 'make logcopter-check' step to push.yml if not already present."""
    if "logcopter-check" in content:
        return content
    # Find the step after setup-go and before 'go generate' or 'go test'
    # Insert after the setup-go block
    pattern = r"(.*go-version-file: go\.mod\s*\n.*cache: true\s*\n)"
    replacement = r"\1      -\n        name: Verify logcopter package loggers\n        run: make logcopter-check\n"
    content = re.sub(pattern, replacement, content, count=1)
    return content


def fix_dependency_scanning(content):
    """Fix dependency-scanning.yml to use setup-go@v6 + go-version-file."""
    content = fix_actions_checkout(content)
    content = fix_setup_go_blocks(content, "dependency-scanning.yml")
    return content


def fix_lint(content):
    """Fix lint.yml to use setup-go@v6 + go-version-file + golangci-lint-action@v9."""
    content = fix_actions_checkout(content)
    content = fix_setup_go_blocks(content, "lint.yml")
    content = fix_golangci_lint_action(content)
    return content


def fix_push(content):
    """Fix push.yml to use setup-go@v6 + go-version-file."""
    content = fix_actions_checkout(content)
    content = fix_setup_go_blocks(content, "push.yml")
    return content


def fix_release(content):
    """Fix release.yaml/yml to use setup-go@v6 + go-version-file."""
    content = fix_actions_checkout(content)
    content = fix_setup_go_blocks(content, "release.yaml")
    return content


CANONICAL_GOLANGCI_LINT_VERSION = "v2.11.2\n"

CANONICAL_GOLANGCI_YML = """---
# This file contains the configuration for golangci-lint
# See https://golangci-lint.run/usage/configuration/ for reference

# Defines the configuration version.
# The only possible value is "2".
version: "2"

# Linters configuration
linters:
  # Default set of linters.
  default: none
  # Enable specific linters
  enable:
    # defaults
    - errcheck
    - govet
    - ineffassign
    - staticcheck
    - unused
    # additional linters
    - exhaustive
    #    - gochecknoglobals
    #    - gochecknoinits
    - nonamedreturns
    - predeclared
  # Exclusions configuration
  exclusions:
    rules:
      - linters:
          - staticcheck
        text: 'SA1019: cli.CreateProcessorLegacy'
  settings:
    errcheck:
      exclude-functions:
        - (io.Closer).Close
        - fmt.Fprintf
        - fmt.Fprintln

# Formatters configuration
formatters:
  enable:
    - gofmt
"""


def process_repo(repo_dir, dry_run=False, write_golangci=False):
    workflows_dir = os.path.join(repo_dir, ".github", "workflows")
    if not os.path.isdir(workflows_dir):
        print(f"  No .github/workflows/ found, skipping.")
        return

    changes = []

    # Fix lint.yml
    lint_path = os.path.join(workflows_dir, "lint.yml")
    if os.path.exists(lint_path):
        orig = read_file(lint_path)
        fixed = fix_lint(orig)
        if fixed != orig:
            changes.append("lint.yml")
            if not dry_run:
                write_file(lint_path, fixed)

    # Fix push.yml
    push_path = os.path.join(workflows_dir, "push.yml")
    if os.path.exists(push_path):
        orig = read_file(push_path)
        fixed = fix_push(orig)
        if fixed != orig:
            changes.append("push.yml")
            if not dry_run:
                write_file(push_path, fixed)

    # Fix dependency-scanning.yml
    ds_path = os.path.join(workflows_dir, "dependency-scanning.yml")
    if os.path.exists(ds_path):
        orig = read_file(ds_path)
        fixed = fix_dependency_scanning(orig)
        if fixed != orig:
            changes.append("dependency-scanning.yml")
            if not dry_run:
                write_file(ds_path, fixed)

    # Fix release.yaml or release.yml
    for rname in ["release.yaml", "release.yml"]:
        rpath = os.path.join(workflows_dir, rname)
        if os.path.exists(rpath):
            orig = read_file(rpath)
            fixed = fix_release(orig)
            if fixed != orig:
                changes.append(rname)
                if not dry_run:
                    write_file(rpath, fixed)

    # Write .golangci-lint-version if missing
    version_path = os.path.join(repo_dir, ".golangci-lint-version")
    if not os.path.exists(version_path):
        changes.append(".golangci-lint-version (created)")
        if not dry_run:
            write_file(version_path, CANONICAL_GOLANGCI_LINT_VERSION)
    else:
        existing = read_file(version_path).strip()
        canonical = CANONICAL_GOLANGCI_LINT_VERSION.strip()
        if existing != canonical:
            changes.append(f".golangci-lint-version ({existing} -> {canonical})")
            if not dry_run:
                write_file(version_path, CANONICAL_GOLANGCI_LINT_VERSION)

    # Optionally write canonical .golangci.yml
    if write_golangci:
        golangci_path = os.path.join(repo_dir, ".golangci.yml")
        orig = read_file(golangci_path) if os.path.exists(golangci_path) else ""
        if orig.strip() != CANONICAL_GOLANGCI_YML.strip():
            changes.append(".golangci.yml (updated to canonical)")
            if not dry_run:
                write_file(golangci_path, CANONICAL_GOLANGCI_YML)

    if changes:
        print(f"  Changed: {', '.join(changes)}")
    else:
        print(f"  No changes needed.")


def main():
    parser = argparse.ArgumentParser(description="Fix CI workflows to match go-template canonical patterns")
    parser.add_argument("repo_dir", help="Path to the repository")
    parser.add_argument("--dry-run", action="store_true", help="Show what would change without writing")
    parser.add_argument("--write-golangci", action="store_true", help="Also overwrite .golangci.yml with canonical")
    args = parser.parse_args()

    process_repo(args.repo_dir, dry_run=args.dry_run, write_golangci=args.write_golangci)


if __name__ == "__main__":
    main()
