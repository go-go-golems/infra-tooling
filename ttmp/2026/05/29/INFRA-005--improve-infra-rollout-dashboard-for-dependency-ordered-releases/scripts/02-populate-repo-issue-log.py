#!/usr/bin/env python3
"""Populate issue/fix history tables from rollout events and validations.

The rollout tracker originally stored a chronological `events` stream and a
`validations` table. Those are useful, but a repo detail page needs a more
structured view: what issue was found, how it was classified, what fixed it,
and what evidence proved it was fixed.

This script derives an initial issue log from existing events/validations. It is
safe to rerun: it rebuilds the derived issue tables from the source tables.
"""
from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import sqlite3
from collections import defaultdict
from pathlib import Path

DEFAULT_DB = Path(
    "/home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/"
    "INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/"
    "sources/05-rollout-progress.sqlite"
)

SCHEMA = """
CREATE TABLE IF NOT EXISTS repo_issue_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo TEXT NOT NULL REFERENCES repos(repo) ON DELETE CASCADE,
  issue_key TEXT NOT NULL,
  category TEXT NOT NULL,
  severity TEXT NOT NULL DEFAULT 'medium',
  title TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'observed',
  detected_at TEXT,
  detected_by TEXT,
  evidence_summary TEXT NOT NULL DEFAULT '',
  root_cause TEXT NOT NULL DEFAULT '',
  fix_summary TEXT NOT NULL DEFAULT '',
  fixed_at TEXT,
  fix_commits_json TEXT NOT NULL DEFAULT '[]',
  validation_summary TEXT NOT NULL DEFAULT '',
  source_refs_json TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(repo, issue_key)
);
CREATE INDEX IF NOT EXISTS idx_repo_issue_log_repo ON repo_issue_log(repo, status, category);
CREATE INDEX IF NOT EXISTS idx_repo_issue_log_category ON repo_issue_log(category, status, repo);

CREATE TABLE IF NOT EXISTS repo_issue_steps (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  issue_id INTEGER NOT NULL REFERENCES repo_issue_log(id) ON DELETE CASCADE,
  repo TEXT NOT NULL,
  step_time TEXT NOT NULL,
  step_kind TEXT NOT NULL,
  source_table TEXT NOT NULL,
  source_id INTEGER,
  command TEXT,
  status TEXT,
  message TEXT NOT NULL,
  url TEXT,
  commit_sha TEXT,
  artifact_path TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repo_issue_steps_issue ON repo_issue_steps(issue_id, step_time, id);
CREATE INDEX IF NOT EXISTS idx_repo_issue_steps_repo ON repo_issue_steps(repo, step_time, id);
"""

CATEGORIES = [
    (
        "workflow_yaml",
        re.compile(r"push\.yml|workflow|yaml|unit-test step|test step|main push", re.I),
        "GitHub workflow structure or main workflow execution issue",
        "high",
    ),
    (
        "glazed_lint",
        re.compile(r"glazed-lint|glazed lint|GLAZED|Glazed CLI|analyzer", re.I),
        "Glazed CLI policy or analyzer issue",
        "high",
    ),
    (
        "logcopter_generation",
        re.compile(r"logcopter|generated loggers|generated files", re.I),
        "Logcopter generation/check issue",
        "high",
    ),
    (
        "govulncheck",
        re.compile(r"govulncheck|vulnerability check|go-jose|vulnerab", re.I),
        "Go vulnerability scan issue",
        "high",
    ),
    (
        "gosec",
        re.compile(r"gosec|G\d{3}|security scan|suppressions", re.I),
        "GoSec security scan issue",
        "high",
    ),
    (
        "dependency_review",
        re.compile(r"dependency[- ]review|Dependency Review|dependency graph", re.I),
        "Dependency Review or dependency graph issue",
        "medium",
    ),
    (
        "nancy_ossindex",
        re.compile(r"Nancy|OSS Index|401", re.I),
        "Nancy/OSS Index scan issue",
        "medium",
    ),
    (
        "golangci_lint",
        re.compile(r"golangci|staticcheck|errcheck|QF1012|lint findings|lint pass", re.I),
        "golangci-lint/static analysis issue",
        "high",
    ),
    (
        "codex_feedback",
        re.compile(r"Codex|codex_feedback|current-head", re.I),
        "Codex review feedback issue",
        "medium",
    ),
    (
        "generation_artifact",
        re.compile(r"inventory\.db|artifact|accidentally committed|data/", re.I),
        "Generated artifact hygiene issue",
        "medium",
    ),
    (
        "release_or_main_verification",
        re.compile(r"main action|main workflow|main branch|release|tagged|verified post-merge|main_actions_verified", re.I),
        "Release or post-merge main verification issue",
        "medium",
    ),
    (
        "local_validation",
        re.compile(r"go test|validation|preflight|make ", re.I),
        "Local validation issue",
        "medium",
    ),
]

FIX_RE = re.compile(r"\b(fix|fixed|repaired|restored|upgraded|bumped|removed|added|included|validated|verified|passed|merged|released|succeeded|triggered|addressed)\b", re.I)
DETECT_RE = re.compile(r"\b(fail|failed|failure|needs|actionable|blocked|warning|warn|missing|unsupported|unknown|stale|feedback|finding|findings)\b", re.I)
COMMIT_RE = re.compile(r"\b[0-9a-f]{7,40}\b")

ROOT_CAUSES = {
    "workflow_yaml": "Workflow steps were malformed, incomplete, or behaved differently on main than the rollout expected.",
    "glazed_lint": "The Glazed analyzer target, version, package coverage, or allow-path policy did not match the repository's CLI code.",
    "logcopter_generation": "Generated logcopter files or logcopter generate/check package lists were missing, stale, ignored, or inconsistent.",
    "govulncheck": "A vulnerable module version or Go toolchain mismatch caused govulncheck to fail.",
    "gosec": "Gosec reported a security finding or required a scoped rollout suppression for legacy/pre-existing behavior.",
    "dependency_review": "GitHub Dependency Review or dependency graph support was unavailable or reported a blocking dependency issue.",
    "nancy_ossindex": "Nancy/OSS Index scanning was unavailable or returned external-service failures in CI.",
    "golangci_lint": "CI-pinned lint/static analysis exposed findings that needed code cleanup or scoped exception handling.",
    "codex_feedback": "Current-head Codex review feedback required a follow-up commit or retrigger before readiness.",
    "generation_artifact": "A validation or generation command produced repository state that needed cleanup or ignore rules.",
    "release_or_main_verification": "Post-merge or release verification required additional workflow observation or repair.",
    "local_validation": "A local validation command failed or warned and required follow-up before progressing.",
}


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")


def classify(text: str) -> tuple[str, str, str]:
    for key, rx, title, severity in CATEGORIES:
        if rx.search(text):
            return key, title, severity
    return "general_rollout", "General rollout issue or notable event", "medium"


def step_kind(text: str, status: str | None = None) -> str:
    if status == "fail" or DETECT_RE.search(text):
        if FIX_RE.search(text):
            return "fix_progress"
        return "detected"
    if status == "pass" or FIX_RE.search(text):
        return "fix_or_validation"
    if status == "warn":
        return "warning"
    return "note"


def final_status(steps: list[dict]) -> str:
    kinds = [s["step_kind"] for s in steps]
    blocking_indices = [i for i, s in enumerate(steps) if re.search(r"blocked|archived|read-only|discarded", s["message"], re.I)]
    fix_indices = [i for i, s in enumerate(steps) if s["step_kind"] == "fix_or_validation"]
    detected_indices = [i for i, s in enumerate(steps) if s["step_kind"] in {"detected", "warning", "fix_progress"}]
    if fix_indices and (not detected_indices or max(fix_indices) >= max(detected_indices)):
        return "fixed"
    if blocking_indices and (not fix_indices or max(blocking_indices) > max(fix_indices)):
        return "blocked"
    if any(k == "warning" for k in kinds):
        return "warning"
    if fix_indices:
        return "fixed"
    return "observed"


def summarize_evidence(steps: list[dict]) -> str:
    detected = [s for s in steps if s["step_kind"] in {"detected", "warning"}]
    if detected:
        return detected[0]["message"][:1000]
    return steps[0]["message"][:1000]


def summarize_fix(steps: list[dict]) -> str:
    fixes = [s for s in steps if s["step_kind"] == "fix_or_validation"]
    if fixes:
        return fixes[-1]["message"][:1000]
    progress = [s for s in steps if s["step_kind"] == "fix_progress"]
    if progress:
        return progress[-1]["message"][:1000]
    return ""


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--db", type=Path, default=DEFAULT_DB)
    args = ap.parse_args()
    ts = now()
    con = sqlite3.connect(args.db)
    con.row_factory = sqlite3.Row
    con.executescript(SCHEMA)
    con.execute("DELETE FROM repo_issue_steps")
    con.execute("DELETE FROM repo_issue_log")

    grouped: dict[tuple[str, str], list[dict]] = defaultdict(list)

    for e in con.execute("SELECT id, repo, kind, message, url, created_at FROM events WHERE repo IS NOT NULL ORDER BY id"):
        text = f"{e['kind']}: {e['message']}"
        category, title, severity = classify(text)
        # Keep routine merge/release events in release verification; skip totally generic status chatter.
        if category == "general_rollout" and e["kind"] not in {"merge", "release", "validation"}:
            continue
        grouped[(e["repo"], category)].append(
            {
                "source_table": "events",
                "source_id": e["id"],
                "step_time": e["created_at"],
                "step_kind": step_kind(text),
                "command": None,
                "status": e["kind"],
                "message": e["message"],
                "url": e["url"],
                "category": category,
                "title": title,
                "severity": severity,
                "commits": COMMIT_RE.findall(e["message"] or ""),
            }
        )

    for v in con.execute("SELECT id, repo, command, status, note, created_at FROM validations ORDER BY id"):
        text = f"{v['status']}: {v['command']} — {v['note'] or ''}"
        category, title, severity = classify(text)
        grouped[(v["repo"], category)].append(
            {
                "source_table": "validations",
                "source_id": v["id"],
                "step_time": v["created_at"],
                "step_kind": step_kind(text, v["status"]),
                "command": v["command"],
                "status": v["status"],
                "message": text,
                "url": None,
                "category": category,
                "title": title,
                "severity": severity,
                "commits": COMMIT_RE.findall(text),
            }
        )

    issue_count = 0
    step_count = 0
    for (repo, category), steps in sorted(grouped.items()):
        steps.sort(key=lambda s: (s["step_time"] or "", s["source_table"], s["source_id"] or 0))
        title = steps[0]["title"]
        severity = steps[0]["severity"]
        status = final_status(steps)
        detected_at = next((s["step_time"] for s in steps if s["step_kind"] in {"detected", "warning"}), steps[0]["step_time"])
        fixed_at = next((s["step_time"] for s in reversed(steps) if s["step_kind"] == "fix_or_validation"), None)
        commits = []
        for s in steps:
            commits.extend(s["commits"])
        # preserve order while deduping
        commits = list(dict.fromkeys(commits))
        validations = [s for s in steps if s["source_table"] == "validations"]
        validation_summary = "; ".join(f"{s['status']}: {s['command']}" for s in validations[-5:])
        refs = [f"{s['source_table']}:{s['source_id']}" for s in steps if s["source_id"] is not None]
        cur = con.execute(
            """
            INSERT INTO repo_issue_log(repo,issue_key,category,severity,title,status,detected_at,detected_by,evidence_summary,root_cause,fix_summary,fixed_at,fix_commits_json,validation_summary,source_refs_json,created_at,updated_at)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
            """,
            (
                repo,
                category,
                category,
                severity,
                title,
                status,
                detected_at,
                "derived from tracker events/validations",
                summarize_evidence(steps),
                ROOT_CAUSES.get(category, "Derived from rollout tracker events."),
                summarize_fix(steps),
                fixed_at,
                json.dumps(commits),
                validation_summary,
                json.dumps(refs),
                ts,
                ts,
            ),
        )
        issue_id = cur.lastrowid
        issue_count += 1
        for s in steps:
            con.execute(
                """
                INSERT INTO repo_issue_steps(issue_id,repo,step_time,step_kind,source_table,source_id,command,status,message,url,commit_sha,artifact_path,created_at)
                VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
                """,
                (
                    issue_id,
                    repo,
                    s["step_time"],
                    s["step_kind"],
                    s["source_table"],
                    s["source_id"],
                    s["command"],
                    s["status"],
                    s["message"],
                    s["url"],
                    s["commits"][0] if s["commits"] else None,
                    None,
                    ts,
                ),
            )
            step_count += 1

    con.execute(
        "INSERT INTO events(kind,message,created_at) VALUES(?,?,?)",
        ("issue_log_scan", f"populated repo issue log: {issue_count} issues, {step_count} issue steps", ts),
    )
    con.commit()
    con.close()
    print(f"populated issue log in {args.db}")
    print(f"issues={issue_count} steps={step_count}")


if __name__ == "__main__":
    main()
