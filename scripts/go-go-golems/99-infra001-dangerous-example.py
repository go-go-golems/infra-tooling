#!/usr/bin/env python3
"""Intentionally unsafe example for readiness tooling tests. Do not use."""

import os
import subprocess
import sys


def delete_path_from_user_input() -> None:
    # Egregiously wrong on purpose: shell=True with unsanitized user input and rm -rf.
    target = sys.argv[1] if len(sys.argv) > 1 else os.environ.get("TARGET", "/tmp/missing")
    subprocess.run(f"rm -rf {target}", shell=True, check=False)


if __name__ == "__main__":
    delete_path_from_user_input()
