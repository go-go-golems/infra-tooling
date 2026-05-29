#!/usr/bin/env python3
"""Populate normalized Go-Go-Golems dependency tables in the INFRA-004 tracker DB.

The original rollout DB stores `repos.upstreams` as JSON because the first use case
was PR tracking. Release trains need something more queryable: exact go.mod
versions, direct/indirect edges, whether the dependency is tracked, and computed
release layers.

This script is intentionally read-only with respect to Git checkouts. It scans
local go.mod files under the Go-Go-Golems workspace, normalizes internal module
edges, and writes derived tables into the existing SQLite DB.
"""
from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import sqlite3
import subprocess
from dataclasses import dataclass
from pathlib import Path

DEFAULT_DB = Path(
    "/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/"
    "INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/"
    "sources/05-rollout-progress.sqlite"
)
DEFAULT_WORKSPACE = Path("/home/manuel/code/wesen/go-go-golems")

SCHEMA = """
CREATE TABLE IF NOT EXISTS internal_modules (
  module TEXT PRIMARY KEY,
  repo TEXT NOT NULL,
  path TEXT,
  in_tracker INTEGER NOT NULL DEFAULT 0,
  tracker_state TEXT,
  tracker_batch TEXT,
  tracker_tag TEXT,
  tracker_release_url TEXT,
  latest_local_tag TEXT,
  head_sha TEXT,
  scanned_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_internal_modules_repo ON internal_modules(repo);
CREATE INDEX IF NOT EXISTS idx_internal_modules_tracker ON internal_modules(in_tracker, tracker_state, repo);

CREATE TABLE IF NOT EXISTS internal_dependency_edges (
  repo TEXT NOT NULL,
  module TEXT NOT NULL,
  dependency_repo TEXT NOT NULL,
  dependency_module TEXT NOT NULL,
  required_version TEXT,
  indirect INTEGER NOT NULL DEFAULT 0,
  replace_target TEXT,
  dependency_in_tracker INTEGER NOT NULL DEFAULT 0,
  dependency_state TEXT,
  dependency_batch TEXT,
  dependency_tracker_tag TEXT,
  dependency_release_url TEXT,
  dependency_latest_local_tag TEXT,
  source TEXT NOT NULL DEFAULT 'go.mod require',
  scanned_at TEXT NOT NULL,
  PRIMARY KEY (repo, dependency_module)
);
CREATE INDEX IF NOT EXISTS idx_internal_dependency_edges_repo ON internal_dependency_edges(repo);
CREATE INDEX IF NOT EXISTS idx_internal_dependency_edges_dep ON internal_dependency_edges(dependency_repo);
CREATE INDEX IF NOT EXISTS idx_internal_dependency_edges_direct ON internal_dependency_edges(indirect, repo, dependency_repo);

CREATE TABLE IF NOT EXISTS release_order_layers (
  train_name TEXT NOT NULL,
  repo TEXT NOT NULL,
  layer INTEGER NOT NULL,
  state TEXT NOT NULL,
  rationale TEXT NOT NULL,
  depends_on_json TEXT NOT NULL DEFAULT '[]',
  dependents_json TEXT NOT NULL DEFAULT '[]',
  scanned_at TEXT NOT NULL,
  PRIMARY KEY (train_name, repo)
);
CREATE INDEX IF NOT EXISTS idx_release_order_layers_train ON release_order_layers(train_name, layer, repo);

CREATE TABLE IF NOT EXISTS dependency_bump_candidates (
  repo TEXT NOT NULL,
  dependency_repo TEXT NOT NULL,
  dependency_module TEXT NOT NULL,
  current_required_version TEXT,
  available_tag TEXT,
  dependency_state TEXT,
  dependency_release_url TEXT,
  direct INTEGER NOT NULL DEFAULT 1,
  priority TEXT NOT NULL,
  reason TEXT NOT NULL,
  scanned_at TEXT NOT NULL,
  PRIMARY KEY (repo, dependency_module)
);
CREATE INDEX IF NOT EXISTS idx_dependency_bump_candidates_priority ON dependency_bump_candidates(priority, dependency_repo, repo);
"""

REQUIRE_LINE = re.compile(r"^\s*([\w./-]+)\s+([^\s]+)(?:\s+//\s*(indirect))?\s*$")
SINGLE_REQUIRE = re.compile(r"^\s*require\s+([\w./-]+)\s+([^\s]+)(?:\s+//\s*(indirect))?\s*$")
MODULE_LINE = re.compile(r"^\s*module\s+(\S+)\s*$")
REPLACE_LINE = re.compile(r"^\s*replace\s+([\w./-]+)(?:\s+[^=\s]+)?\s+=>\s+(.+?)\s*$")


@dataclass
class ModuleInfo:
    module: str
    repo: str
    path: str | None
    in_tracker: bool
    tracker_state: str | None = None
    tracker_batch: str | None = None
    tracker_tag: str | None = None
    tracker_release_url: str | None = None
    latest_local_tag: str | None = None
    head_sha: str | None = None


@dataclass
class RequireEdge:
    module: str
    version: str
    indirect: bool


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")


def run_git(path: Path, *args: str) -> str | None:
    try:
        return subprocess.check_output(["git", "-C", str(path), *args], text=True, stderr=subprocess.DEVNULL).strip() or None
    except Exception:
        return None


def latest_tag(path: Path) -> str | None:
    out = run_git(path, "tag", "--sort=-v:refname")
    if not out:
        return None
    return out.splitlines()[0]


def module_from_go_mod(path: Path) -> str | None:
    go_mod = path / "go.mod"
    if not go_mod.exists():
        return None
    for line in go_mod.read_text(errors="ignore").splitlines():
        m = MODULE_LINE.match(line)
        if m:
            return m.group(1)
    return None


def parse_go_mod(path: Path) -> tuple[list[RequireEdge], dict[str, str]]:
    go_mod = path / "go.mod"
    if not go_mod.exists():
        return [], {}
    requires: list[RequireEdge] = []
    replaces: dict[str, str] = {}
    in_require_block = False
    in_replace_block = False
    for raw in go_mod.read_text(errors="ignore").splitlines():
        line = raw.strip()
        if not line or line.startswith("//"):
            continue
        if line == "require (":
            in_require_block = True
            continue
        if line == "replace (":
            in_replace_block = True
            continue
        if line == ")":
            in_require_block = False
            in_replace_block = False
            continue
        if in_require_block:
            m = REQUIRE_LINE.match(raw)
            if m:
                requires.append(RequireEdge(m.group(1), m.group(2), bool(m.group(3))))
            continue
        if in_replace_block:
            # Block replace lines do not include the leading `replace` keyword.
            parts = raw.split("=>", 1)
            if len(parts) == 2:
                left = parts[0].strip().split()[0]
                replaces[left] = parts[1].strip()
            continue
        m = SINGLE_REQUIRE.match(raw)
        if m:
            requires.append(RequireEdge(m.group(1), m.group(2), bool(m.group(3))))
            continue
        m = REPLACE_LINE.match(raw)
        if m:
            replaces[m.group(1)] = m.group(2)
    return requires, replaces


def repo_from_module(module: str) -> str:
    return module.rstrip("/").split("/")[-1]


def load_tracker(con: sqlite3.Connection) -> dict[str, sqlite3.Row]:
    con.row_factory = sqlite3.Row
    return {r["repo"]: r for r in con.execute("SELECT * FROM repos")}


def discover_modules(workspace: Path, tracker: dict[str, sqlite3.Row]) -> dict[str, ModuleInfo]:
    modules: dict[str, ModuleInfo] = {}
    for go_mod in sorted(workspace.glob("*/go.mod")):
        path = go_mod.parent
        module = module_from_go_mod(path)
        if not module:
            continue
        repo = path.name
        tr = tracker.get(repo)
        modules[module] = ModuleInfo(
            module=module,
            repo=repo,
            path=str(path),
            in_tracker=tr is not None,
            tracker_state=tr["state"] if tr else None,
            tracker_batch=tr["batch_id"] if tr else None,
            tracker_tag=tr["tag"] if tr else None,
            tracker_release_url=tr["release_url"] if tr else None,
            latest_local_tag=latest_tag(path),
            head_sha=run_git(path, "rev-parse", "HEAD"),
        )
    # Preserve tracker rows even if local checkout is missing or module path differs.
    for repo, tr in tracker.items():
        module = tr["module"] or f"github.com/go-go-golems/{repo}"
        if module not in modules:
            modules[module] = ModuleInfo(
                module=module,
                repo=repo,
                path=tr["path"],
                in_tracker=True,
                tracker_state=tr["state"],
                tracker_batch=tr["batch_id"],
                tracker_tag=tr["tag"],
                tracker_release_url=tr["release_url"],
            )
    return modules


def internal_module(module: str) -> bool:
    return module.startswith("github.com/go-go-golems/")


def topological_layers(candidates: set[str], deps: dict[str, set[str]]) -> list[list[str]]:
    remaining = set(candidates)
    layers: list[list[str]] = []
    while remaining:
        layer = sorted(repo for repo in remaining if not (deps.get(repo, set()) & remaining))
        if not layer:
            # Keep cycles visible rather than failing silently.
            layers.append(sorted(remaining))
            break
        layers.append(layer)
        remaining -= set(layer)
    return layers


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--db", type=Path, default=DEFAULT_DB)
    ap.add_argument("--workspace", type=Path, default=DEFAULT_WORKSPACE)
    args = ap.parse_args()

    ts = now()
    con = sqlite3.connect(args.db)
    con.row_factory = sqlite3.Row
    con.executescript(SCHEMA)
    tracker = load_tracker(con)
    modules = discover_modules(args.workspace, tracker)
    module_by_repo = {m.repo: m for m in modules.values()}

    con.execute("DELETE FROM internal_modules")
    con.execute("DELETE FROM internal_dependency_edges")
    con.execute("DELETE FROM release_order_layers")
    con.execute("DELETE FROM dependency_bump_candidates")

    for m in modules.values():
        con.execute(
            """
            INSERT INTO internal_modules(module,repo,path,in_tracker,tracker_state,tracker_batch,tracker_tag,tracker_release_url,latest_local_tag,head_sha,scanned_at)
            VALUES(?,?,?,?,?,?,?,?,?,?,?)
            """,
            (m.module, m.repo, m.path, int(m.in_tracker), m.tracker_state, m.tracker_batch, m.tracker_tag, m.tracker_release_url, m.latest_local_tag, m.head_sha, ts),
        )

    edge_count = 0
    direct_deps: dict[str, set[str]] = {}
    for module, m in modules.items():
        if not m.path:
            continue
        requires, replaces = parse_go_mod(Path(m.path))
        for req in requires:
            if not internal_module(req.module) or req.module == module:
                continue
            dep = modules.get(req.module)
            dep_repo = dep.repo if dep else repo_from_module(req.module)
            dep_tracker = tracker.get(dep_repo)
            dep_latest = dep.latest_local_tag if dep else None
            con.execute(
                """
                INSERT INTO internal_dependency_edges(repo,module,dependency_repo,dependency_module,required_version,indirect,replace_target,dependency_in_tracker,dependency_state,dependency_batch,dependency_tracker_tag,dependency_release_url,dependency_latest_local_tag,source,scanned_at)
                VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
                """,
                (
                    m.repo,
                    module,
                    dep_repo,
                    req.module,
                    req.version,
                    int(req.indirect),
                    replaces.get(req.module),
                    int(dep_tracker is not None),
                    dep_tracker["state"] if dep_tracker else None,
                    dep_tracker["batch_id"] if dep_tracker else None,
                    dep_tracker["tag"] if dep_tracker else None,
                    dep_tracker["release_url"] if dep_tracker else None,
                    dep_latest,
                    "go.mod require",
                    ts,
                ),
            )
            edge_count += 1
            if not req.indirect:
                direct_deps.setdefault(m.repo, set()).add(dep_repo)

            available = (dep_tracker["tag"] if dep_tracker and dep_tracker["tag"] else None) or dep_latest
            if available and available != req.version:
                dep_state = dep_tracker["state"] if dep_tracker else None
                priority = "high" if dep_state in {"released", "main_actions_verified"} else "normal"
                reason = f"{dep_repo} has available tag {available}; {m.repo} requires {req.version}."
                con.execute(
                    """
                    INSERT INTO dependency_bump_candidates(repo,dependency_repo,dependency_module,current_required_version,available_tag,dependency_state,dependency_release_url,direct,priority,reason,scanned_at)
                    VALUES(?,?,?,?,?,?,?,?,?,?,?)
                    """,
                    (m.repo, dep_repo, req.module, req.version, available, dep_state, dep_tracker["release_url"] if dep_tracker else None, int(not req.indirect), priority, reason, ts),
                )

    # Release layers based on actual direct go.mod internal dependencies.
    candidates = {repo for repo, tr in tracker.items() if tr["state"] == "main_actions_verified" and not tr["release_url"]}
    deps = {repo: {d for d in direct_deps.get(repo, set()) if d in candidates} for repo in candidates}
    reverse: dict[str, list[str]] = {repo: [] for repo in candidates}
    for repo, ds in deps.items():
        for dep in ds:
            reverse.setdefault(dep, []).append(repo)
    for layer_num, layer in enumerate(topological_layers(candidates, deps), 1):
        for repo in layer:
            depends = sorted(deps.get(repo, set()))
            dependents = sorted(reverse.get(repo, []))
            rationale = "No unreleased direct internal go.mod dependency in this train." if not depends else "Wait for " + ", ".join(depends) + "."
            con.execute(
                """
                INSERT INTO release_order_layers(train_name,repo,layer,state,rationale,depends_on_json,dependents_json,scanned_at)
                VALUES(?,?,?,?,?,?,?,?)
                """,
                ("verified_unreleased_go_mod_direct", repo, layer_num, tracker[repo]["state"], rationale, json.dumps(depends), json.dumps(dependents), ts),
            )

    # Release layers based on tracker upstreams, preserving the original rollout dependency intent.
    tracker_deps: dict[str, set[str]] = {}
    for repo in candidates:
        upstreams = set(json.loads(tracker[repo]["upstreams"] or "[]"))
        tracker_deps[repo] = {u for u in upstreams if u in candidates}
    reverse = {repo: [] for repo in candidates}
    for repo, ds in tracker_deps.items():
        for dep in ds:
            reverse.setdefault(dep, []).append(repo)
    for layer_num, layer in enumerate(topological_layers(candidates, tracker_deps), 1):
        for repo in layer:
            depends = sorted(tracker_deps.get(repo, set()))
            dependents = sorted(reverse.get(repo, []))
            rationale = "No unreleased tracker upstream in this train." if not depends else "Wait for " + ", ".join(depends) + "."
            con.execute(
                """
                INSERT INTO release_order_layers(train_name,repo,layer,state,rationale,depends_on_json,dependents_json,scanned_at)
                VALUES(?,?,?,?,?,?,?,?)
                """,
                ("verified_unreleased_tracker_upstreams", repo, layer_num, tracker[repo]["state"], rationale, json.dumps(depends), json.dumps(dependents), ts),
            )

    con.execute(
        "INSERT INTO events(kind,message,created_at) VALUES(?,?,?)",
        ("dependency_scan", f"populated internal dependency tables: {len(modules)} modules, {edge_count} internal go.mod edges", ts),
    )
    con.commit()
    con.close()
    print(f"populated {args.db}")
    print(f"modules={len(modules)} internal_edges={edge_count}")


if __name__ == "__main__":
    main()
