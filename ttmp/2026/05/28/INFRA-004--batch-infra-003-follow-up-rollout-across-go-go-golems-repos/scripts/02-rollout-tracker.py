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
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from pathlib import Path
from urllib.parse import parse_qs, urlparse

TICKET_DIR = Path(__file__).resolve().parents[1]
DEFAULT_DB = TICKET_DIR / "sources" / "05-rollout-progress.sqlite"
DEFAULT_BATCHES = TICKET_DIR / "sources" / "01-rollout-batches.json"

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
        rows.append(f"<tr><td>{esc(r['batch_id'])}</td><td>{esc(r['repo'])}</td><td><span class='state {state_badge.get(r['state'],'gray')}'>{esc(r['state'])}</span></td><td>{tracks}</td><td>{pr}</td><td><code>{esc((r['merge_sha'] or '')[:10])}</code></td><td>{esc(r['tag'])}</td><td>{esc(r['action_status'])}</td><td>{esc(r['notes'])}</td></tr>")
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
<div class='card'><b>CLI snippets</b><br><code>./scripts/02-rollout-tracker.py list --batch B2</code><br><code>./scripts/02-rollout-tracker.py update-repo REPO --state ready</code></div>
</div>
<table><thead><tr><th>Batch</th><th>Repo</th><th>State</th><th>Tracks</th><th>PR</th><th>Merge SHA</th><th>Tag</th><th>Actions</th><th>Notes</th></tr></thead><tbody>{''.join(rows)}</tbody></table>
<h2>Recent events</h2><table><thead><tr><th>Time</th><th>Repo</th><th>Kind</th><th>Message</th><th>URL</th></tr></thead><tbody>{''.join(f"<tr><td>{esc(e['created_at'])}</td><td>{esc(e['repo'])}</td><td>{esc(e['kind'])}</td><td>{esc(e['message'])}</td><td>{('<a href='+esc(e['url'])+'>link</a>') if e['url'] else ''}</td></tr>" for e in events)}</tbody></table>
</body></html>"""


def dashboard(args: argparse.Namespace) -> None:
    db = args.db
    class Handler(BaseHTTPRequestHandler):
        def do_GET(self):  # noqa: N802
            qs = parse_qs(urlparse(self.path).query)
            batch = qs.get("batch", [None])[0]
            body = html_dashboard(db, batch).encode()
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
