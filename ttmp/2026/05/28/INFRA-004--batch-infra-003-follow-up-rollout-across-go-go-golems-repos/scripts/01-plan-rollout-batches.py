#!/usr/bin/env python3
"""Build INFRA-004 rollout batches from the INFRA-003 follow-up inventory.

The script reads the machine-readable inventory, inspects local go.mod files for
first-party go-go-golems dependencies, and emits JSON/TSV/Markdown summaries for
batched rollout planning.
"""
from __future__ import annotations

import json
import re
from collections import defaultdict, deque
from pathlib import Path

ROOT = Path("/home/manuel/code/wesen/go-go-golems")
INV = Path("/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/41-repository-follow-up-inventory.json")
OUT = Path("/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources")

FIRST_PARTY_PREFIXES = (
    "github.com/go-go-golems/",
)


def parse_requires(go_mod: Path) -> list[str]:
    if not go_mod.exists():
        return []
    deps: list[str] = []
    for line in go_mod.read_text().splitlines():
        line = line.strip()
        if not line or line.startswith("//") or line in {"require (", ")"}:
            continue
        m = re.match(r"(github\.com/go-go-golems/[A-Za-z0-9_.\-/]+)\s+", line)
        if m:
            deps.append(m.group(1))
    return deps


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    rows = json.loads(INV.read_text())
    flagged = [r for r in rows if any(r.get(k) for k in [
        "needs_logcopter_addition", "needs_docsctl_cicd_push", "needs_glazed_linting_added", "needs_xgoja_bindings"
    ])]
    module_to_repo = {r["module"]: r["repo"] for r in rows if str(r.get("module", "")).startswith("github.com/go-go-golems/")}
    repo_to_row = {r["repo"]: r for r in rows}

    enriched = []
    first_party_edges = defaultdict(list)  # repo -> upstream repos it imports
    downstream_count = defaultdict(int)
    for r in flagged:
        repo = r["repo"]
        path = Path(r["path"])
        deps = parse_requires(path / "go.mod")
        upstreams = sorted({module_to_repo[d] for d in deps if d in module_to_repo and module_to_repo[d] != repo})
        for u in upstreams:
            first_party_edges[repo].append(u)
            downstream_count[u] += 1
        r2 = dict(r)
        r2["first_party_upstreams"] = upstreams
        r2["track_count"] = sum(bool(r2.get(k)) for k in [
            "needs_logcopter_addition", "needs_docsctl_cicd_push", "needs_glazed_linting_added", "needs_xgoja_bindings"
        ])
        enriched.append(r2)

    def flags(r):
        return {
            "logcopter": bool(r.get("needs_logcopter_addition")),
            "docsctl": bool(r.get("needs_docsctl_cicd_push")),
            "glazed_lint": bool(r.get("needs_glazed_linting_added")),
            "xgoja": bool(r.get("needs_xgoja_bindings")),
        }

    xgoja = sorted([r for r in enriched if r.get("needs_xgoja_bindings")], key=lambda r: r["repo"])
    docs_glazed_leaf = sorted([
        r for r in enriched
        if r.get("needs_docsctl_cicd_push") and r.get("needs_glazed_linting_added") and not r.get("needs_xgoja_bindings")
    ], key=lambda r: (len(r["first_party_upstreams"]), r["repo"]))
    glazed_only = sorted([
        r for r in enriched
        if r.get("needs_glazed_linting_added") and not r.get("needs_docsctl_cicd_push") and not r.get("needs_xgoja_bindings")
    ], key=lambda r: (len(r["first_party_upstreams"]), r["repo"]))
    log_only = sorted([
        r for r in enriched
        if r.get("needs_logcopter_addition") and not r.get("needs_docsctl_cicd_push") and not r.get("needs_glazed_linting_added") and not r.get("needs_xgoja_bindings")
    ], key=lambda r: (downstream_count[r["repo"]], r["repo"]))
    docs_only = sorted([
        r for r in enriched
        if r.get("needs_docsctl_cicd_push") and not r.get("needs_logcopter_addition") and not r.get("needs_glazed_linting_added") and not r.get("needs_xgoja_bindings")
    ], key=lambda r: r["repo"])

    # 5 operational batches. They are track/risk batches, not necessarily a single PR wave.
    batches = [
        {"id": "B1", "name": "foundation and upstream libraries", "policy": "Run first where needed as dependencies; avoid downstream PRs until these are merged/released.", "repos": sorted([r for r in log_only + glazed_only if downstream_count[r["repo"]] > 0 or r["repo"] in {"logcopter", "infra-tooling", "common-sense", "dmeta", "esper", "go-sqlite-regexp"}], key=lambda r: (r["repo"] != "logcopter", r["repo"]))},
        {"id": "B2", "name": "leaf logcopter-only repositories", "policy": "Low API risk; one baseline PR per repo, run validation, merge before release train tagging.", "repos": [r for r in log_only if downstream_count[r["repo"]] == 0 and r["repo"] not in {"logcopter"}]},
        {"id": "B3", "name": "Glazed linting without docsctl", "policy": "Add glazed-lint with logcopter where safe; no docs publish/Vault work.", "repos": [r for r in glazed_only if r["repo"] not in {x["repo"] for x in (log_only + glazed_only) if downstream_count[x["repo"]] > 0}]},
        {"id": "B4", "name": "docsctl + Glazed CLI leaf packages", "policy": "Add docsctl release job only after local help export/validate succeeds and Vault role is ready.", "repos": docs_glazed_leaf + docs_only},
        {"id": "B5", "name": "xgoja provider/API-intent candidates", "policy": "Do not implement provider bindings until API intent is confirmed; may still do logcopter/glazed baseline separately.", "repos": xgoja},
    ]

    # de-duplicate repos in order across batches.
    seen = set()
    for b in batches:
        unique = []
        for r in b["repos"]:
            if r["repo"] not in seen:
                unique.append(r)
                seen.add(r["repo"])
        b["repos"] = unique

    data = {"root": str(ROOT), "inventory": str(INV), "batches": []}
    md = [
        "---",
        "Title: INFRA-004 rollout batches",
        "Ticket: INFRA-004",
        "Status: active",
        "Topics:",
        "  - automation",
        "  - cli",
        "  - release",
        "  - docsctl",
        "  - logcopter",
        "  - github",
        "DocType: reference",
        "Intent: short-term",
        "Owners: []",
        "RelatedFiles: []",
        "ExternalSources: []",
        "Summary: Generated human-readable rollout batch plan derived from the INFRA-003 inventory.",
        "LastUpdated: 2026-05-28T00:00:00-04:00",
        "WhatFor: Use as a quick scan of INFRA-004 repository batches and rollout tracks.",
        "WhenToUse: When selecting the next rollout PR wave or explaining batch membership.",
        "---",
        "",
        "# INFRA-004 rollout batches",
        "",
        f"Inventory: `{INV}`",
        "",
    ]
    tsv = ["batch\trepo\tlogcopter\tdocsctl\tglazed_lint\txgoja\tmodule\tfirst_party_upstreams"]
    for b in batches:
        md += [f"## {b['id']} — {b['name']}", "", b["policy"], "", "|repo|logcopter|docsctl|glazed lint|xgoja|upstreams|", "|---|---:|---:|---:|---:|---|"]
        brepos = []
        for r in b["repos"]:
            f = flags(r)
            upstreams = ", ".join(r["first_party_upstreams"])
            md.append(f"|`{r['repo']}`|{'yes' if f['logcopter'] else ''}|{'yes' if f['docsctl'] else ''}|{'yes' if f['glazed_lint'] else ''}|{'yes' if f['xgoja'] else ''}|{upstreams}|")
            tsv.append("\t".join([b["id"], r["repo"], *("yes" if f[k] else "" for k in ["logcopter", "docsctl", "glazed_lint", "xgoja"]), r["module"], ",".join(r["first_party_upstreams"]) or "-"]))
            brepos.append({"repo": r["repo"], "path": r["path"], "module": r["module"], "flags": f, "first_party_upstreams": r["first_party_upstreams"]})
        md.append("")
        data["batches"].append({"id": b["id"], "name": b["name"], "policy": b["policy"], "repos": brepos})

    (OUT / "01-rollout-batches.json").write_text(json.dumps(data, indent=2) + "\n")
    (OUT / "02-rollout-batches.tsv").write_text("\n".join(tsv) + "\n")
    (OUT / "03-rollout-batches.md").write_text("\n".join(md) + "\n")
    print(f"wrote {OUT / '01-rollout-batches.json'}")
    print(f"flagged={len(flagged)} batched={len(seen)}")

if __name__ == "__main__":
    main()
