#!/usr/bin/env python3
"""Repair Makefile shell continuations in generated glazed-lint-build targets."""
from pathlib import Path

TICKET_DIR = Path(__file__).resolve().parents[1]
TARGETS = TICKET_DIR / "scripts" / "02-active-workspace-targets.txt"

GOOD_LINES = [
    "glazed-lint-build:",
    "\t@echo \"Building glazed-lint from Glazed module...\"",
    "\t@if [ -n \"$(GLAZED_VERSION)\" ] && [ \"$(GLAZED_VERSION)\" != \"(devel)\" ]; then \\",
    "\t\techo \"Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)\"; \\",
    "\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \\",
    "\telse \\",
    "\t\techo \"Installing $(GLAZED_LINT_PKG) from workspace/module\"; \\",
    "\t\tGOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \\",
    "\tfi",
]


def fix(path: Path) -> bool:
    old_lines = path.read_text().splitlines()
    text = "\n".join(old_lines) + "\n"
    if "GLAZED_LINT_PKG" not in text:
        return False
    try:
        start = next(i for i, l in enumerate(old_lines) if l == "glazed-lint-build:")
    except StopIteration:
        return False
    # Do not replace Glazed's own local build target.
    block_preview = "\n".join(old_lines[start:start+5])
    if "go build -o $(GLAZED_LINT_BIN) ./cmd/tools/glazed-lint" in block_preview:
        return False
    try:
        end = next(i for i in range(start + 1, len(old_lines)) if old_lines[i].startswith("glazed-lint:"))
    except StopIteration:
        return False
    # Keep blank line immediately before glazed-lint target.
    new_lines = old_lines[:start] + GOOD_LINES + [""] + old_lines[end:]
    new_text = "\n".join(new_lines) + "\n"
    if new_text != text:
        path.write_text(new_text)
        return True
    return False

for raw in TARGETS.read_text().splitlines():
    raw = raw.strip()
    if not raw or raw.startswith("#"):
        continue
    mf = Path(raw) / "Makefile"
    if mf.exists() and fix(mf):
        print(f"fixed\t{mf}")
