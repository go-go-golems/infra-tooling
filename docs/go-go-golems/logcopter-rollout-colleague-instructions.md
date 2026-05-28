# Concise instructions: roll out logcopter across go-go-golems repos

Use this checklist when applying generated logcopter package loggers to additional go-go-golems repositories. A complete conversion means checked-in generated `logcopter.go` files, `logcopter_generate.go`, `make logcopter-check`, and the generic `bump-go-go-golems` release-train target.

## 1. Build the dependency sequence first

For every candidate repo, inspect direct go-go-golems dependencies:

```bash
cd /path/to/repo
awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $1 }' go.mod | sort -u
```

Create edges from `repo -> dependency`. Roll out from dependencies to dependents. In practice: foundational libraries first, leaf CLIs/apps last.

Rules:

- If repo A depends on repo B, migrate/merge/release B before bumping A.
- Do not trust local `go.work`; validate downstream with `GOWORK=off`.
- A merged upstream PR is not enough if downstream needs published symbols; confirm tags/module versions exist.

Useful checks:

```bash
go list -m -versions github.com/go-go-golems/<repo>
git fetch --tags
git describe origin/main --tags --always
```

## 2. For each repo, add logcopter and the generator

```bash
go get github.com/go-go-golems/logcopter@latest
go get -tool github.com/go-go-golems/logcopter/cmd/logcopter-gen@latest
go mod tidy
go tool logcopter-gen -h
```

Remove any temporary local `replace github.com/go-go-golems/logcopter => ...` before opening a PR.

## 3. Add `logcopter_generate.go`

At repo root, create `logcopter_generate.go`. Use the real root package name, not `package main`, unless the repo root is already a command package.

Library-style repo:

```go
package <repo>

//go:generate go tool logcopter-gen -area-prefix go-go-golems.<repo> -strip-prefix github.com/go-go-golems/<repo> ./pkg/...
```

Repo with command packages that should also get package loggers:

```go
package <repo>

//go:generate go tool logcopter-gen -area-prefix go-go-golems.<repo> -strip-prefix github.com/go-go-golems/<repo> ./pkg/... ./cmd/...
```

Area convention: `go-go-golems.<repo>.<path-tail>`, keeping `pkg` and `cmd` path components.

## 4. Generate and fix collisions

```bash
go generate ./...
```

If generated `var log` conflicts with imports named `log`:

- remove `github.com/rs/zerolog/log` imports that were only used for package diagnostics;
- keep call sites like `log.Debug().Msg(...)` unchanged;
- alias intentional global zerolog package use as `zlog`;
- alias standard library `log` as `stdlog`;
- do not convert APIs that intentionally accept/inject `zerolog.Logger`.

## 5. Add Makefile targets

Add or copy the generic dependency bump target from infra-tooling after adding generated package loggers. Do this for every release-train repository that has direct `github.com/go-go-golems/...` dependencies:

```text
examples/go-go-golems/Makefile.bump-go-go-golems.snippet.mk
```

For repos that must avoid workspace leakage, use:

```text
examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk
```

Validate the target without mutating dependencies:

```bash
make -n bump-go-go-golems
```

Add logcopter generation targets when this repository is adopting generated package loggers:

```make
.PHONY: logcopter-generate
logcopter-generate:
	go generate ./...

.PHONY: logcopter-check
logcopter-check:
	go tool logcopter-gen -check -area-prefix go-go-golems.<repo> -strip-prefix github.com/go-go-golems/<repo> ./pkg/...
```

If `logcopter_generate.go` covers `./cmd/...`, include the same package patterns in `logcopter-check`.

## 6. Update CI ordering

In CI, run the non-mutating check before any mutating generation command:

```bash
make logcopter-check
go generate ./...
```

Never rely on `go generate ./...` as the drift check; it rewrites files.

## 7. Validate locally

Minimum validation:

```bash
make logcopter-check
GOWORK=off go test ./...
```

Also run repo-specific checks (`make test`, `make lint`, smoke tests, web checks, etc.).

Confirm no generated drift:

```bash
git status --short
git diff -- go.mod go.sum '**/logcopter.go' logcopter_generate.go Makefile .github/workflows
```

## 8. Commit, push, PR, Codex, merge

Commit focused changes on a non-main branch. Never push rollout changes directly to `main`, even if the repository allows it technically:

```bash
git checkout -b task/logcopter-<repo>
git add go.mod go.sum logcopter_generate.go Makefile .github/workflows '**/logcopter.go'
git commit -m "Adopt logcopter package loggers"
git push <remote> HEAD
```

Open a PR. Let `ggg` wait 20-30 seconds for the automatic Codex review to appear, then trigger Codex only if no run/satisfied signal appears:

```bash
ggg pr codex-trigger https://github.com/go-go-golems/<repo>/pull/<n> --wait-for-auto 30s
```

For many PRs, put the URLs in a YAML file and trigger/check them as a batch:

```yaml
prs:
  - https://github.com/go-go-golems/<repo-a>/pull/<n>
  - repo: go-go-golems/<repo-b>
    number: <n>
```

```bash
ggg pr codex-trigger --file /tmp/prs.yaml --wait-for-auto 30s
ggg batch ready /tmp/prs.yaml
```

Wait until CI, mergeability, and Codex are ready for one PR:

```bash
ggg pr watch https://github.com/go-go-golems/<repo>/pull/<n> --interval-seconds 30 --timeout-seconds 1800
```

For batch work, keep using:

```bash
printf "prs:\n  - https://github.com/go-go-golems/<repo>/pull/<n>\n" > /tmp/prs.yaml
ggg batch ready /tmp/prs.yaml --watch --until actionable --interval-seconds 30 --timeout-seconds 1800
```

Merge only after readiness succeeds, and delete the remote branch (`gh pr merge --squash --delete-branch`) so follow-up rollouts do not accumulate stale branches. If readiness reports `merge_conflict`, rebase or merge the base branch first.

## 9. Downstream bump after upstream release

After an upstream repo merges and publishes, bump each downstream repo in dependency order:

```bash
make bump-go-go-golems
GOWORK=off go test ./...
git add go.mod go.sum
git commit -m "Bump go-go-golems dependencies"
git push <remote> <branch>
```

Then rerun `ggg pr ready` or `ggg batch ready` and merge only after readiness succeeds.

## 10. Reference docs

- Full rollout playbook: `docs/go-go-golems/playbooks/logcopter-package-rollout-playbook.md`
- Release train playbook: `docs/go-go-golems/package-publishing-release-train.md`
- PR readiness with `ggg`: `docs/go-go-golems/playbooks/pr-readiness-check-scripts.md`
