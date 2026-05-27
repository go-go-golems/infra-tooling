#!/usr/bin/env bash
set -euo pipefail

# Inventory local repositories that depend on github.com/go-go-golems/glazed.
# Output is TSV so it can be pasted into docs or processed with awk/cut.
#
# Usage:
#   01-inventory-glazed-repos.sh [root ...]
#
# If no roots are provided, scan the active go-go-golems workspace roots used by
# the recent xgoja/infra-tooling work.

if [[ $# -gt 0 ]]; then
  roots=("$@")
else
  roots=(
    "/home/manuel/workspaces/2026-05-24/add-js-providers"
    "/home/manuel/code/wesen/go-go-golems"
  )
fi

printf 'repo_path\tmodule\tglazed_version\thas_makefile\tlint_targets\thas_workflows\thas_lefthook\tpackages\tgit_status\n'

seen_tmp="$(mktemp)"
trap 'rm -f "$seen_tmp"' EXIT

for root in "${roots[@]}"; do
  [[ -d "$root" ]] || continue
  while IFS= read -r -d '' gomod; do
    repo="$(dirname "$gomod")"
    real_repo="$(cd "$repo" && pwd -P)"
    if grep -Fxq "$real_repo" "$seen_tmp"; then
      continue
    fi
    echo "$real_repo" >> "$seen_tmp"
    if ! grep -q 'github.com/go-go-golems/glazed' "$gomod"; then
      continue
    fi
    module="$(awk '/^module[[:space:]]+/ {print $2; exit}' "$gomod")"
    glazed_version="$(awk '$1 == "github.com/go-go-golems/glazed" {print $2; found=1} found && $1 == "github.com/go-go-golems/glazed" {print $2; exit}' "$gomod" | head -1)"
    [[ -n "$glazed_version" ]] || glazed_version="indirect-or-replace"
    has_makefile=no
    lint_targets=""
    if [[ -f "$repo/Makefile" ]]; then
      has_makefile=yes
      lint_targets="$(awk -F: '/^[A-Za-z0-9_.-]+:/ {print $1}' "$repo/Makefile" | grep -E '^(lint|lintmax|ci|test|glazed-lint)' | paste -sd ',' - || true)"
    fi
    has_workflows=no
    [[ -d "$repo/.github/workflows" ]] && has_workflows=yes
    has_lefthook=no
    [[ -f "$repo/lefthook.yml" || -f "$repo/lefthook.yaml" ]] && has_lefthook=yes
    packages=""
    [[ -d "$repo/cmd" ]] && packages="${packages}cmd "
    [[ -d "$repo/pkg" ]] && packages="${packages}pkg "
    packages="${packages%% }"
    [[ -n "$packages" ]] || packages="root/other"
    git_status="not-git"
    if git -C "$repo" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
      if [[ -n "$(git -C "$repo" status --short)" ]]; then
        git_status="dirty"
      else
        git_status="clean"
      fi
    fi
    printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n' \
      "$real_repo" "$module" "$glazed_version" "$has_makefile" "$lint_targets" "$has_workflows" "$has_lefthook" "$packages" "$git_status"
  done < <(find "$root" -mindepth 2 -maxdepth 4 -name go.mod -type f -print0 2>/dev/null)
done | sort -u
