#!/usr/bin/env python3
"""Apply Glazed lint Makefile/CI wiring to the active workspace repositories.

This script intentionally edits only repositories listed in
02-active-workspace-targets.txt. It is conservative and idempotent: repositories
that already have glazed-lint targets or CI steps are left mostly unchanged,
except that lint/lintmax and lint workflows are wired when missing.
"""
from __future__ import annotations

from pathlib import Path
import re
import sys

TICKET_DIR = Path(__file__).resolve().parents[1]
TARGETS = TICKET_DIR / "scripts" / "02-active-workspace-targets.txt"

GLAZED_BLOCK = """
GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell GOWORK=off go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/
""".lstrip()

TARGET_BLOCK = """
glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)
""".lstrip()

CI_STEP = """      - name: Run Glazed CLI policy linters
        run: make glazed-lint
"""


def has_dir(repo: Path, name: str) -> bool:
    return (repo / name).is_dir()


def default_dirs(repo: Path) -> str:
    dirs = [f"./{d}/..." for d in ("cmd", "internal", "pkg", "runtime", "examples/cmd", "dev-tools") if has_dir(repo, d)]
    return " ".join(dirs) if dirs else "./..."


def add_phony(text: str) -> str:
    if "glazed-lint-build" in text.splitlines()[0:3] and "glazed-lint" in text.splitlines()[0:3]:
        return text
    lines = text.splitlines()
    for i, line in enumerate(lines[:8]):
        if line.startswith(".PHONY:"):
            if "glazed-lint-build" not in line:
                line += " glazed-lint-build glazed-lint"
                lines[i] = line
            return "\n".join(lines) + ("\n" if text.endswith("\n") else "")
    return ".PHONY: glazed-lint-build glazed-lint\n" + text


def ensure_vars(text: str, repo: Path) -> str:
    if "GLAZED_LINT_BIN" not in text:
        marker_match = re.search(r"^(?:GOLANGCI_LINT_ARGS|GOLANGCI_LINT_BIN|GORELEASER_TARGET|VERSION|MODULE).*\n", text, re.M)
        insert_at = marker_match.end() if marker_match else 0
        text = text[:insert_at] + GLAZED_BLOCK + text[insert_at:]
    if "GLAZED_LINT_DIRS" not in text:
        dirs = default_dirs(repo)
        if "LINT_DIRS :=" in text:
            line = "GLAZED_LINT_DIRS ?= $(LINT_DIRS)\n"
        else:
            line = f"GLAZED_LINT_DIRS ?= {dirs}\n"
        # Put it after GLAZED_LINT_FLAGS if possible.
        m = re.search(r"^GLAZED_LINT_FLAGS .*\n", text, re.M)
        if m:
            text = text[:m.end()] + line + text[m.end():]
        else:
            text = line + text
    return text


def ensure_targets(text: str) -> str:
    if re.search(r"^glazed-lint-build:", text, re.M):
        # Keep existing targets. Only normalize glazed-lint command if it lacks flags.
        if re.search(r"^glazed-lint:.*\n\tgo vet -vettool=\$\(GLAZED_LINT_BIN\) \$\(GLAZED_LINT_DIRS\)", text, re.M):
            text = text.replace("\tgo vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_DIRS)", "\tgo vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)")
        return text
    insert_at = text.find("\n\nlint:")
    if insert_at == -1:
        insert_at = len(text)
    return text[:insert_at] + "\n" + TARGET_BLOCK + text[insert_at:]


def add_dep(target_line: str) -> str:
    if "glazed-lint-build" in target_line:
        return target_line
    if ":" in target_line:
        return target_line.rstrip() + " glazed-lint-build\n"
    return target_line


def vet_prefix(text: str) -> str:
    # Match the repository style. If lint already runs GOWORK=off, keep it.
    return "GOWORK=off " if re.search(r"^\tGOWORK=off .*golangci", text, re.M) else ""


def wire_lint_target(text: str, name: str) -> str:
    m = re.search(rf"^{name}:([^\n]*)\n(?P<body>(?:\t.*\n)+)", text, re.M)
    if not m:
        return text
    block = m.group(0)
    if "go vet -vettool=$(GLAZED_LINT_BIN)" in block:
        # Still ensure dependency.
        new_header = add_dep(block.splitlines()[0] + "\n").rstrip("\n")
        new_block = new_header + "\n" + "\n".join(block.splitlines()[1:]) + "\n"
        return text[:m.start()] + new_block + text[m.end():]
    lines = block.splitlines()
    lines[0] = add_dep(lines[0] + "\n").rstrip("\n")
    prefix = vet_prefix(block)
    lines.append(f"\t{prefix}go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)")
    new_block = "\n".join(lines) + "\n"
    return text[:m.start()] + new_block + text[m.end():]


def update_makefile(repo: Path) -> bool:
    path = repo / "Makefile"
    if not path.exists():
        print(f"SKIP no Makefile: {repo}")
        return False
    old = path.read_text()
    text = old
    text = add_phony(text)
    text = ensure_vars(text, repo)
    text = ensure_targets(text)
    text = wire_lint_target(text, "lint")
    text = wire_lint_target(text, "lintmax")
    if text != old:
        path.write_text(text)
        return True
    return False


def update_ci(repo: Path) -> bool:
    wf = repo / ".github" / "workflows" / "lint.yml"
    if not wf.exists():
        return False
    old = wf.read_text()
    if "Run Glazed CLI policy linters" in old:
        return False
    lines = old.splitlines(keepends=True)
    insert_at = None
    for i, line in enumerate(lines):
        if "golangci-lint-action" in line:
            # insert after the current step, including its with: block
            insert_at = i + 1
            while insert_at < len(lines):
                nxt = lines[insert_at]
                if nxt.startswith("      - name:") or nxt.startswith("      - uses:"):
                    break
                insert_at += 1
            break
        if re.match(r"\s*run: .*(golangci-lint|/tmp/golangci-lint)", line):
            insert_at = i + 1
    if insert_at is None:
        return False
    new = "".join(lines[:insert_at]) + CI_STEP + "".join(lines[insert_at:])
    wf.write_text(new)
    return True


def main() -> int:
    if not TARGETS.exists():
        print(f"missing targets file: {TARGETS}", file=sys.stderr)
        return 2
    changed = []
    for raw in TARGETS.read_text().splitlines():
        raw = raw.strip()
        if not raw or raw.startswith("#"):
            continue
        repo = Path(raw)
        m = update_makefile(repo)
        c = update_ci(repo)
        if m or c:
            changed.append((repo, m, c))
            print(f"changed\t{repo}\tmakefile={m}\tci={c}")
        else:
            print(f"unchanged\t{repo}")
    print(f"changed_count={len(changed)}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
