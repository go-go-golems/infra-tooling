package rollout

import (
	"path/filepath"
	"testing"
)

func TestPlanGlazedLintReportsNeededOperations(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/repo\n\nrequire github.com/go-go-golems/glazed v1.2.5\n")
	writeFile(t, filepath.Join(repo, "Makefile"), "lint:\n\tgo test ./...\n")

	cfg := Config{Workspace: root}
	cfg.Selection.Include = []string{"repo"}
	ops, err := Plan(cfg, PlanOptions{Profile: ProfileGlazedLint})
	if err != nil {
		t.Fatal(err)
	}
	if !hasPlanOp(ops, "makefile-variable", "needed", "Ensure GLAZED_LINT_BIN is defined") {
		t.Fatalf("expected missing GLAZED_LINT_BIN operation: %#v", ops)
	}
	if !hasPlanOp(ops, "makefile-target", "needed", "Add glazed-lint-build target") {
		t.Fatalf("expected missing build target operation: %#v", ops)
	}
	if !hasPlanOp(ops, "lint-integration", "needed", "Integrate Glazed lint into lint") {
		t.Fatalf("expected lint integration operation: %#v", ops)
	}
}

func TestPlanGlazedLintRecognizesHardenedMakefile(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/repo\n\nrequire github.com/go-go-golems/glazed v1.3.4\n")
	writeFile(t, filepath.Join(repo, "Makefile"), `GLAZED_LINT_BIN ?= /tmp/glazed-lint
GLAZED_LINT_PKG ?= github.com/go-go-golems/glazed/cmd/tools/glazed-lint
GLAZED_VERSION ?= $(shell GOWORK=off go list -m -f '{{.Version}}' github.com/go-go-golems/glazed 2>/dev/null)
GLAZED_LINT_TOOL_VERSION ?= v1.3.4
GLAZED_LINT_FLAGS ?= -glazedclilint.allow-paths=pkg/legacy/
GLAZED_LINT_DIRS ?= ./cmd/... ./pkg/...

glazed-lint-build:
	@GOBIN=$(dir $(GLAZED_LINT_BIN)) GOWORK=off go install $(GLAZED_LINT_PKG)@$(GLAZED_LINT_TOOL_VERSION)

glazed-lint: glazed-lint-build
	GOWORK=off go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)

lint: glazed-lint-build
	golangci-lint run -v
	GOWORK=off go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)

lintmax: glazed-lint-build
	golangci-lint run -v --max-same-issues=100
	GOWORK=off go vet -vettool=$(GLAZED_LINT_BIN) $(GLAZED_LINT_FLAGS) $(GLAZED_LINT_DIRS)
`)
	writeFile(t, filepath.Join(repo, ".github", "workflows", "lint.yml"), "name: lint\njobs:\n  lint:\n    steps:\n      - name: Run Glazed CLI policy linters\n        run: make glazed-lint\n")

	cfg := Config{Workspace: root}
	cfg.Selection.Include = []string{"repo"}
	ops, err := Plan(cfg, PlanOptions{Profile: ProfileGlazedLint})
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range ops {
		if op.Status == "needed" {
			t.Fatalf("did not expect needed operation: %#v", op)
		}
	}
	if !hasPlanOp(ops, "ci-workflow", "present", "Run Glazed CLI policy linters in CI") {
		t.Fatalf("expected CI present operation: %#v", ops)
	}
}

func hasPlanOp(ops []PlanOperation, kind, status, description string) bool {
	for _, op := range ops {
		if op.Kind == kind && op.Status == status && op.Description == description {
			return true
		}
	}
	return false
}
