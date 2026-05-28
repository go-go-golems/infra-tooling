#!/usr/bin/env bash
set -euo pipefail

# Inventory generated logcopter package-logger adoption for a set of Go repos.
# Usage:
#   11-inventory-logcopter-adoption.sh /path/to/repo [...]
#
# Prints tab-separated rows:
#   repo path dep generate_file generated_count make_check make_bump zerolog_log_imports branch_status

if [ "$#" -eq 0 ]; then
  cat >&2 <<'USAGE'
Usage: 11-inventory-logcopter-adoption.sh /path/to/repo [...]

Pass repository roots explicitly. This script is intentionally path-based so it
can work with temporary clean clones as well as normal worktrees.
USAGE
  exit 2
fi

printf 'repo\tpath\tdep\tgenerate_file\tgenerated_count\tmake_check\tmake_bump\tzerolog_log_imports\tbranch_status\n'
for repo in "$@"; do
  if [ ! -f "$repo/go.mod" ]; then
    printf '%s\t%s\tmissing-go-mod\t-\t-\t-\t-\t-\t-\n' "$(basename "$repo")" "$repo"
    continue
  fi

  name=$(basename "$repo")
  dep=no
  grep -q 'github.com/go-go-golems/logcopter' "$repo/go.mod" && dep=yes

  generate_file=no
  [ -f "$repo/logcopter_generate.go" ] && generate_file=yes

  generated_count=$(find "$repo" -path "$repo/.git" -prune -o -name logcopter.go -print | wc -l | tr -d ' ')

  make_check=no
  if [ -f "$repo/Makefile" ] && grep -q '^logcopter-check:' "$repo/Makefile"; then
    make_check=yes
  fi

  make_bump=no
  if [ -f "$repo/Makefile" ] && grep -q '^bump-go-go-golems:' "$repo/Makefile"; then
    make_bump=yes
  fi

  zerolog_imports=$((rg -l '"github.com/rs/zerolog/log"' "$repo" --glob '*.go' --glob '!**/vendor/**' --glob '!**/.git/**' 2>/dev/null || true) | wc -l | tr -d ' ')
  branch_status=$(git -C "$repo" status --short --branch 2>/dev/null | head -1 | tr '\t' ' ')

  printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
    "$name" "$repo" "$dep" "$generate_file" "$generated_count" "$make_check" "$make_bump" "$zerolog_imports" "$branch_status"
done
