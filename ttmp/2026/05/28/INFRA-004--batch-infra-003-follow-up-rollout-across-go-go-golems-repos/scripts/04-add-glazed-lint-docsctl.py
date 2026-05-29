#!/usr/bin/env python3
"""
Add glazed-lint Makefile targets and publish-docs release job to a repo.

Usage:
    python3 04-add-glazed-lint-docsctl.py <repo-dir> [--dry-run]
"""
import argparse
import os
import re
import subprocess
import sys


def run(cmd, **kwargs):
    return subprocess.run(cmd, shell=True, capture_output=True, text=True, **kwargs)


def get_cmd_binary(repo_dir):
    """Find the primary cmd binary name."""
    cmd_dir = os.path.join(repo_dir, "cmd")
    if not os.path.isdir(cmd_dir):
        return None
    # Find main.go files
    mains = []
    for root, dirs, files in os.walk(cmd_dir):
        for f in files:
            if f == "main.go":
                rel = os.path.relpath(os.path.join(root, f), cmd_dir)
                binary = rel.replace("/main.go", "")
                if "main.go" == rel:
                    # root-level cmd/main.go - use dir name
                    return os.path.basename(repo_dir)
                mains.append(binary)
    if not mains:
        return None
    # Prefer the one that matches the repo name or has no subdirs
    repo_name = os.path.basename(repo_dir)
    for m in mains:
        if m == repo_name or m.replace("-", "") == repo_name.replace("-", ""):
            return m
    # Return first one
    return mains[0]


def get_module_path(repo_dir):
    """Get module path from go.mod."""
    go_mod = os.path.join(repo_dir, "go.mod")
    with open(go_mod) as f:
        for line in f:
            if line.startswith("module "):
                return line.split()[1]
    return None


def get_package_name(module_path, repo_name):
    """Derive docsctl package name from module path."""
    # Usually the last segment of the module path
    return module_path.split("/")[-1] if module_path else repo_name


def add_glazed_lint_makefile(repo_dir, module_path, dry_run=False):
    """Add glazed-lint targets to Makefile."""
    makefile = os.path.join(repo_dir, "Makefile")
    if not os.path.exists(makefile):
        return "no Makefile"
    
    with open(makefile) as f:
        content = f.read()
    
    if "glazed-lint" in content:
        return "already has glazed-lint"
    
    # Get the glazed version from go.mod
    version_line = ""
    go_mod = os.path.join(repo_dir, "go.mod")
    with open(go_mod) as f:
        in_require = False
        for line in f:
            if line.strip().startswith("github.com/go-go-golems/glazed "):
                version_line = line.strip().split()[1]
                break
            if "require (" in line:
                in_require = True
                continue
            if in_require and "github.com/go-go-golems/glazed " in line:
                version_line = line.strip().split()[1]
                break
    
    # Default to main if version is not found
    if not version_line or version_line == "v0.0.0":
        version_line = "main"
    
    glazed_lint_block = f"""

GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= {version_line}

.PHONY: glazed-lint-build glazed-lint

glazed-lint-build:
\t@echo "Building glazed-lint from Glazed module..."
\t@if [ -n "$(GLAZED_VERSION)" ]; then \\
\t\techo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \\
\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \\
\telse \\
\t\techo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \\
\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \\
techo "fi"

glazed-lint: glazed-lint-build
\tGOWORK=off go vet -vettool=$(GLAZED_LINT_BIN) ./cmd/... ./pkg/..."""
    
    if not dry_run:
        with open(makefile, 'a') as f:
            f.write(glazed_lint_block)
    
    return f"added (version={version_line})"


def add_publish_docs_workflow(repo_dir, cmd_binary, package_name, dry_run=False):
    """Add publish-docs job to release workflow."""
    if not cmd_binary:
        return "no cmd binary found"
    
    # Find release workflow
    release_file = None
    for name in ["release.yaml", "release.yml"]:
        path = os.path.join(repo_dir, ".github", "workflows", name)
        if os.path.exists(path):
            release_file = path
            break
    
    if not release_file:
        return "no release workflow"
    
    with open(release_file) as f:
        content = f.read()
    
    if "publish-docsctl" in content or "publish-docs" in content:
        return "already has publish-docs"
    
    # Find the goreleaser-merge job to insert publish-docs before it
    # The publish-docs job should go right before goreleaser-merge
    publish_docs_job = f"""
  publish-docs:
    name: Publish docs
    needs:
      - goreleaser-merge
    if: ${{{{ false && startsWith(github.ref, 'refs/tags/v') }}}}
    uses: go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@main
    with:
      package_name: {package_name}
      package_version: ${{{{ github.ref_name }}}}
      export_command: GOWORK=off go run ./cmd/{cmd_binary} help export --format sqlite --output-path .docsctl/help.sqlite
      sqlite_path: .docsctl/help.sqlite
      docsctl_install_command: go install github.com/go-go-golems/glazed/cmd/docsctl@latest
      vault_role: docsctl-{package_name}-publisher
      vault_token_role: docsctl-{package_name}-publisher
      registry_url: https://docs-registry.yolo.scapegoat.dev
      verify_packages_url: https://docs.yolo.scapegoat.dev/api/packages
      verify_publish: true

"""
    
    # Insert before "  goreleaser-merge:"
    if "  goreleaser-merge:" in content:
        content = content.replace("  goreleaser-merge:", publish_docs_job + "  goreleaser-merge:")
    else:
        return "no goreleaser-merge job found"
    
    if not dry_run:
        with open(release_file, 'w') as f:
            f.write(content)
    
    return f"added (cmd={cmd_binary}, package={package_name})"


def add_glazed_lint_to_push_ci(repo_dir, dry_run=False):
    """Add make glazed-lint step to push.yml CI."""
    push_file = os.path.join(repo_dir, ".github", "workflows", "push.yml")
    if not os.path.exists(push_file):
        return "no push.yml"
    
    with open(push_file) as f:
        content = f.read()
    
    if "glazed-lint" in content:
        return "already has glazed-lint in CI"
    
    # Add after logcopter-check step
    if "logcopter-check" in content:
        insertion = """
      - name: Verify Glazed CLI policy
        run: make glazed-lint
"""
        content = content.replace(
            "        run: make logcopter-check\n",
            "        run: make logcopter-check\n" + insertion
        )
    else:
        return "no logcopter-check step found"
    
    if not dry_run:
        with open(push_file, 'w') as f:
            f.write(content)
    
    return "added to CI"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("repo_dir")
    parser.add_argument("--dry-run", action="store_true")
    args = parser.parse_args()
    
    repo_dir = os.path.abspath(args.repo_dir)
    repo_name = os.path.basename(repo_dir)
    module_path = get_module_path(repo_dir)
    cmd_binary = get_cmd_binary(repo_dir)
    package_name = get_package_name(module_path, repo_name)
    
    print(f"  repo: {repo_name}")
    print(f"  module: {module_path}")
    print(f"  cmd: {cmd_binary}")
    print(f"  package: {package_name}")
    
    # Add glazed-lint to Makefile
    result = add_glazed_lint_makefile(repo_dir, module_path, args.dry_run)
    print(f"  glazed-lint: {result}")
    
    # Add glazed-lint to CI
    result = add_glazed_lint_to_push_ci(repo_dir, args.dry_run)
    print(f"  CI glazed-lint: {result}")
    
    # Add publish-docs to release workflow
    if cmd_binary:
        result = add_publish_docs_workflow(repo_dir, cmd_binary, package_name, args.dry_run)
        print(f"  publish-docs: {result}")
    else:
        print(f"  publish-docs: skipped (no cmd binary)")


if __name__ == "__main__":
    main()
