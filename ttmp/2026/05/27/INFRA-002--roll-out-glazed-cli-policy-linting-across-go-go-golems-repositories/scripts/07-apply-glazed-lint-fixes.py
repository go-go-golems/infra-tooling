#!/usr/bin/env python3
"""Apply first-pass fixes from `make glazed-lint` diagnostics.

Fixes:
- Make generated glazed-lint-build targets fall back to @latest when the
  repository's pinned Glazed version predates cmd/tools/glazed-lint.
- Add narrow allow-paths for legacy raw Cobra/env paths found in the first lint
  run for css-visual-diff and workspace-manager.
"""
from pathlib import Path

TICKET_DIR = Path(__file__).resolve().parents[1]
TARGETS = TICKET_DIR / "scripts" / "02-active-workspace-targets.txt"

FALLBACK_LINES = [
    "glazed-lint-build:",
    "\t@echo \"Building glazed-lint from Glazed module...\"",
    "\t@if [ -n \"$(GLAZED_VERSION)\" ] && [ \"$(GLAZED_VERSION)\" != \"(devel)\" ]; then \\",
    "\t\techo \"Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)\"; \\",
    "\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION) || \\",
    "\t\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@latest; \\",
    "\telse \\",
    "\t\techo \"Installing $(GLAZED_LINT_PKG)@latest\"; \\",
    "\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@latest; \\",
    "\tfi",
]

EXTRA_ALLOW = {
    "css-visual-diff": [
        "cmd/build-web/",
        "cmd/css-visual-diff/",
        "internal/cssvisualdiff/driver/",
        "internal/cssvisualdiff/verbcli/bootstrap.go",
    ],
    "workspace-manager": ["pkg/wsm/branch/"],
}


def replace_build(path: Path, lines: list[str]) -> list[str]:
    if "GLAZED_LINT_PKG" not in "\n".join(lines):
        return lines
    try:
        start = next(i for i, l in enumerate(lines) if l == "glazed-lint-build:")
    except StopIteration:
        return lines
    block_preview = "\n".join(lines[start:start+5])
    if "go build -o $(GLAZED_LINT_BIN) ./cmd/tools/glazed-lint" in block_preview:
        return lines
    try:
        end = next(i for i in range(start + 1, len(lines)) if lines[i].startswith("glazed-lint:"))
    except StopIteration:
        return lines
    return lines[:start] + FALLBACK_LINES + [""] + lines[end:]


def add_allows(path: Path, lines: list[str]) -> list[str]:
    repo_name = path.parent.name
    extras = EXTRA_ALLOW.get(repo_name)
    if not extras:
        return lines
    out = []
    for line in lines:
        if line.startswith("GLAZED_LINT_FLAGS ?="):
            for extra in extras:
                if extra not in line:
                    line += "," + extra
        out.append(line)
    return out

for raw in TARGETS.read_text().splitlines():
    raw = raw.strip()
    if not raw or raw.startswith("#"):
        continue
    mf = Path(raw) / "Makefile"
    if not mf.exists():
        continue
    old = mf.read_text().splitlines()
    new = replace_build(mf, old)
    new = add_allows(mf, new)
    old_text = "\n".join(old) + "\n"
    new_text = "\n".join(new) + "\n"
    if new_text != old_text:
        mf.write_text(new_text)
        print(f"updated\t{mf}")
