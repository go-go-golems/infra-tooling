#!/usr/bin/env python3
"""INFRA-004 rollout progress tracker.

Stores rollout state in a small SQLite database and serves a tiny auto-refreshing
HTML dashboard. The dashboard intentionally reads the DB on every request so it
reflects CLI updates without a build step.

Examples:
  # Initialize DB from generated batches
  ./scripts/02-rollout-tracker.py init

  # Show summary / table
  ./scripts/02-rollout-tracker.py summary
  ./scripts/02-rollout-tracker.py list --batch B2

  # Record work
  ./scripts/02-rollout-tracker.py update-repo oak-git-db --state pr_open --branch infra/baseline-rollout --pr-url https://github.com/go-go-golems/oak-git-db/pull/1
  ./scripts/02-rollout-tracker.py validation oak-git-db --command 'GOWORK=off go test ./...' --status pass --note 'no test files'
  ./scripts/02-rollout-tracker.py merge oak-git-db --sha 4f5c6aa0c4d54fbb897bdaef8cea26ab691cbcde

  # Serve dashboard
  ./scripts/02-rollout-tracker.py dashboard --port 8765
"""
from __future__ import annotations

import argparse
import datetime as dt
import html
import json
import sqlite3
import subprocess
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

TICKET_DIR = Path(__file__).resolve().parents[1]
DEFAULT_DB = TICKET_DIR / "sources" / "05-rollout-progress.sqlite"
DEFAULT_BATCHES = TICKET_DIR / "sources" / "01-rollout-batches.json"
INFRA005_DIR = TICKET_DIR.parent.parent / "29" / "INFRA-005--improve-infra-rollout-dashboard-for-dependency-ordered-releases"
DEFAULT_DEP_SCAN_SCRIPT = INFRA005_DIR / "scripts" / "01-populate-internal-dependencies.py"
DEFAULT_ISSUE_SCAN_SCRIPT = INFRA005_DIR / "scripts" / "02-populate-repo-issue-log.py"

STATES = [
    "planned",
    "branch_created",
    "local_validation",
    "pr_open",
    "codex_waiting",
    "codex_feedback",
    "ready",
    "merged",
    "main_actions_verified",
    "released",
    "blocked",
    "skipped",
]

SCHEMA = """
PRAGMA foreign_keys = ON;
CREATE TABLE IF NOT EXISTS repos (
  repo TEXT PRIMARY KEY,
  batch_id TEXT NOT NULL,
  batch_name TEXT NOT NULL,
  module TEXT,
  path TEXT,
  needs_logcopter INTEGER NOT NULL DEFAULT 0,
  needs_docsctl INTEGER NOT NULL DEFAULT 0,
  needs_glazed_lint INTEGER NOT NULL DEFAULT 0,
  needs_xgoja INTEGER NOT NULL DEFAULT 0,
  upstreams TEXT NOT NULL DEFAULT '[]',
  state TEXT NOT NULL DEFAULT 'planned',
  branch TEXT,
  pr_url TEXT,
  pr_number INTEGER,
  head_sha TEXT,
  merge_sha TEXT,
  tag TEXT,
  release_url TEXT,
  docs_url TEXT,
  action_status TEXT,
  notes TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS validations (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo TEXT NOT NULL REFERENCES repos(repo) ON DELETE CASCADE,
  command TEXT NOT NULL,
  status TEXT NOT NULL CHECK(status IN ('pass','fail','skip','warn')),
  note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repo TEXT REFERENCES repos(repo) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  message TEXT NOT NULL,
  url TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_repos_batch ON repos(batch_id, state, repo);
CREATE INDEX IF NOT EXISTS idx_events_repo_created ON events(repo, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_validations_repo_created ON validations(repo, created_at DESC);
"""


def now() -> str:
    return dt.datetime.now(dt.timezone.utc).isoformat(timespec="seconds")


def connect(db: Path) -> sqlite3.Connection:
    db.parent.mkdir(parents=True, exist_ok=True)
    con = sqlite3.connect(db)
    con.row_factory = sqlite3.Row
    con.execute("PRAGMA foreign_keys = ON")
    return con


def init_db(db: Path, batches_path: Path) -> None:
    data = json.loads(batches_path.read_text())
    con = connect(db)
    con.executescript(SCHEMA)
    ts = now()
    count = 0
    for batch in data["batches"]:
        for r in batch["repos"]:
            flags = r["flags"]
            con.execute(
                """
                INSERT INTO repos(repo,batch_id,batch_name,module,path,needs_logcopter,needs_docsctl,needs_glazed_lint,needs_xgoja,upstreams,updated_at)
                VALUES(?,?,?,?,?,?,?,?,?,?,?)
                ON CONFLICT(repo) DO UPDATE SET
                  batch_id=excluded.batch_id,
                  batch_name=excluded.batch_name,
                  module=excluded.module,
                  path=excluded.path,
                  needs_logcopter=excluded.needs_logcopter,
                  needs_docsctl=excluded.needs_docsctl,
                  needs_glazed_lint=excluded.needs_glazed_lint,
                  needs_xgoja=excluded.needs_xgoja,
                  upstreams=excluded.upstreams,
                  updated_at=excluded.updated_at
                """,
                (
                    r["repo"], batch["id"], batch["name"], r.get("module"), r.get("path"),
                    int(flags.get("logcopter", False)), int(flags.get("docsctl", False)),
                    int(flags.get("glazed_lint", False)), int(flags.get("xgoja", False)),
                    json.dumps(r.get("first_party_upstreams", [])), ts,
                ),
            )
            count += 1
    con.execute("INSERT INTO events(kind,message,created_at) VALUES(?,?,?)", ("init", f"initialized/updated {count} repos from {batches_path}", ts))
    con.commit()
    con.close()
    print(f"initialized {db} from {batches_path} ({count} repo rows)")


def ensure_schema(db: Path) -> None:
    con = connect(db)
    con.executescript(SCHEMA)
    con.commit()
    con.close()


def parse_pr_number(url: str | None) -> int | None:
    if not url:
        return None
    try:
        return int(url.rstrip("/").split("/")[-1])
    except Exception:
        return None


def add_event(con: sqlite3.Connection, repo: str | None, kind: str, message: str, url: str | None = None) -> None:
    con.execute("INSERT INTO events(repo,kind,message,url,created_at) VALUES(?,?,?,?,?)", (repo, kind, message, url, now()))


def update_repo(args: argparse.Namespace) -> None:
    con = connect(args.db)
    fields = []
    values = []
    for cli_name, col in [
        ("state", "state"), ("branch", "branch"), ("pr_url", "pr_url"), ("head_sha", "head_sha"),
        ("merge_sha", "merge_sha"), ("tag", "tag"), ("release_url", "release_url"),
        ("docs_url", "docs_url"), ("action_status", "action_status"), ("notes", "notes"),
    ]:
        val = getattr(args, cli_name, None)
        if val is not None:
            fields.append(f"{col}=?")
            values.append(val)
    if args.pr_url is not None:
        fields.append("pr_number=?")
        values.append(parse_pr_number(args.pr_url))
    if not fields:
        raise SystemExit("nothing to update")
    fields.append("updated_at=?")
    values.append(now())
    values.append(args.repo)
    cur = con.execute(f"UPDATE repos SET {', '.join(fields)} WHERE repo=?", values)
    if cur.rowcount == 0:
        raise SystemExit(f"unknown repo: {args.repo}; run init first")
    add_event(con, args.repo, "repo_update", args.event or f"updated repo fields: {', '.join(f.split('=')[0] for f in fields if not f.startswith('updated_at'))}", args.pr_url or args.release_url or args.docs_url)
    con.commit()
    con.close()


def validation(args: argparse.Namespace) -> None:
    con = connect(args.db)
    ts = now()
    con.execute("INSERT INTO validations(repo,command,status,note,created_at) VALUES(?,?,?,?,?)", (args.repo, args.command, args.status, args.note or "", ts))
    current = con.execute("SELECT state FROM repos WHERE repo=?", (args.repo,)).fetchone()
    if current is None:
        raise SystemExit(f"unknown repo: {args.repo}; run init first")
    if args.status == "fail":
        next_state = "blocked"
    elif current["state"] in {"planned", "branch_created"}:
        next_state = "local_validation"
    else:
        next_state = current["state"]
    con.execute("UPDATE repos SET state=?, updated_at=? WHERE repo=?", (next_state, ts, args.repo))
    add_event(con, args.repo, "validation", f"{args.status}: {args.command}" + (f" — {args.note}" if args.note else ""))
    con.commit()
    con.close()


def merge(args: argparse.Namespace) -> None:
    con = connect(args.db)
    ts = now()
    con.execute("UPDATE repos SET state='merged', merge_sha=?, updated_at=? WHERE repo=?", (args.sha, ts, args.repo))
    add_event(con, args.repo, "merge", f"merged with merge commit {args.sha}", args.url)
    con.commit()
    con.close()


def release(args: argparse.Namespace) -> None:
    con = connect(args.db)
    ts = now()
    con.execute("UPDATE repos SET state='released', tag=?, release_url=?, docs_url=?, updated_at=? WHERE repo=?", (args.tag, args.release_url, args.docs_url, ts, args.repo))
    add_event(con, args.repo, "release", f"released {args.tag}", args.release_url)
    con.commit()
    con.close()


def rows_for_list(args: argparse.Namespace) -> list[sqlite3.Row]:
    con = connect(args.db)
    where = []
    vals = []
    if args.batch:
        where.append("batch_id=?")
        vals.append(args.batch)
    if args.state:
        where.append("state=?")
        vals.append(args.state)
    if args.track:
        col = {
            "logcopter": "needs_logcopter",
            "docsctl": "needs_docsctl",
            "glazed": "needs_glazed_lint",
            "xgoja": "needs_xgoja",
        }[args.track]
        where.append(f"{col}=1")
    sql = "SELECT * FROM repos" + (" WHERE " + " AND ".join(where) if where else "") + " ORDER BY batch_id, repo"
    rows = con.execute(sql, vals).fetchall()
    con.close()
    return rows


def list_cmd(args: argparse.Namespace) -> None:
    rows = rows_for_list(args)
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
        return
    print("batch\trepo\tstate\ttracks\tpr\tmerge\ttag\tnotes")
    for r in rows:
        tracks = ",".join([name for name, col in [("logcopter","needs_logcopter"),("docsctl","needs_docsctl"),("glazed","needs_glazed_lint"),("xgoja","needs_xgoja")] if r[col]])
        print("\t".join([r["batch_id"], r["repo"], r["state"], tracks, r["pr_url"] or "", r["merge_sha"] or "", r["tag"] or "", (r["notes"] or "").replace("\n", " ")]))


def summary_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    print("By batch/state:")
    for r in con.execute("SELECT batch_id, state, count(*) c FROM repos GROUP BY batch_id,state ORDER BY batch_id,state"):
        print(f"  {r['batch_id']:>2} {r['state']:<22} {r['c']}")
    print("\nBy track:")
    row = con.execute("SELECT sum(needs_logcopter) logcopter, sum(needs_docsctl) docsctl, sum(needs_glazed_lint) glazed, sum(needs_xgoja) xgoja, count(*) total FROM repos").fetchone()
    print(dict(row))
    print("\nRecent events:")
    for r in con.execute("SELECT * FROM events ORDER BY id DESC LIMIT 10"):
        print(f"  {r['created_at']} {r['repo'] or '-'} {r['kind']}: {r['message']}")
    con.close()



def table_exists(con: sqlite3.Connection, name: str) -> bool:
    return con.execute("SELECT 1 FROM sqlite_master WHERE type='table' AND name=?", (name,)).fetchone() is not None


def parse_json_list(value: str | None) -> list[str]:
    if not value:
        return []
    try:
        data = json.loads(value)
        return [str(x) for x in data]
    except Exception:
        return []


def html_page(db: Path, title: str, body: str) -> str:
    def esc(x): return html.escape(str(x or ""))
    nav = """
    <div class='nav'>
      <a href='/'>Overview</a>
      <a href='/release-train'>Release train</a>
      <a href='/bumps'>Bump queue</a>
    </div>
    """
    return f"""<!doctype html>
<html><head><meta charset='utf-8'><meta http-equiv='refresh' content='20'>
<title>{esc(title)}</title>
<style>
body {{ font-family: system-ui, sans-serif; margin: 24px; background: #fafafa; color: #222; }}
h1 {{ margin-bottom: 0; }} h2 {{ margin-top: 24px; }} .muted {{ color:#666; }}
.nav {{ display:flex; gap:12px; margin:14px 0 18px; }} .nav a {{ background:#fff; border:1px solid #ddd; border-radius:999px; padding:6px 10px; text-decoration:none; color:#174ea6; }}
.grid {{ display:grid; grid-template-columns: repeat(auto-fit,minmax(240px,1fr)); gap:12px; margin:16px 0; }}
.card {{ background:white; border:1px solid #ddd; border-radius:10px; padding:12px; box-shadow:0 1px 2px #0001; }}
table {{ width:100%; border-collapse: collapse; background:white; }} th,td {{ border-bottom:1px solid #eee; padding:7px; text-align:left; vertical-align:top; }} th {{ position:sticky; top:0; background:#f3f3f3; }}
.state,.pill,.badge {{ border-radius:999px; padding:2px 8px; font-size:12px; display:inline-block; margin:1px; }} .pill {{ background:#eef; }}
.green {{ background:#d9f7df; }} .red {{ background:#ffd9d9; }} .orange {{ background:#ffe8c2; }} .blue {{ background:#dcecff; }} .purple {{ background:#eadcff; }} .gray {{ background:#eee; }} .yellow {{ background:#fff6bf; }}
code {{ font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }} pre {{ white-space: pre-wrap; background:#f6f8fa; padding:10px; border-radius:8px; border:1px solid #e5e5e5; }}
.small {{ font-size:12px; }} .ok {{ color:#176b2c; }} .warn {{ color:#9a6700; }} .bad {{ color:#a40e26; }}
</style></head><body>
<h1>{esc(title)}</h1><p class='muted'>DB: <code>{esc(db)}</code>. Auto-refreshes every 20s. Last render: {esc(now())}</p>
{nav}
{body}
</body></html>"""


def badge(text: str, cls: str = "gray") -> str:
    return f"<span class='badge {cls}'>{html.escape(str(text))}</span>"


def html_release_train(db: Path, train: str = "verified_unreleased_tracker_upstreams") -> str:
    con = connect(db)
    if not table_exists(con, "release_order_layers"):
        con.close()
        return html_page(db, "Release train", "<div class='card red'>Dependency tables are not populated yet. Run the INFRA-005 dependency scanner.</div>")
    trains = [r["train_name"] for r in con.execute("SELECT DISTINCT train_name FROM release_order_layers ORDER BY train_name")]
    if train not in trains and trains:
        train = trains[0]
    rows = con.execute(
        """
        SELECT l.*, r.batch_id, r.tag, r.release_url, r.action_status
        FROM release_order_layers l
        LEFT JOIN repos r ON r.repo = l.repo
        WHERE l.train_name=?
        ORDER BY l.layer, l.repo
        """,
        (train,),
    ).fetchall()
    con.close()
    esc = lambda x: html.escape(str(x or ""))
    train_links = " ".join(f"<a href='/release-train?train={esc(t)}'>{esc(t)}</a>" for t in trains)
    trs = []
    for r in rows:
        depends = " ".join(badge(x, "orange") for x in parse_json_list(r["depends_on_json"])) or "—"
        dependents = " ".join(badge(x, "blue") for x in parse_json_list(r["dependents_json"])) or "—"
        action = "Tag/release now" if not parse_json_list(r["depends_on_json"]) and not r["release_url"] else ("Released" if r["release_url"] else "Wait/bump upstreams first")
        cls = "green" if action == "Tag/release now" else ("gray" if action == "Released" else "orange")
        trs.append(
            f"<tr><td>{r['layer']}</td><td><a href='/repo?repo={esc(r['repo'])}'><code>{esc(r['repo'])}</code></a></td><td>{esc(r['batch_id'])}</td><td>{depends}</td><td>{dependents}</td><td>{badge(r['state'], 'green' if r['state']=='main_actions_verified' else 'gray')}</td><td>{esc(r['tag'])}</td><td>{('<a href='+esc(r['release_url'])+'>release</a>') if r['release_url'] else ''}</td><td>{badge(action, cls)}</td><td>{esc(r['rationale'])}</td></tr>"
        )
    body = f"""
    <div class='card'><b>Train:</b> {esc(train)}<br><span class='muted'>Switch:</span> {train_links}</div>
    <table><thead><tr><th>Layer</th><th>Repo</th><th>Batch</th><th>Depends on</th><th>Dependents</th><th>State</th><th>Tag</th><th>Release</th><th>Action</th><th>Rationale</th></tr></thead><tbody>{''.join(trs)}</tbody></table>
    """
    return html_page(db, "Dependency-ordered release train", body)


def html_bumps(db: Path) -> str:
    con = connect(db)
    if not table_exists(con, "dependency_bump_candidates"):
        con.close()
        return html_page(db, "Bump queue", "<div class='card red'>Dependency bump table is not populated yet. Run the INFRA-005 dependency scanner.</div>")
    rows = con.execute(
        """
        SELECT * FROM dependency_bump_candidates
        ORDER BY CASE priority WHEN 'high' THEN 0 ELSE 1 END, dependency_repo, repo
        """
    ).fetchall()
    con.close()
    esc = lambda x: html.escape(str(x or ""))
    grouped: dict[str, list[sqlite3.Row]] = {}
    for r in rows:
        grouped.setdefault(r["dependency_repo"], []).append(r)
    sections = []
    for dep, items in grouped.items():
        trs = []
        for r in items:
            cmd = f"GOWORK=off go get {r['dependency_module']}@{r['available_tag']} && GOWORK=off go mod tidy"
            trs.append(f"<tr><td><a href='/repo?repo={esc(r['repo'])}'><code>{esc(r['repo'])}</code></a></td><td><code>{esc(r['current_required_version'])}</code></td><td><code>{esc(r['available_tag'])}</code></td><td>{badge(r['priority'], 'red' if r['priority']=='high' else 'gray')}</td><td>{'direct' if r['direct'] else 'indirect'}</td><td>{esc(r['reason'])}<pre>{esc(cmd)}</pre></td></tr>")
        sections.append(f"<h2><a href='/repo?repo={esc(dep)}'><code>{esc(dep)}</code></a></h2><table><thead><tr><th>Consumer</th><th>Current</th><th>Available</th><th>Priority</th><th>Edge</th><th>Reason / command</th></tr></thead><tbody>{''.join(trs)}</tbody></table>")
    return html_page(db, "Dependency bump queue", "".join(sections) if sections else "<div class='card green'>No bump candidates.</div>")


def html_repo_detail(db: Path, repo: str | None) -> str:
    esc = lambda x: html.escape(str(x or ""))
    if not repo:
        return html_page(db, "Repository detail", "<div class='card red'>Missing ?repo=...</div>")
    con = connect(db)
    r = con.execute("SELECT * FROM repos WHERE repo=?", (repo,)).fetchone()
    if r is None:
        con.close()
        return html_page(db, "Repository detail", f"<div class='card red'>Unknown repo: {esc(repo)}</div>")
    deps = con.execute("SELECT * FROM internal_dependency_edges WHERE repo=? ORDER BY indirect, dependency_repo", (repo,)).fetchall() if table_exists(con, "internal_dependency_edges") else []
    dependents = con.execute("SELECT * FROM internal_dependency_edges WHERE dependency_repo=? ORDER BY indirect, repo", (repo,)).fetchall() if table_exists(con, "internal_dependency_edges") else []
    bumps = con.execute("SELECT * FROM dependency_bump_candidates WHERE repo=? OR dependency_repo=? ORDER BY dependency_repo, repo", (repo, repo)).fetchall() if table_exists(con, "dependency_bump_candidates") else []
    issues = con.execute("SELECT * FROM repo_issue_log WHERE repo=? ORDER BY CASE status WHEN 'blocked' THEN 0 WHEN 'observed' THEN 1 WHEN 'warning' THEN 2 ELSE 3 END, category", (repo,)).fetchall() if table_exists(con, "repo_issue_log") else []
    issue_steps = {}
    if issues and table_exists(con, "repo_issue_steps"):
        for issue in issues:
            issue_steps[issue["id"]] = con.execute("SELECT * FROM repo_issue_steps WHERE issue_id=? ORDER BY step_time, id", (issue["id"],)).fetchall()
    validations = con.execute("SELECT * FROM validations WHERE repo=? ORDER BY id DESC LIMIT 20", (repo,)).fetchall()
    events = con.execute("SELECT * FROM events WHERE repo=? ORDER BY id DESC LIMIT 30", (repo,)).fetchall()
    con.close()
    tracks = " ".join(badge(name, "blue") for name, col in [("logcopter","needs_logcopter"),("docsctl","needs_docsctl"),("glazed","needs_glazed_lint"),("xgoja","needs_xgoja")] if r[col]) or "—"
    summary = f"""
    <div class='grid'>
      <div class='card'><b>Repo</b><br><code>{esc(repo)}</code><br>{badge(r['state'], 'green' if r['state'] in {'released','main_actions_verified'} else 'orange')}</div>
      <div class='card'><b>Tracks</b><br>{tracks}</div>
      <div class='card'><b>PR / release</b><br>{('<a href='+esc(r['pr_url'])+'>PR '+esc(r['pr_number'])+'</a>') if r['pr_url'] else '—'}<br>{('<a href='+esc(r['release_url'])+'>release</a>') if r['release_url'] else ''}</div>
      <div class='card'><b>SHAs</b><br>head <code>{esc((r['head_sha'] or '')[:10])}</code><br>merge <code>{esc((r['merge_sha'] or '')[:10])}</code></div>
    </div>
    """
    dep_rows = "".join(f"<tr><td><a href='/repo?repo={esc(d['dependency_repo'])}'><code>{esc(d['dependency_repo'])}</code></a></td><td>{esc(d['required_version'])}</td><td>{'indirect' if d['indirect'] else 'direct'}</td><td>{esc(d['replace_target'])}</td><td>{esc(d['dependency_state'])}</td><td>{esc(d['dependency_latest_local_tag'])}</td></tr>" for d in deps)
    dependent_rows = "".join(f"<tr><td><a href='/repo?repo={esc(d['repo'])}'><code>{esc(d['repo'])}</code></a></td><td>{esc(d['required_version'])}</td><td>{'indirect' if d['indirect'] else 'direct'}</td><td>{esc(d['replace_target'])}</td></tr>" for d in dependents)
    issue_cards = []
    for issue in issues:
        steps = issue_steps.get(issue["id"], [])
        step_items = "".join(f"<li><span class='muted'>{esc(s['step_time'])}</span> {badge(s['step_kind'], 'green' if s['step_kind']=='fix_or_validation' else 'orange' if s['step_kind'] in {'detected','warning','fix_progress'} else 'gray')} {esc(s['message'])}</li>" for s in steps[:12])
        more = f"<li class='muted'>... {len(steps)-12} more steps</li>" if len(steps) > 12 else ""
        issue_cards.append(f"<div class='card'><h3>{esc(issue['title'])} {badge(issue['status'], 'green' if issue['status']=='fixed' else 'red' if issue['status']=='blocked' else 'orange')}</h3><p><b>Category:</b> <code>{esc(issue['category'])}</code> <b>Severity:</b> {esc(issue['severity'])}</p><p><b>Evidence:</b> {esc(issue['evidence_summary'])}</p><p><b>Root cause:</b> {esc(issue['root_cause'])}</p><p><b>Fix:</b> {esc(issue['fix_summary'])}</p><p><b>Validation:</b> {esc(issue['validation_summary'])}</p><ol>{step_items}{more}</ol></div>")
    bump_rows = "".join(f"<tr><td>{esc(b['dependency_repo'])}</td><td>{esc(b['repo'])}</td><td>{esc(b['current_required_version'])}</td><td>{esc(b['available_tag'])}</td><td>{esc(b['priority'])}</td><td>{esc(b['reason'])}</td></tr>" for b in bumps)
    val_rows = "".join(f"<tr><td>{esc(v['created_at'])}</td><td>{esc(v['status'])}</td><td><code>{esc(v['command'])}</code></td><td>{esc(v['note'])}</td></tr>" for v in validations)
    event_rows = "".join(f"<tr><td>{esc(e['created_at'])}</td><td>{esc(e['kind'])}</td><td>{esc(e['message'])}</td><td>{('<a href='+esc(e['url'])+'>link</a>') if e['url'] else ''}</td></tr>" for e in events)
    body = summary + f"""
    <h2>Issue / fix history</h2>{''.join(issue_cards) if issue_cards else '<div class="card gray">No derived issue history. Run issue refresh.</div>'}
    <h2>Internal dependencies</h2><table><thead><tr><th>Dependency</th><th>Required</th><th>Edge</th><th>Replace</th><th>State</th><th>Latest tag</th></tr></thead><tbody>{dep_rows}</tbody></table>
    <h2>Internal dependents</h2><table><thead><tr><th>Consumer</th><th>Required</th><th>Edge</th><th>Replace</th></tr></thead><tbody>{dependent_rows}</tbody></table>
    <h2>Bump candidates touching this repo</h2><table><thead><tr><th>Dependency</th><th>Consumer</th><th>Current</th><th>Available</th><th>Priority</th><th>Reason</th></tr></thead><tbody>{bump_rows}</tbody></table>
    <h2>Recent validations</h2><table><thead><tr><th>Time</th><th>Status</th><th>Command</th><th>Note</th></tr></thead><tbody>{val_rows}</tbody></table>
    <h2>Recent raw events</h2><table><thead><tr><th>Time</th><th>Kind</th><th>Message</th><th>URL</th></tr></thead><tbody>{event_rows}</tbody></table>
    """
    return html_page(db, f"Repository detail: {repo}", body)


def deps_modules_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    if not table_exists(con, "internal_modules"):
        raise SystemExit("internal_modules table missing; run deps-scan first")
    rows = con.execute("SELECT * FROM internal_modules ORDER BY repo").fetchall()
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
    else:
        print("repo\tmodule\ttracker_state\tlatest_local_tag\tin_tracker")
        for r in rows:
            print(f"{r['repo']}\t{r['module']}\t{r['tracker_state'] or ''}\t{r['latest_local_tag'] or ''}\t{r['in_tracker']}")
    con.close()


def deps_edges_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    if not table_exists(con, "internal_dependency_edges"):
        raise SystemExit("internal_dependency_edges table missing; run deps-scan first")
    where, vals = [], []
    if args.repo:
        where.append("repo=?"); vals.append(args.repo)
    if args.dependency:
        where.append("dependency_repo=?"); vals.append(args.dependency)
    sql = "SELECT * FROM internal_dependency_edges" + (" WHERE " + " AND ".join(where) if where else "") + " ORDER BY repo, indirect, dependency_repo"
    rows = con.execute(sql, vals).fetchall()
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
    else:
        print("repo\tdependency\trequired\tedge\tdep_state\tlatest_tag\treplace")
        for r in rows:
            print(f"{r['repo']}\t{r['dependency_repo']}\t{r['required_version'] or ''}\t{'indirect' if r['indirect'] else 'direct'}\t{r['dependency_state'] or ''}\t{r['dependency_latest_local_tag'] or ''}\t{r['replace_target'] or ''}")
    con.close()


def deps_release_order_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    if not table_exists(con, "release_order_layers"):
        raise SystemExit("release_order_layers table missing; run deps-scan first")
    rows = con.execute("SELECT * FROM release_order_layers WHERE train_name=? ORDER BY layer, repo", (args.train,)).fetchall()
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
    else:
        print("layer\trepo\tdepends_on\tdependents\trationale")
        for r in rows:
            print(f"{r['layer']}\t{r['repo']}\t{','.join(parse_json_list(r['depends_on_json']))}\t{','.join(parse_json_list(r['dependents_json']))}\t{r['rationale']}")
    con.close()


def deps_bumps_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    if not table_exists(con, "dependency_bump_candidates"):
        raise SystemExit("dependency_bump_candidates table missing; run deps-scan first")
    where, vals = [], []
    if args.repo:
        where.append("repo=?"); vals.append(args.repo)
    if args.dependency:
        where.append("dependency_repo=?"); vals.append(args.dependency)
    sql = "SELECT * FROM dependency_bump_candidates" + (" WHERE " + " AND ".join(where) if where else "") + " ORDER BY CASE priority WHEN 'high' THEN 0 ELSE 1 END, dependency_repo, repo"
    rows = con.execute(sql, vals).fetchall()
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
    else:
        print("repo\tdependency\tcurrent\tavailable\tpriority\treason")
        for r in rows:
            print(f"{r['repo']}\t{r['dependency_repo']}\t{r['current_required_version'] or ''}\t{r['available_tag'] or ''}\t{r['priority']}\t{r['reason']}")
    con.close()


def issue_list_cmd(args: argparse.Namespace) -> None:
    con = connect(args.db)
    if not table_exists(con, "repo_issue_log"):
        raise SystemExit("repo_issue_log table missing; run issue-refresh first")
    where, vals = [], []
    if args.repo:
        where.append("repo=?"); vals.append(args.repo)
    if args.category:
        where.append("category=?"); vals.append(args.category)
    if args.status:
        where.append("status=?"); vals.append(args.status)
    sql = "SELECT * FROM repo_issue_log" + (" WHERE " + " AND ".join(where) if where else "") + " ORDER BY repo, category"
    rows = con.execute(sql, vals).fetchall()
    if args.json:
        print(json.dumps([dict(r) for r in rows], indent=2))
    else:
        print("repo\tcategory\tstatus\ttitle\tevidence\tfix")
        for r in rows:
            print(f"{r['repo']}\t{r['category']}\t{r['status']}\t{r['title']}\t{(r['evidence_summary'] or '')[:100]}\t{(r['fix_summary'] or '')[:100]}")
    con.close()


def run_refresh_script(script: Path, db: Path) -> None:
    if not script.exists():
        raise SystemExit(f"refresh script not found: {script}")
    subprocess.check_call([sys.executable, str(script), "--db", str(db)])

def html_dashboard(db: Path, batch: str | None = None) -> str:
    con = connect(db)
    where = "WHERE batch_id=?" if batch else ""
    vals = [batch] if batch else []
    repos = con.execute(f"SELECT * FROM repos {where} ORDER BY batch_id, CASE state WHEN 'blocked' THEN 0 WHEN 'codex_feedback' THEN 1 WHEN 'pr_open' THEN 2 ELSE 3 END, repo", vals).fetchall()
    states = con.execute("SELECT state, count(*) c FROM repos GROUP BY state ORDER BY state").fetchall()
    batches = con.execute("SELECT batch_id, batch_name, count(*) c FROM repos GROUP BY batch_id,batch_name ORDER BY batch_id").fetchall()
    events = con.execute("SELECT * FROM events ORDER BY id DESC LIMIT 25").fetchall()
    con.close()
    state_badge = {
        "planned": "gray", "branch_created": "blue", "local_validation": "blue", "pr_open": "purple",
        "codex_waiting": "orange", "codex_feedback": "red", "ready": "green", "merged": "green",
        "main_actions_verified": "green", "released": "green", "blocked": "red", "skipped": "gray",
    }
    def esc(x): return html.escape(str(x or ""))
    rows = []
    for r in repos:
        tracks = " ".join([f"<span class='pill'>{name}</span>" for name, col in [("log","needs_logcopter"),("docs","needs_docsctl"),("glazed","needs_glazed_lint"),("xgoja","needs_xgoja")] if r[col]])
        pr = f"<a href='{esc(r['pr_url'])}'>PR {esc(r['pr_number'])}</a>" if r["pr_url"] else ""
        rows.append(f"<tr><td>{esc(r['batch_id'])}</td><td><a href='/repo?repo={esc(r['repo'])}'><code>{esc(r['repo'])}</code></a></td><td><span class='state {state_badge.get(r['state'],'gray')}'>{esc(r['state'])}</span></td><td>{tracks}</td><td>{pr}</td><td><code>{esc((r['merge_sha'] or '')[:10])}</code></td><td>{esc(r['tag'])}</td><td>{esc(r['action_status'])}</td><td>{esc(r['notes'])}</td></tr>")
    return f"""<!doctype html>
<html><head><meta charset='utf-8'><meta http-equiv='refresh' content='10'>
<title>INFRA-004 rollout</title>
<style>
body {{ font-family: system-ui, sans-serif; margin: 24px; background: #fafafa; color: #222; }}
h1 {{ margin-bottom: 0; }} .muted {{ color:#666; }}
.grid {{ display:grid; grid-template-columns: repeat(auto-fit,minmax(220px,1fr)); gap:12px; margin:16px 0; }}
.card {{ background:white; border:1px solid #ddd; border-radius:10px; padding:12px; box-shadow:0 1px 2px #0001; }}
table {{ width:100%; border-collapse: collapse; background:white; }} th,td {{ border-bottom:1px solid #eee; padding:7px; text-align:left; vertical-align:top; }} th {{ position:sticky; top:0; background:#f3f3f3; }}
.state,.pill {{ border-radius:999px; padding:2px 8px; font-size:12px; display:inline-block; }} .pill {{ background:#eef; margin:1px; }}
.green {{ background:#d9f7df; }} .red {{ background:#ffd9d9; }} .orange {{ background:#ffe8c2; }} .blue {{ background:#dcecff; }} .purple {{ background:#eadcff; }} .gray {{ background:#eee; }}
code {{ font-family: ui-monospace, SFMono-Regular, Menlo, monospace; }}
</style></head><body>
<h1>INFRA-004 rollout dashboard</h1><p class='muted'>DB: <code>{esc(db)}</code>. Auto-refreshes every 10s. Last render: {esc(now())}</p>
<div class='grid'>
<div class='card'><b>States</b><br>{''.join(f"<div>{esc(s['state'])}: <b>{s['c']}</b></div>" for s in states)}</div>
<div class='card'><b>Batches</b><br>{''.join(f"<div><a href='/?batch={esc(b['batch_id'])}'>{esc(b['batch_id'])}</a>: {esc(b['batch_name'])} <b>{b['c']}</b></div>" for b in batches)}<div><a href='/'>all</a></div></div>
<div class='card'><b>Views</b><br><a href='/release-train'>Release train</a><br><a href='/bumps'>Bump queue</a><br><code>./scripts/02-rollout-tracker.py deps-release-order</code></div>
</div>
<table><thead><tr><th>Batch</th><th>Repo</th><th>State</th><th>Tracks</th><th>PR</th><th>Merge SHA</th><th>Tag</th><th>Actions</th><th>Notes</th></tr></thead><tbody>{''.join(rows)}</tbody></table>
<h2>Recent events</h2><table><thead><tr><th>Time</th><th>Repo</th><th>Kind</th><th>Message</th><th>URL</th></tr></thead><tbody>{''.join(f"<tr><td>{esc(e['created_at'])}</td><td>{esc(e['repo'])}</td><td>{esc(e['kind'])}</td><td>{esc(e['message'])}</td><td>{('<a href='+esc(e['url'])+'>link</a>') if e['url'] else ''}</td></tr>" for e in events)}</tbody></table>
</body></html>"""


def dashboard(args: argparse.Namespace) -> None:
    db = args.db
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self):  # noqa: N802
            parsed = urlparse(self.path)
            qs = parse_qs(parsed.query)
            if parsed.path == "/release-train":
                body_text = html_release_train(db, qs.get("train", ["verified_unreleased_tracker_upstreams"])[0])
            elif parsed.path == "/bumps":
                body_text = html_bumps(db)
            elif parsed.path == "/repo":
                body_text = html_repo_detail(db, qs.get("repo", [None])[0])
            else:
                batch = qs.get("batch", [None])[0]
                body_text = html_dashboard(db, batch)
            body = body_text.encode()
            self.send_response(200)
            self.send_header("Content-Type", "text/html; charset=utf-8")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        def log_message(self, fmt, *a):
            sys.stderr.write("dashboard: " + fmt % a + "\n")
    print(f"serving http://127.0.0.1:{args.port}/ from {db}")
    ThreadingHTTPServer(("127.0.0.1", args.port), Handler).serve_forever()


def main() -> None:
    p = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    p.add_argument("--db", type=Path, default=DEFAULT_DB)
    sub = p.add_subparsers(dest="cmd", required=True)

    s = sub.add_parser("init")
    s.add_argument("--batches", type=Path, default=DEFAULT_BATCHES)
    s.set_defaults(func=lambda a: init_db(a.db, a.batches))

    s = sub.add_parser("summary")
    s.set_defaults(func=summary_cmd)

    s = sub.add_parser("list")
    s.add_argument("--batch")
    s.add_argument("--state", choices=STATES)
    s.add_argument("--track", choices=["logcopter", "docsctl", "glazed", "xgoja"])
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=list_cmd)

    s = sub.add_parser("update-repo")
    s.add_argument("repo")
    s.add_argument("--state", choices=STATES)
    s.add_argument("--branch")
    s.add_argument("--pr-url")
    s.add_argument("--head-sha")
    s.add_argument("--merge-sha")
    s.add_argument("--tag")
    s.add_argument("--release-url")
    s.add_argument("--docs-url")
    s.add_argument("--action-status")
    s.add_argument("--notes")
    s.add_argument("--event")
    s.set_defaults(func=update_repo)

    s = sub.add_parser("validation")
    s.add_argument("repo")
    s.add_argument("--command", required=True)
    s.add_argument("--status", choices=["pass", "fail", "skip", "warn"], required=True)
    s.add_argument("--note")
    s.set_defaults(func=validation)

    s = sub.add_parser("merge")
    s.add_argument("repo")
    s.add_argument("--sha", required=True)
    s.add_argument("--url")
    s.set_defaults(func=merge)

    s = sub.add_parser("release")
    s.add_argument("repo")
    s.add_argument("--tag", required=True)
    s.add_argument("--release-url")
    s.add_argument("--docs-url")
    s.set_defaults(func=release)

    s = sub.add_parser("event")
    s.add_argument("--repo")
    s.add_argument("--kind", required=True)
    s.add_argument("--message", required=True)
    s.add_argument("--url")
    s.set_defaults(func=lambda a: (lambda con: (add_event(con, a.repo, a.kind, a.message, a.url), con.commit(), con.close()))(connect(a.db)))

    s = sub.add_parser("deps-scan")
    s.add_argument("--script", type=Path, default=DEFAULT_DEP_SCAN_SCRIPT)
    s.set_defaults(func=lambda a: run_refresh_script(a.script, a.db))

    s = sub.add_parser("issue-refresh")
    s.add_argument("--script", type=Path, default=DEFAULT_ISSUE_SCAN_SCRIPT)
    s.set_defaults(func=lambda a: run_refresh_script(a.script, a.db))

    s = sub.add_parser("deps-modules")
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=deps_modules_cmd)

    s = sub.add_parser("deps-edges")
    s.add_argument("--repo")
    s.add_argument("--dependency")
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=deps_edges_cmd)

    s = sub.add_parser("deps-release-order")
    s.add_argument("--train", default="verified_unreleased_tracker_upstreams")
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=deps_release_order_cmd)

    s = sub.add_parser("deps-bumps")
    s.add_argument("--repo")
    s.add_argument("--dependency")
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=deps_bumps_cmd)

    s = sub.add_parser("issue-list")
    s.add_argument("--repo")
    s.add_argument("--category")
    s.add_argument("--status")
    s.add_argument("--json", action="store_true")
    s.set_defaults(func=issue_list_cmd)

    s = sub.add_parser("dashboard")
    s.add_argument("--port", type=int, default=8765)
    s.set_defaults(func=dashboard)

    args = p.parse_args()
    if args.cmd != "init" and not args.db.exists():
        init_db(args.db, DEFAULT_BATCHES)
    else:
        ensure_schema(args.db)
    args.func(args)

if __name__ == "__main__":
    main()
