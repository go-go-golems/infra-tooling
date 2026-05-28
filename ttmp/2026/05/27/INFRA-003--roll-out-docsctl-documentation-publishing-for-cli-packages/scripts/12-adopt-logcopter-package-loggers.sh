#!/usr/bin/env bash
set -euo pipefail

# Adopt generated logcopter package loggers in a Go repository.
#
# This script performs the mechanical part only:
#   - add logcopter runtime dependency and logcopter-gen Go tool;
#   - create logcopter_generate.go at repo root;
#   - append logcopter-generate/logcopter-check Makefile targets if missing;
#   - run go generate ./...;
#   - remove github.com/rs/zerolog/log imports in packages where a generated
#     logcopter.go now provides the package-local `log` variable;
#   - gofmt, go mod tidy, and run logcopter-check.
#
# Usage:
#   12-adopt-logcopter-package-loggers.sh \
#     --repo /path/to/repo \
#     --module github.com/go-go-golems/<repo> \
#     --package <root-package-name> \
#     --area-prefix go-go-golems.<repo> \
#     --patterns './internal/... ./pkg/... ./cmd/...'
#
# Notes:
#   - Choose patterns deliberately. Avoid generated asset directories and
#     examples/tools if their package logs should not be area-controlled.
#   - Review all import removals; APIs that intentionally use injected
#     zerolog.Logger should not be changed by this script.
#   - The script intentionally does not commit.

repo=
module=
root_package=
area_prefix=
patterns=

while [ "$#" -gt 0 ]; do
  case "$1" in
    --repo) repo=${2:?}; shift 2 ;;
    --module) module=${2:?}; shift 2 ;;
    --package) root_package=${2:?}; shift 2 ;;
    --area-prefix) area_prefix=${2:?}; shift 2 ;;
    --patterns) patterns=${2:?}; shift 2 ;;
    -h|--help) sed -n '1,45p' "$0"; exit 0 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

: "${repo:?--repo is required}"
: "${module:?--module is required}"
: "${root_package:?--package is required}"
: "${area_prefix:?--area-prefix is required}"
: "${patterns:?--patterns is required}"

cd "$repo"

if [ ! -f go.mod ]; then
  echo "no go.mod in $repo" >&2
  exit 1
fi

if ! grep -q "^module ${module}$" go.mod; then
  echo "warning: go.mod module does not exactly match ${module}" >&2
  grep '^module ' go.mod >&2 || true
fi

GOWORK=off go get github.com/go-go-golems/logcopter@latest
GOWORK=off go get -tool github.com/go-go-golems/logcopter/cmd/logcopter-gen@latest

cat > logcopter_generate.go <<EOF
package ${root_package}

//go:generate go tool logcopter-gen -area-prefix ${area_prefix} -strip-prefix ${module} ${patterns}
EOF

if [ -f Makefile ] && ! grep -q '^logcopter-generate:' Makefile; then
  cat >> Makefile <<EOF

.PHONY: logcopter-generate
logcopter-generate:
	GOWORK=off go generate ./...
EOF
fi

if [ -f Makefile ] && ! grep -q '^logcopter-check:' Makefile; then
  cat >> Makefile <<EOF

.PHONY: logcopter-check
logcopter-check:
	GOWORK=off go tool logcopter-gen -area-prefix ${area_prefix} -strip-prefix ${module} -check ${patterns}
EOF
fi

# Generate package loggers directly first, before running repository-wide
# go:generate hooks. This lets us remove conflicting zerolog/log imports before
# unrelated generators compile packages that now have a generated `var log`.
GOWORK=off go tool logcopter-gen -area-prefix "${area_prefix}" -strip-prefix "${module}" ${patterns}

python3 - <<'PY'
from pathlib import Path
for lf in Path('.').rglob('logcopter.go'):
    if '.git' in lf.parts:
        continue
    d = lf.parent
    for f in d.glob('*.go'):
        if f.name == 'logcopter.go':
            continue
        s = f.read_text()
        ns = s.replace('\n\t"github.com/rs/zerolog/log"', '')
        ns = ns.replace('\n\tlog "github.com/rs/zerolog/log"', '')
        ns = ns.replace('\n"github.com/rs/zerolog/log"', '')
        if ns != s:
            f.write_text(ns)
            print(f)
PY

# Avoid shell glob limits by asking gofmt for tracked + new Go files excluding common frontend/vendor dirs.
find . \
  -path './.git' -prune -o \
  -path './vendor' -prune -o \
  -path './web' -prune -o \
  -path './ttmp' -prune -o \
  -name '*.go' -print0 | xargs -0 gofmt -w

# Now run the repository's go:generate hooks so embedded/generated assets remain
# fresh. Review the diff because this can touch non-logcopter generated files.
GOWORK=off go generate ./...

GOWORK=off go mod tidy
if command -v make >/dev/null && [ -f Makefile ]; then
  make logcopter-check
else
  GOWORK=off go tool logcopter-gen -area-prefix "${area_prefix}" -strip-prefix "${module}" -check ${patterns}
fi
