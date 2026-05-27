#!/usr/bin/env python3
"""Check whether a GitHub PR is ready based on CI and Codex review signals.

Criteria implemented for go-go-golems rollout PRs:
  1. GitHub Actions/status checks have run and every check is completed with an
     acceptable conclusion.
  2. A Codex signal exists. This can be a Codex-authored review/comment, or
     the newest human '@codex review' trigger comment that Codex reacts to.
  3. The newest Codex signal has a thumbs-up reaction and no eyes reaction.
  4. Codex-authored bodies do not contain substantive review comments.

The script intentionally exits non-zero when a PR is not ready so it can be used
in batch automation.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from dataclasses import dataclass
from typing import Any

ACCEPTABLE_CHECK_CONCLUSIONS = {"SUCCESS", "SKIPPED", "NEUTRAL"}
DEFAULT_CODEX_RE = r"(?i)(^|[-_])(codex|openai-codex|chatgpt)([-_]|$)|codex|openai"
BENIGN_CODEX_BODY_RE = re.compile(
    r"^\s*(?:approved|looks good|lgtm|no issues found|✅|👍|:\+1:|:thumbsup:|thumbs up|nit:)?\s*$",
    re.IGNORECASE,
)
SATISFIED_CODEX_BODY_RE = re.compile(
    r"(?is)(didn'?t find (?:any )?major issues|no major issues|looks good|lgtm).*(?:👍|:\+1:|:thumbsup:|thumbs up)",
)

QUERY = r"""
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      url
      number
      title
      mergeStateStatus
      reviewDecision
      headRefOid
      statusCheckRollup {
        contexts(first: 100) {
          nodes {
            __typename
            ... on CheckRun {
              name
              status
              conclusion
              detailsUrl
            }
            ... on StatusContext {
              context
              state
              targetUrl
            }
          }
        }
      }
      reviews(last: 100) {
        nodes {
          databaseId
          author { login }
          state
          body
          submittedAt
          url
          reactionGroups {
            content
            users(first: 20) { totalCount nodes { login } }
          }
        }
      }
      comments(last: 100) {
        nodes {
          databaseId
          author { login }
          body
          createdAt
          url
          reactionGroups {
            content
            users(first: 20) { totalCount nodes { login } }
          }
        }
      }
    }
  }
}
"""

@dataclass
class Finding:
    ok: bool
    message: str


def run_gh_json(args: list[str]) -> Any:
    p = subprocess.run(args, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if p.returncode != 0:
        print(p.stderr, file=sys.stderr)
        raise SystemExit(p.returncode)
    return json.loads(p.stdout)


def parse_pr(pr: str) -> tuple[str, str, int]:
    if pr.startswith("http://") or pr.startswith("https://"):
        m = re.search(r"github\.com/([^/]+)/([^/]+)/pull/(\d+)", pr)
        if not m:
            raise SystemExit(f"could not parse PR URL: {pr}")
        return m.group(1), m.group(2), int(m.group(3))
    if "#" in pr:
        repo, num = pr.split("#", 1)
        owner, name = repo.split("/", 1)
        return owner, name, int(num)
    raise SystemExit("PR must be a GitHub PR URL or owner/repo#number")


def reaction_count(node: dict[str, Any], content: str) -> int:
    for group in node.get("reactionGroups") or []:
        if group.get("content") == content:
            return int(((group.get("users") or {}).get("totalCount")) or 0)
    return 0


def codex_body_is_satisfied(body: str) -> bool:
    return bool(SATISFIED_CODEX_BODY_RE.search(body or ""))


def codex_body_is_benign(body: str) -> bool:
    stripped = (body or "").strip()
    return not stripped or bool(BENIGN_CODEX_BODY_RE.match(stripped)) or codex_body_is_satisfied(stripped)


def checks_findings(pr: dict[str, Any]) -> list[Finding]:
    nodes = (((pr.get("statusCheckRollup") or {}).get("contexts") or {}).get("nodes")) or []
    findings: list[Finding] = []
    if not nodes:
        return [Finding(False, "no status checks found; actions may not have run")]
    bad = []
    pending = []
    for n in nodes:
        typ = n.get("__typename")
        if typ == "CheckRun":
            name = n.get("name") or "<unnamed check>"
            status = n.get("status")
            conclusion = n.get("conclusion")
            if status != "COMPLETED":
                pending.append(f"{name}: status={status}")
            elif conclusion not in ACCEPTABLE_CHECK_CONCLUSIONS:
                bad.append(f"{name}: conclusion={conclusion}")
        elif typ == "StatusContext":
            name = n.get("context") or "<unnamed status>"
            state = n.get("state")
            if state != "SUCCESS":
                bad.append(f"{name}: state={state}")
    if pending:
        findings.append(Finding(False, "pending checks: " + "; ".join(pending)))
    if bad:
        findings.append(Finding(False, "failing/non-success checks: " + "; ".join(bad)))
    if not pending and not bad:
        findings.append(Finding(True, f"all {len(nodes)} status checks completed successfully"))
    return findings


def collect_codex_signals(pr: dict[str, Any], codex_re: re.Pattern[str]) -> list[dict[str, Any]]:
    signals: list[dict[str, Any]] = []
    for kind, connection, time_key in (
        ("review", "reviews", "submittedAt"),
        ("comment", "comments", "createdAt"),
    ):
        for n in ((pr.get(connection) or {}).get("nodes") or []):
            login = ((n.get("author") or {}).get("login")) or ""
            body = n.get("body") or ""
            is_codex_authored = bool(codex_re.search(login))
            is_codex_trigger = kind == "comment" and bool(re.search(r"(?im)^\s*@codex\s+review\s*$", body))
            if is_codex_authored or is_codex_trigger:
                nn = dict(n)
                nn["kind"] = "codex-trigger" if is_codex_trigger and not is_codex_authored else kind
                nn["login"] = login
                nn["time"] = n.get(time_key) or ""
                nn["codexAuthored"] = is_codex_authored
                signals.append(nn)
    signals.sort(key=lambda n: n.get("time") or "")
    return signals


def codex_findings(pr: dict[str, Any], codex_re: re.Pattern[str]) -> list[Finding]:
    signals = collect_codex_signals(pr, codex_re)
    if not signals:
        return [Finding(False, "no Codex-authored review/comment signal found")]
    latest = signals[-1]
    authored_signals = [s for s in signals if s.get("codexAuthored")]
    latest_authored = authored_signals[-1] if authored_signals else None

    thumbs = reaction_count(latest, "THUMBS_UP")
    eyes = reaction_count(latest, "EYES")
    body = latest.get("body") or ""
    where = f"latest Codex signal ({latest['kind']}) by {latest.get('login') or '<unknown>'}: {latest.get('url')}"
    findings: list[Finding] = [Finding(True, where)]

    # A newer human "@codex review" trigger is not enough to hide substantive
    # Codex review comments that are already present. Release-train automation
    # should stop on those comments so an operator can either fix them or decide
    # to wait for the retriggered review intentionally.
    if latest_authored is not None:
        authored_body = latest_authored.get("body") or ""
        authored_where = f"latest Codex-authored signal ({latest_authored['kind']}) by {latest_authored.get('login') or '<unknown>'}: {latest_authored.get('url')}"
        findings.append(Finding(True, authored_where))
        if not codex_body_is_benign(authored_body):
            preview = re.sub(r"\s+", " ", authored_body.strip())[:240]
            findings.append(Finding(False, f"latest Codex-authored body contains substantive comments: {preview!r}"))
        else:
            findings.append(Finding(True, "latest Codex-authored body is empty/benign/satisfied"))

    body_satisfied = latest.get("codexAuthored") and codex_body_is_satisfied(body)
    if thumbs <= 0 and not body_satisfied:
        findings.append(Finding(False, "latest Codex signal has no thumbs-up reaction or satisfied thumbs-up body"))
    elif thumbs > 0:
        findings.append(Finding(True, f"latest Codex signal has {thumbs} thumbs-up reaction(s)"))
    else:
        findings.append(Finding(True, "latest Codex-authored body contains a satisfied thumbs-up signal"))
    if eyes > 0:
        findings.append(Finding(False, f"latest Codex signal has {eyes} eyes reaction(s), review may still be running"))
    else:
        findings.append(Finding(True, "latest Codex signal has no eyes reaction"))
    if not latest.get("codexAuthored"):
        findings.append(Finding(True, "latest signal is a human @codex review trigger; trigger does not mask existing Codex-authored findings"))
    return findings


def classify_findings(findings: list[Finding]) -> tuple[str, bool, list[str]]:
    if all(f.ok for f in findings):
        return "ready", True, []

    failed = [f.message for f in findings if not f.ok]
    failed_check_kinds: list[str] = []
    for msg in failed:
        if "pending checks:" in msg:
            failed_check_kinds.append("pending_checks")
        if "failing/non-success checks:" in msg:
            details = msg.split(":", 1)[1] if ":" in msg else msg
            lower = details.lower()
            if "test" in lower:
                failed_check_kinds.append("test")
            if "lint" in lower:
                failed_check_kinds.append("lint")
            if "vulnerability" in lower or "govuln" in lower:
                failed_check_kinds.append("govulncheck")
            if "gosec" in lower or "security scan" in lower:
                failed_check_kinds.append("gosec")
            if "dependency review" in lower:
                failed_check_kinds.append("dependency_review")
            failed_check_kinds.append("checks")

    failed_check_kinds = sorted(set(failed_check_kinds))

    if any("latest Codex-authored body contains substantive comments" in msg for msg in failed):
        return "codex_feedback", True, failed_check_kinds
    if any("failing/non-success checks:" in msg for msg in failed):
        return "failed_checks", True, failed_check_kinds
    if any("pending checks:" in msg for msg in failed):
        return "waiting_checks", False, failed_check_kinds
    if any("no Codex-authored review/comment signal found" in msg for msg in failed):
        return "no_codex", False, failed_check_kinds
    if any("eyes reaction" in msg or "no thumbs-up reaction" in msg for msg in failed):
        return "waiting_codex", False, failed_check_kinds
    return "not_ready", False, failed_check_kinds


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("pr", help="PR URL or owner/repo#number")
    ap.add_argument("--codex-author-regex", default=DEFAULT_CODEX_RE)
    ap.add_argument("--json", action="store_true", help="emit machine-readable JSON")
    args = ap.parse_args()

    owner, repo, number = parse_pr(args.pr)
    data = run_gh_json([
        "gh", "api", "graphql",
        "-f", f"owner={owner}",
        "-f", f"repo={repo}",
        "-F", f"number={number}",
        "-f", f"query={QUERY}",
    ])
    pr = data["data"]["repository"]["pullRequest"]
    codex_re = re.compile(args.codex_author_regex)
    findings = checks_findings(pr) + codex_findings(pr, codex_re)
    state, terminal, failed_check_kinds = classify_findings(findings)
    ok = state == "ready"

    if args.json:
        print(json.dumps({
            "ok": ok,
            "state": state,
            "terminal": terminal,
            "failedCheckKinds": failed_check_kinds,
            "url": pr.get("url"),
            "mergeStateStatus": pr.get("mergeStateStatus"),
            "reviewDecision": pr.get("reviewDecision"),
            "findings": [f.__dict__ for f in findings],
        }, indent=2))
    else:
        print(f"PR: {pr.get('url')}")
        print(f"READY: {'yes' if ok else 'no'}")
        print(f"STATE: {state}")
        print(f"TERMINAL: {'yes' if terminal else 'no'}")
        if failed_check_kinds:
            print(f"FAILED_CHECK_KINDS: {', '.join(failed_check_kinds)}")
        for f in findings:
            print(f"{'OK' if f.ok else 'FAIL'}: {f.message}")
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
