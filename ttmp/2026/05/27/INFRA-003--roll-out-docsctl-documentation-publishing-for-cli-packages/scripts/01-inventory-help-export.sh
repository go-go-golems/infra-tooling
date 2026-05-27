#!/usr/bin/env bash
set -u

workspace=${1:-/home/manuel/workspaces/2026-05-24/add-js-providers}
out_dir=${2:-/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/help-export-inventory}
mkdir -p "$out_dir"
summary="$out_dir/summary.tsv"
: > "$summary"
printf 'repo\tcmd\tstatus\texit_code\tsqlite_size\tnote\n' >> "$summary"

for mod in "$workspace"/*/go.mod; do
  repo_dir=$(dirname "$mod")
  repo=$(basename "$repo_dir")
  if [ ! -d "$repo_dir/cmd" ]; then
    printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$repo" "" "no_cmd_dir" "" "" "no cmd directory" >> "$summary"
    continue
  fi
  found=0
  for main in "$repo_dir"/cmd/*/main.go; do
    [ -f "$main" ] || continue
    found=1
    cmd_dir="./${main#"$repo_dir/"}"
    cmd_dir="${cmd_dir%/main.go}"
    safe_cmd=${cmd_dir#./cmd/}
    log="$out_dir/${repo}__${safe_cmd}.log"
    sqlite_dir="$out_dir/sqlite/${repo}/${safe_cmd}"
    sqlite="$sqlite_dir/help.sqlite"
    rm -rf "$sqlite_dir"
    mkdir -p "$sqlite_dir"
    (
      cd "$repo_dir" || exit 99
      echo "repo=$repo"
      echo "cmd_dir=$cmd_dir"
      echo "branch=$(git branch --show-current 2>/dev/null || true)"
      echo "head=$(git rev-parse --short HEAD 2>/dev/null || true)"
      echo "command=GOWORK=off go run $cmd_dir help export --format sqlite --output-path $sqlite"
      GOWORK=off timeout 90s go run "$cmd_dir" help export --format sqlite --output-path "$sqlite"
    ) > "$log" 2>&1
    code=$?
    size=0
    if [ -s "$sqlite" ]; then
      size=$(wc -c < "$sqlite" | tr -d ' ')
      status="export_ok"
    else
      status="export_failed"
    fi
    note=$(tail -n 5 "$log" | tr '\n\t' '  ' | cut -c1-240)
    printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$repo" "$cmd_dir" "$status" "$code" "$size" "$note" >> "$summary"
  done
  if [ "$found" -eq 0 ]; then
    printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$repo" "" "no_cmd_main" "" "" "no cmd/*/main.go" >> "$summary"
  fi
done

column -t -s $'\t' "$summary" | tee "$out_dir/summary.txt"
