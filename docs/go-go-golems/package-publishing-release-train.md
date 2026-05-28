# go-go-golems package publishing and dependency rollout playbook

This playbook captures the reusable operational workflow used during the logcopter rollout across Glazed, Clay, Geppetto, and Pinocchio. It is for publishing upstream go-go-golems packages, bumping downstream modules, validating against published dependencies, and merging PRs only after CI and Codex review are satisfied.

## Tooling in this repo

Installed CLI:

- `ggg pr ready` — one-shot PR readiness check with optional `--findings` and Glazed structured output.
- `ggg pr codex-trigger` — posts `@codex review` when it is safe to do so; supports `--file prs.yaml`, `--wait-for-auto`, `--dry-run`, and `--force`.
- `ggg pr codex-comments` — lists Codex-authored review bodies and inline review comments.
- `ggg batch ready` — checks or watches a YAML PR list without blocking on one PR.
- `ggg release tag-patch`, `ggg release tag-minor`, and `ggg release tag-major` — compute, create, push, and proxy-verify Go module release tags without detaching the operator worktree.
- `ggg release watch` — wait for a tag-triggered release workflow and optionally verify docs publishing.
- `ggg release verify-docs` — verify that `docs.yolo.scapegoat.dev/<package>/<version>` is live and contains the requested package/version.
- `ggg release preflight` — catch common release-tag failures before pushing tags, including stale GoReleaser paths, tree-sitter/CGO conflicts, frontend generation prerequisites, and docsctl workflow inputs.

Historical scripts remain under `scripts/go-go-golems/`, but new operator workflows should use the installed `ggg` binary.

Playbooks and snippets:

- `docs/go-go-golems/playbooks/logcopter-package-rollout-playbook.md` — detailed logcopter package-logger rollout guide.
- `docs/go-go-golems/playbooks/pr-readiness-check-scripts.md` — design and usage notes for `ggg` PR readiness commands.
- `examples/go-go-golems/Makefile.bump-go-go-golems.snippet.mk` — generic dependency bump target.
- `examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk` — dependency bump target that forces published-module resolution.

## Release train principle

Rollouts must follow dependency order. Do not bump and merge downstream repositories until the upstream repositories they need have been merged and published.

Typical order for core go-go-golems packages:

```text
logcopter -> glazed/clay -> geppetto -> pinocchio -> leaf applications
```

The exact order depends on the current `go.mod` graph. Inspect direct dependencies rather than relying on memory.

Every release-train participant should have two logcopter-related baselines: generated package loggers (`logcopter_generate.go`, checked-in `**/logcopter.go`, and `make logcopter-check`) and a generic `make bump-go-go-golems` target. Generated loggers provide area-scoped runtime logging; the bump target is the standard handoff point between upstream releases and downstream validation. Together they replace stale `bump-glazed` lists and keep logcopter, Glazed, xgoja, docsctl-capable CLIs, and leaf applications on the same dependency-bump path. Prefer the `GOWORK=off` variant in active workspaces.

## Main-branch protection rule

Never push release-train changes directly to `main`, even for small Makefile, generated-file, or documentation updates. Always create a branch, open a pull request, wait for CI and Codex/readiness checks, and merge through the repository policy.

This also applies to emergency fixes after an accidental main push: repair with a follow-up PR unless the user explicitly authorizes a direct hotfix. Release tags may be pushed by the release tooling after the commit is already on `main`, but source changes must arrive through PRs.

## Early downstream PRs

You may open downstream PRs before every upstream release is published in order to get CI and Codex feedback early. This is useful for large release trains with many leaf repositories.

Rules for early PRs:

1. It is fine if a downstream PR temporarily fails because an upstream package has not been tagged yet.
2. Do not merge a downstream PR until its required upstream tags are visible and `GOWORK=off` validation passes.
3. Use `ggg batch ready` to monitor all open PRs while still merging/releasing in dependency order.

Store PRs as YAML:

```yaml
prs:
  - https://github.com/go-go-golems/<repo-a>/pull/<n>
  - repo: go-go-golems/<repo-b>
    number: <n>
  - ref: go-go-golems/<repo-c>#<n>
```

Then trigger and watch them with `ggg`:

```bash
ggg pr codex-trigger --file /tmp/prs.yaml --wait-for-auto 30s
ggg batch ready /tmp/prs.yaml
ggg batch ready /tmp/prs.yaml --watch --until actionable --interval-seconds 30 --timeout-seconds 1800
```

Batch watch mode stops as soon as there is operator work: a terminal failure, a Codex feedback state, all PRs ready, or even one ready PR while others are still waiting. Treat exit code `5` as “partial progress is actionable”; inspect the table and proceed with the next dependency-order merge/release step.

## Per-repository workflow

### 1. Land and publish the upstream package

In the upstream repository:

1. Ensure the PR is merged.
2. Ensure the release/tag/published module version exists.
3. Confirm what downstream should consume:

```bash
go list -m -versions github.com/go-go-golems/<upstream>
git fetch --tags
git describe origin/main --tags --always
```

If `origin/main` is ahead of the latest tag and downstream needs those commits, publish a new release before proceeding downstream.

For Go module releases, run preflight before creating any tag, then tag:

```bash
ggg release preflight --output json
ggg release tag-patch --dry-run --yes --output json
ggg release tag-patch --yes --output json
```

Use `ggg release preflight --strict` when you want warnings (for example frontend generation or pnpm setup warnings) to block the release until reviewed.

If the package also publishes docs through docsctl, watch the release and verify the docs URL:

```bash
ggg release watch \
  --repo go-go-golems/<repo> \
  --workflow release.yaml \
  --tag vX.Y.Z \
  --verify-docs \
  --package <package>
```

Use `--no-stream --output json` when saving rollout evidence to a ticket. If the release fails, inspect the emitted `failed_log_command` first.

### 2. Bump downstream go-go-golems dependencies

In the downstream repository, use the generic target:

```bash
make bump-go-go-golems
```

If the repository still only has a legacy `bump-glazed` target, add `bump-go-go-golems` before continuing. Keep `bump-glazed` as a temporary compatibility alias only if callers still use it; do not extend hand-maintained dependency lists.

If there is any chance a local `go.work` can hide missing releases, use a `GOWORK=off` variant or run the equivalent commands manually:

```bash
deps="$(awk '/^require[[:space:]]+github\.com\/go-go-golems\// { print $2 } /^[[:space:]]*github\.com\/go-go-golems\// { print $1 }' go.mod | sort -u)"
for dep in $deps; do GOWORK=off go get "${dep}@latest"; done
GOWORK=off go mod tidy
```

Review the result:

```bash
git diff -- go.mod go.sum
go list -m github.com/go-go-golems/...
```

### 3. Validate without local workspace assumptions

Prefer `GOWORK=off` for smoke tests that prove the published dependency graph works:

```bash
GOWORK=off go test ./...
```

For logcopter-enabled repositories, run the non-mutating generated-file freshness check before any mutating generation command:

```bash
make logcopter-check
# only when intentionally refreshing generated files:
# go generate ./...
```

Run repo-specific checks too, for example:

```bash
make test
make lint
make ci
```

### 4. Commit and push

Commit only the intended dependency changes and related generated/check artifacts:

```bash
git status --short
git add go.mod go.sum
git commit -m "Bump go-go-golems dependencies"
git push <remote> <branch>
```

### 5. Trigger or wait for Codex review

After opening a PR, let `ggg` wait 20-30 seconds before manually triggering Codex because Codex often starts an automatic review shortly after PR creation. If no automatic review appears, or if a later push needs a fresh review, trigger it without `--force` first:

```bash
ggg pr codex-trigger https://github.com/go-go-golems/<repo>/pull/<n> --wait-for-auto 30s
```

Do not repeatedly force-trigger Codex when `ggg pr ready` already reports a satisfied signal; use `--force` only when intentionally replacing a stale or stuck run.

Check readiness once, or include detailed findings when debugging:

```bash
ggg pr ready https://github.com/go-go-golems/<repo>/pull/<n>
ggg pr ready https://github.com/go-go-golems/<repo>/pull/<n> --findings
```

For single-PR watch behavior, use:

```bash
ggg pr watch https://github.com/go-go-golems/<repo>/pull/<n> --interval-seconds 30 --timeout-seconds 1800
```

For post-merge GitHub Actions status, use `ggg run status` for one repository or `ggg batch actions` for a rollout manifest. Ignore only known-noisy workflow failures such as Secret Scanning; every other failure blocks release tags and downstream bumps.

```bash
ggg run status \
  --repo go-go-golems/<repo> \
  --branch main \
  --sha <merge-sha> \
  --ignore-workflow "Secret Scanning" \
  --output table

cat > /tmp/actions.yaml <<'EOF'
repos:
  - repo: go-go-golems/<repo-a>
    branch: main
    sha: <merge-sha-a>
  - repo: go-go-golems/<repo-b>
    branch: main
    sha: <merge-sha-b>
EOF

ggg batch actions /tmp/actions.yaml \
  --ignore-workflow "Secret Scanning" \
  --watch \
  --output json
```

`ggg run status` and `ggg batch actions` exit `0` when all matching runs are successful or ignored, `1` when a non-ignored workflow failed, and `2` while matching runs are still pending.

For batch watch behavior, put the PRs in a YAML list and use batch watch. Use `--until actionable` for release trains where partial readiness should wake the operator; use `--until all-ready` when you want to keep polling through partial readiness.

```bash
printf "prs:\n  - https://github.com/go-go-golems/<repo>/pull/<n>\n" > /tmp/prs.yaml
ggg batch ready /tmp/prs.yaml --watch --until actionable --interval-seconds 30 --timeout-seconds 1800
```

A PR is considered ready when:

- GitHub mergeability is clean (no conflicts / blocked merge state);
- status checks exist;
- every status check is completed successfully, skipped, or neutral;
- a Codex signal exists;
- the latest Codex signal has a thumbs-up reaction or a satisfied body such as `Didn't find any major issues. :+1:`;
- the latest Codex signal has no `EYES` reaction;
- the latest Codex-authored body is empty, benign, or satisfied rather than substantive review feedback.

### 6. Merge only after readiness succeeds

After `ggg pr ready` or `ggg batch ready` exits successfully for the target PR, merge using a real merge commit and delete the remote branch:

```bash
gh pr merge <n> --merge --delete-branch
```

Never use squash merges for release-train work. Preserving branch commits matters for auditability, tag ancestry, generated-code review, and downstream dependency-bump archaeology. If the repository UI defaults to “Squash and merge”, stop and switch to a merge commit.

If you intentionally used `--delete-branch=false` during a sensitive rollout, delete the branch immediately after verifying the merge commit:

```bash
gh api -X DELETE repos/go-go-golems/<repo>/git/refs/heads/<branch>
```


### 7. Cross-cutting rollout checks

Use these checks before declaring a repository release-train-ready:

- **Logcopter:** release-train repos should have `logcopter_generate.go`, checked-in generated `**/logcopter.go`, a direct/tool dependency on published `github.com/go-go-golems/logcopter`, and `make logcopter-check`; run the non-mutating check before mutating `go generate ./...`.
- **Docsctl:** release workflows should call `go-go-golems/infra-tooling/.github/workflows/publish-docsctl.yml@main`, keep `id-token: write` scoped to the publish job, and be verified with `ggg release watch --verify-docs` plus `ggg release verify-docs`.
- **xgoja:** JavaScript-provider or runtime repos must validate with `GOWORK=off` after upstream xgoja tags; check GoReleaser CGO settings when tree-sitter packages are present.
- **Glazed linting:** run `make glazed-lint` after dependency bumps so downstream code is checked against the released Glazed analyzer; use reasoned `//glazedclilint:ignore ...` suppressions or exact allow paths rather than broad directory exclusions.
- **Release preflight:** run `ggg release preflight --output json` before every tag and save the JSON when working from a ticket.

If the PR touches `.github/workflows/*`, the GitHub CLI token needs `workflow` scope. If merge fails with a workflow-scope error, refresh auth:

```bash
gh auth refresh -h github.com -s workflow
```

Then retry the merge.

## Common gotchas

- Local `go.work` can hide missing published upstream symbols. Use `GOWORK=off` for downstream readiness checks.
- `go generate ./...` is mutating. For generated-file drift checks, run the non-mutating checker first.
- A merged upstream PR is not the same as a published upstream module version. Check tags/module versions before bumping downstream.
- Codex `EYES` reactions mean review may still be running; do not merge until the readiness checker accepts the latest signal.
- Merge conflicts or blocked mergeability are terminal readiness failures. Rebase/merge the base branch and rerun readiness instead of trusting green checks or Codex alone.
- If Codex leaves substantive review text, treat the PR as not ready even when Actions are green. `ggg pr ready` and `ggg batch ready` exit immediately with status `3` in this case so the operator can inspect and address the review instead of looping until timeout.
- If `govulncheck` reports standard-library vulnerabilities, bump the repo's Go directive/toolchain to the fixed Go version and rerun `GOWORK=off govulncheck ./...`.
- If `golangci-lint-action` fails because its binary was built with an older Go version than the repo target, bump the action version or switch to a repo-managed lint install after `actions/setup-go`.
- If `securego/gosec@master` runs with an older Go than `actions/setup-go`, prefer installing `gosec` with `go install` after setup and running the binary directly.
- If Dependency Review is unsupported because the repository dependency graph is disabled, either enable the dependency graph in repository settings or mark that workflow step `continue-on-error: true` until settings are fixed.
