#!/usr/bin/env python3
"""Poll a PR until a Codex signal with an eyes/thumbs-up reaction appears."""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from pathlib import Path

THIS_DIR = Path(__file__).resolve().parent
CHECK = THIS_DIR / "01-pr-ready-check.py"


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("pr")
    ap.add_argument("--interval", type=int, default=30)
    ap.add_argument("--timeout", type=int, default=900)
    args = ap.parse_args()
    deadline = time.time() + args.timeout
    while True:
        p = subprocess.run([sys.executable, str(CHECK), args.pr, "--json"], text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if p.stdout:
            data = json.loads(p.stdout)
            print(f"[{time.strftime('%H:%M:%S')}] ready={data['ok']}")
            for f in data["findings"]:
                print(f"  {'OK' if f['ok'] else 'FAIL'}: {f['message']}")
            text = json.dumps(data)
            if "eyes reaction" in text or "thumbs-up reaction" in text:
                # Continue if only 'no eyes reaction' appears with no Codex signal; otherwise callers can read state.
                if "latest Codex signal" in text:
                    return 0
        if time.time() >= deadline:
            print("timed out waiting for Codex reaction signal", file=sys.stderr)
            return 1
        time.sleep(args.interval)


if __name__ == "__main__":
    raise SystemExit(main())
