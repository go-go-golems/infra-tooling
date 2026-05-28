# Apply Glazed CLI policy linting to a repository

Use this playbook for go-go-golems repositories that depend on `github.com/go-go-golems/glazed` and should enforce Glazed CLI conventions in local hooks and CI.

## What the Glazed linter enforces

The reusable vettool is provided by Glazed:

```text
github.com/go-go-golems/glazed/cmd/tools/glazed-lint
```

It currently bundles `glazedclilint`, which checks:

- CLI code should not read raw environment variables with `os.Getenv`; prefer Glazed config/env middleware or an explicit command field.
- CLI commands should not define raw `cobra`/`pflag`/stdlib `flag` flags; prefer `cmds.WithFlags(fields.New(...))`.
- Commands that expose Glazed output sections should implement `RunIntoGlazeProcessor`.

The analyzer skips tests, generated files, and standard Glazed framework bridge paths by default. Repos may add narrowly scoped `-glazedclilint.allow-paths=...` entries for legacy bridge code or intentional non-Glazed helper tools.

Current Glazed releases also support reasoned suppressions:

```go
//glazedclilint:ignore intentional legacy cobra flag bridge while migrating to fields
//glazedclilint:file-ignore generated CLI compatibility layer; tracked for follow-up
```

Every suppression must include a reason. Prefer a one-line suppression on the smallest affected statement; use file-ignore only for reviewed legacy compatibility files. Exact `allow-paths` are still acceptable for generated helpers or packages that should not be analyzed, but avoid broad `cmd/` or `pkg/` exclusions.

## Add Makefile targets

Add these variables near the existing lint variables:

```make
GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell GOWORK=off go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=pkg/analysis/,pkg/cli/,pkg/cmds/fields/,pkg/cmds/logging/,pkg/cmds/sources/,pkg/help/
```

Add the targets:

```make
.PHONY: glazed-lint-build glazed-lint

glazed-lint-build:
	@echo "Building glazed-lint from Glazed module..."
	@if [ -n "$(GLAZED_VERSION)" ] && [ "$(GLAZED_VERSION)" != "(devel)" ]; then \
		echo "Installing $(GLAZED_LINT_PKG)@$(GLAZED_VERSION)"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_VERSION); \
	else \
		echo "Installing $(GLAZED_LINT_PKG) from workspace/module"; \
		GOBIN=$(dir $(GLAZED_LINT_BIN)) go install $(GLAZED_LINT_PKG); \
	fi

glazed-lint: glazed-lint-build
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./cmd/... ./pkg/...
```

Adjust package patterns per repository:

- Use `./cmd/... ./pkg/...` for CLI/application repos.
- Use `./pkg/...` for library repos without `cmd/`.
- If the repo already computes `LINT_DIRS`, prefer `$(LINT_DIRS)` so generated/testdata/temp packages are filtered consistently.

## Wire into existing lint targets

Make local lint run the Glazed analyzer too:

```make
lint: glazed-lint-build golangci-lint-install
	$(GOLANGCI_LINT_BIN) run -v ./...
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./cmd/... ./pkg/...

lintmax: glazed-lint-build golangci-lint-install
	$(GOLANGCI_LINT_BIN) run -v --max-same-issues=100 ./...
	go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) ./cmd/... ./pkg/...
```

If the repo already has custom vettools, keep them and add Glazed lint as an additional `go vet -vettool=...` line.

## Wire into lefthook

If `lefthook.yml` already runs `make lint` or `make lintmax`, no separate hook is required after the Makefile target is wired in.

Otherwise add `make glazed-lint` to the relevant `pre-commit` and `pre-push` commands.

## Wire into GitHub Actions

In the lint workflow, add a step after Go setup and after any generation/setup required by the repo:

```yaml
- name: Run Glazed CLI policy linters
  run: make glazed-lint
```

This keeps CI behavior aligned with local hooks without duplicating the install logic in YAML.

## Deal with existing violations

Run:

```bash
make glazed-lint
```

For each diagnostic, choose one of these options:

1. Fix the command to use Glazed fields/config middleware.
2. Keep intentional concrete logger/config behavior if it is not package CLI policy.
3. Add a narrow `GLAZED_LINT_FLAGS` allow-path for legacy bridge code or helper tools.

Prefer narrow allow paths such as `cmd/tools/` or `pkg/cmds/profiles/` over broad directories like `cmd/` or `pkg/`.

## Validate before committing

Run Glazed lint after any release-train dependency bump so the vettool comes from the same published Glazed version downstream users will consume:

```bash
make bump-go-go-golems   # when participating in a release train
make glazed-lint
make lintmax
```

If `make bump-go-go-golems` is missing, add the generic target from `examples/go-go-golems/Makefile.bump-go-go-golems-gowork-off.snippet.mk` before continuing.

Do not push Glazed lint rollout changes directly to `main`; open a PR and wait for CI/Codex readiness because linter wiring can change developer hooks and release behavior. Merge with a real merge commit (`gh pr merge --merge --delete-branch`), never with a squash merge, so analyzer-policy and suppression changes remain auditable.

Then inspect the diff and commit:

```bash
git diff -- Makefile lefthook.yml .github/workflows
git add Makefile lefthook.yml .github/workflows
git commit -m "Run Glazed CLI policy linting"
```

## Example allow-paths used during rollout

- Geppetto allowed `cmd/tools/` because internal generator/refactor tools use raw flags and are not Glazed user-facing commands.
- Pinocchio allowed existing legacy bridge code in `pkg/cmds/cmdlayers/` plus specific help/pager commands while wiring the policy for the rest of the repo.
- Clay allowed `pkg/cmds/profiles/` and a specific editor command path as existing legacy command-management behavior.
