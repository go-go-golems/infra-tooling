#!/usr/bin/env bash
set -euo pipefail

base=${1:-/home/manuel/code/wesen/go-go-golems/infra-tooling/ttmp/2026/05/27/INFRA-003--roll-out-docsctl-documentation-publishing-for-cli-packages/sources/help-export-inventory}
out="$base/validation.tsv"
: > "$out"
printf 'repo\tcmd\tsqlite\tstatus\tnote\n' >> "$out"

while IFS=$'\t' read -r repo cmd status code size note; do
  [ "$repo" = "repo" ] && continue
  [ "$status" = "export_ok" ] || continue
  safe=${cmd#./cmd/}
  sqlite="$base/sqlite/$repo/$safe/help.sqlite"
  log="$base/${repo}__${safe}.validate.log"
  package="$repo"
  if [ "$repo" = "glazed" ] && [ "$safe" = "glaze" ]; then
    package="glazed"
  fi
  if docsctl validate --file "$sqlite" --package "$package" --version v0.0.0-inventory > "$log" 2>&1; then
    vstatus=validate_ok
  else
    vstatus=validate_failed
  fi
  vnote=$(tail -n 5 "$log" | tr '\n\t' '  ' | cut -c1-240)
  printf '%s\t%s\t%s\t%s\t%s\n' "$repo" "$cmd" "$sqlite" "$vstatus" "$vnote" >> "$out"
done < "$base/summary.tsv"

column -t -s $'\t' "$out" | tee "$base/validation.txt"
