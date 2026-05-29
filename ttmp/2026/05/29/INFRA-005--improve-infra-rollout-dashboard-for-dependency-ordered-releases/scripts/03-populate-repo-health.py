#!/usr/bin/env python3
"""Populate lightweight rollout health checks for dashboard panels.

This scanner reads local repository files and records checks that are useful on
repo detail pages and health dashboards. It is intentionally conservative: a
warning means "operator should inspect", not necessarily "CI is broken".
"""
from __future__ import annotations

import argparse
import datetime as dt
import re
import sqlite3
from pathlib import Path

DEFAULT_DB = Path(
    "/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/"
    "INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/"
    "sources/05-rollout-progress.sqlite"
)

SCHEMA = """
CREATE TABLE IF NOT EXISTS repo_health_checks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo TEXT NOT NULL REFERENCES repos(repo) ON DELETE CASCADE,
  category TEXT NOT NULL,
  check_key TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('pass','warn','fail','skip')),
  summary TEXT NOT NULL,
  details TEXT NOT NULL DEFAULT '',
  file_path TEXT,
  scanned_at TEXT NOT NULL,
  UNIQUE(repo, category, check_key)
);
CREATE INDEX IF NOT EXISTS idx_repo_health_checks_repo ON repo_health_checks(repo, category, status);
CREATE INDEX IF NOT EXISTS idx_repo_health_checks_category ON repo_health_checks(category, status, repo);
"""


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")


def read(path: Path) -> str:
    try:
        return path.read_text(errors="ignore")
    except Exception:
        return ""


def add(checks: list[tuple], repo: str, category: str, key: str, status: str, summary: str, details: str = "", file_path: str | None = None) -> None:
    checks.append((repo, category, key, status, summary, details, file_path))


def extract_target_line(makefile: str, target: str) -> str:
    lines = makefile.splitlines()
    for i, line in enumerate(lines):
        stripped = line.strip()
        if stripped == f"{target}:" or stripped.startswith(f"{target}:"):
            body = []
            for nxt in lines[i + 1 : i + 8]:
                if nxt and not nxt.startswith("\t") and not nxt.startswith(" "):
                    break
                body.append(nxt.strip())
            return "\n".join(body)
    return ""


def package_suffix(line: str) -> str:
    if not line:
        return ""
    # Keep the package pattern tail; enough for comparing generate/check coverage.
    parts = line.split()
    pkgs = [p for p in parts if p.startswith("./")]
    return " ".join(pkgs)


def scan_repo(row: sqlite3.Row, ts: str) -> list[tuple]:
    repo = row["repo"]
    path = Path(row["path"] or "")
    checks: list[tuple] = []
    if not path.exists():
        add(checks, repo, "workspace", "checkout_exists", "fail", "Local checkout path does not exist", str(path), row["path"])
        return checks
    go_mod = read(path / "go.mod")
    makefile = read(path / "Makefile")

    if row["needs_logcopter"]:
        m = re.search(r"github\.com/go-go-golems/logcopter\s+(\S+)", go_mod)
        add(checks, repo, "logcopter", "require_version", "pass" if m else "fail", f"logcopter require {'found: ' + m.group(1) if m else 'missing'}", file_path=str(path / "go.mod"))
        has_tool = "tool github.com/go-go-golems/logcopter/cmd/logcopter-gen" in go_mod
        add(checks, repo, "logcopter", "tool_directive", "pass" if has_tool else "warn", "logcopter tool directive present" if has_tool else "logcopter tool directive missing", file_path=str(path / "go.mod"))
        gen_file = path / "logcopter_generate.go"
        add(checks, repo, "logcopter", "generate_file", "pass" if gen_file.exists() else "warn", "logcopter_generate.go present" if gen_file.exists() else "logcopter_generate.go missing", file_path=str(gen_file))
        gen_target = extract_target_line(makefile, "logcopter-generate")
        check_target = extract_target_line(makefile, "logcopter-check")
        add(checks, repo, "logcopter", "make_generate_target", "pass" if gen_target else "warn", "Makefile logcopter-generate target present" if gen_target else "Makefile logcopter-generate target missing", file_path=str(path / "Makefile"))
        add(checks, repo, "logcopter", "make_check_target", "pass" if check_target else "fail", "Makefile logcopter-check target present" if check_target else "Makefile logcopter-check target missing", file_path=str(path / "Makefile"))
        gen_pkgs = package_suffix(gen_target)
        check_pkgs = package_suffix(check_target)
        if gen_target and check_target:
            status = "pass" if gen_pkgs == check_pkgs else "warn"
            add(checks, repo, "logcopter", "generate_check_package_match", status, "logcopter generate/check package lists match" if status == "pass" else "logcopter generate/check package lists differ", f"generate={gen_pkgs}; check={check_pkgs}", str(path / "Makefile"))

    if row["needs_glazed_lint"]:
        has_target = bool(extract_target_line(makefile, "glazed-lint"))
        add(checks, repo, "glazed_lint", "make_target", "pass" if has_target else "fail", "Makefile glazed-lint target present" if has_target else "Makefile glazed-lint target missing", file_path=str(path / "Makefile"))
        m = re.search(r"GLAZED_VERSION\s*\?=\s*(\S+)", makefile)
        add(checks, repo, "glazed_lint", "version_pin", "pass" if m else "warn", f"GLAZED_VERSION {'pinned to ' + m.group(1) if m else 'not found'}", file_path=str(path / "Makefile"))
        has_pkg = "github.com/go-go-golems/glazed/cmd/tools/glazed-lint" in makefile
        add(checks, repo, "glazed_lint", "analyzer_package", "pass" if has_pkg else "warn", "glazed-lint analyzer package configured" if has_pkg else "glazed-lint analyzer package not found", file_path=str(path / "Makefile"))
        allow = re.search(r"GLAZED_LINT_ALLOW_PATHS\s*\?=\s*(.+)", makefile)
        if allow:
            idx = makefile[: allow.start()].count("\n")
            lines = makefile.splitlines()
            prev = "\n".join(lines[max(0, idx - 3) : idx])
            commented = "#" in prev
            add(checks, repo, "glazed_lint", "allow_paths", "pass" if commented else "warn", "GLAZED_LINT_ALLOW_PATHS is documented" if commented else "GLAZED_LINT_ALLOW_PATHS exists without nearby comment", allow.group(1), str(path / "Makefile"))
        else:
            add(checks, repo, "glazed_lint", "allow_paths", "skip", "No GLAZED_LINT_ALLOW_PATHS configured", file_path=str(path / "Makefile"))
        vet_line = extract_target_line(makefile, "glazed-lint")
        pkgs = package_suffix(vet_line)
        add(checks, repo, "glazed_lint", "package_patterns", "pass" if pkgs else "warn", "glazed-lint package patterns found" if pkgs else "glazed-lint package patterns not detected", pkgs, str(path / "Makefile"))
    return checks


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--db", type=Path, default=DEFAULT_DB)
    args = ap.parse_args()
    ts = now()
    con = sqlite3.connect(args.db)
    con.row_factory = sqlite3.Row
    con.executescript(SCHEMA)
    rows = con.execute("SELECT * FROM repos ORDER BY repo").fetchall()
    con.execute("DELETE FROM repo_health_checks")
    count = 0
    for row in rows:
        for repo, category, key, status, summary, details, file_path in scan_repo(row, ts):
            con.execute(
                """
                INSERT INTO repo_health_checks(repo,category,check_key,status,summary,details,file_path,scanned_at)
                VALUES(?,?,?,?,?,?,?,?)
                """,
                (repo, category, key, status, summary, details, file_path, ts),
            )
            count += 1
    con.execute("INSERT INTO events(kind,message,created_at) VALUES(?,?,?)", ("health_scan", f"populated repo health checks: {count} checks", ts))
    con.commit()
    con.close()
    print(f"populated health checks in {args.db}")
    print(f"checks={count}")


if __name__ == "__main__":
    main()
