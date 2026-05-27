#!/usr/bin/env python3
"""Add narrow allow-paths for existing legacy CLI bridge/tool code.

These allow-paths come from the second `make glazed-lint` pass. They are kept
narrow so new commands outside these paths are still checked by the analyzer.
"""
from pathlib import Path

ALLOWS = {
    "discord-bot": [
        "pkg/framework/",
        "pkg/botcli/bootstrap.go",
        "cmd/discord-bot/",
    ],
    "go-go-goja": [
        "cmd/gen-dts/",
        "cmd/bun-demo/",
        "cmd/jsverbs-example/",
        "cmd/goja-repl/",
        "pkg/hashiplugin/contract/internal/cmd/generate/",
        "pkg/jsverbrepos/bootstrap.go",
        "pkg/jsverbscli/",
        "pkg/replessay/handler.go",
    ],
    "go-minitrace": [
        "cmd/build-web/",
        "cmd/go-minitrace/cmds/annotate/",
        "cmd/go-minitrace/cmds/query/commands.go",
        "cmd/go-minitrace/cmds/serve/serve.go",
    ],
    "loupedeck": [
        "examples/cmd/",
        "cmd/loupedeck/cmds/doc/",
        "cmd/loupedeck/cmds/verbs/bootstrap.go",
        "cmd/loupedeck/main.go",
    ],
}

ROOT = Path("/home/manuel/workspaces/2026-05-24/add-js-providers")
for repo, extras in ALLOWS.items():
    mf = ROOT / repo / "Makefile"
    lines = mf.read_text().splitlines()
    out = []
    changed = False
    for line in lines:
        if line.startswith("GLAZED_LINT_FLAGS ?="):
            for extra in extras:
                if extra not in line:
                    line += "," + extra
                    changed = True
        out.append(line)
    if changed:
        mf.write_text("\n".join(out) + "\n")
        print(f"updated\t{mf}")
