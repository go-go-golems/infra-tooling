#!/usr/bin/env python3
"""Normalize Makefile blocks produced by 03-apply-glazed-lint-wiring.py.

The first wiring script intentionally captured the exact edit logic, but its
Python string escaped Makefile backslash-newline continuations. This follow-up
normalizes the generated block and keeps it as a traceable repair step.
"""
from __future__ import annotations
from pathlib import Path
import re

TICKET_DIR = Path(__file__).resolve().parents[1]
TARGETS = TICKET_DIR / "scripts" / "02-active-workspace-targets.txt"

GOOD_BUILD = """glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi
"""


def normalize_makefile(path: Path) -> bool:
    old = path.read_text()
    text = old

    # Fix one-line/broken generated build blocks, but leave Glazed's own local
    # `go build ./cmd/tools/glazed-lint` target intact.
    if "GLAZED_LINT_PKG" in text:
        text = re.sub(
            r"glazed-lint-build:\n\t@echo \"Building glazed-lint from Glazed module\.\.\.\"\n\t@if \[ -n \"\$\(GLAZED_VERSION\).*?\n(?=\nglazed-lint:)",
            GOOD_BUILD,
            text,
            flags=re.S,
        )

    # Fix `.PHONY: ... \ glazed-lint-build` generated into a continued phony line.
    text = text.replace("goSec govulncheck \\ glazed-lint-build", "gosec govulncheck \\")
    text = text.replace("gosec govulncheck \\ glazed-lint-build glazed-lint\n        ", "gosec govulncheck \\\n        glazed-lint-build glazed-lint ")
    text = text.replace("govulncheck \\ glazed-lint-build glazed-lint\n        ", "govulncheck \\\n        glazed-lint-build glazed-lint ")
    text = text.replace("\\ glazed-lint-build glazed-lint\n        ", "\\\n        glazed-lint-build glazed-lint ")

    # Repositories whose lint target already runs under GOWORK=off should run
    # the vettool under the same module-resolution mode.
    text = re.sub(
        r"^(\t)(go vet -vettool=\$\(GLAZED_LINT_BIN\) \$\(GLAZED_LINT_FLAGS\) \$\(GLAZED_LINT_DIRS\))$",
        lambda m: m.group(0)
        if "GOWORK=off" not in text[max(0, m.start() - 180):m.start()]
        else f"{m.group(1)}GOWORK=off {m.group(2)}",
        text,
        flags=re.M,
    )

    if text != old:
        path.write_text(text)
        return True
    return False


def main() -> int:
    for raw in TARGETS.read_text().splitlines():
        raw = raw.strip()
        if not raw or raw.startswith("#"):
            continue
        mf = Path(raw) / "Makefile"
        if mf.exists() and normalize_makefile(mf):
            print(f"normalized\t{mf}")
    return 0

if __name__ == "__main__":
    raise SystemExit(main())
