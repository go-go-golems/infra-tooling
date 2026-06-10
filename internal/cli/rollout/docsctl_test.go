package rollout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectDocsWorkflowPrefersSeparatePublishDocsWorkflow(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, ".github/workflows/release.yaml", "jobs: {}\n")
	writeFile(t, repo, ".github/workflows/publish-docs.yaml", "jobs: {}\n")

	got := detectDocsWorkflow(repo)
	want := ".github/workflows/publish-docs.yaml"
	if got != want {
		t.Fatalf("detectDocsWorkflow() = %q, want %q", got, want)
	}
}

func TestDocsctlInventoryUsesDocsWorkflow(t *testing.T) {
	workspace := t.TempDir()
	repo := filepath.Join(workspace, "tool")
	writeFile(t, repo, "go.mod", "module example.com/tool\n")
	writeFile(t, repo, "cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeFile(t, repo, ".github/workflows/release.yaml", "jobs: {}\n")
	writeFile(t, repo, ".github/workflows/publish-docs.yaml", "jobs: {}\n")

	candidates, err := docsctlInventory(workspace, nil, nil, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 1 {
		t.Fatalf("got %d candidates, want 1: %#v", len(candidates), candidates)
	}
	if candidates[0].Workflow != ".github/workflows/publish-docs.yaml" {
		t.Fatalf("candidate workflow = %q", candidates[0].Workflow)
	}
}

func TestRewriteDocsctlOutputPathHandlesNestedOverride(t *testing.T) {
	command := "mkdir -p .docsctl && (cd cmd/goja-bleve && GOWORK=off go run . help export --format sqlite --output-path ../../.docsctl/help.sqlite)"
	got := rewriteDocsctlOutputPath(command, "/tmp/ggg-docsctl-123/help.sqlite")
	want := "mkdir -p .docsctl && (cd cmd/goja-bleve && GOWORK=off go run . help export --format sqlite --output-path '/tmp/ggg-docsctl-123/help.sqlite')"
	if got != want {
		t.Fatalf("rewriteDocsctlOutputPath() = %q, want %q", got, want)
	}
}

func TestRewriteDocsctlOutputPathHandlesEqualsAndQuotedValues(t *testing.T) {
	command := `go run ./cmd/tool help export --format sqlite --output-path=".docsctl/help.sqlite"`
	got := rewriteDocsctlOutputPath(command, "/tmp/ggg-docsctl-456/help.sqlite")
	want := `go run ./cmd/tool help export --format sqlite --output-path '/tmp/ggg-docsctl-456/help.sqlite'`
	if got != want {
		t.Fatalf("rewriteDocsctlOutputPath() = %q, want %q", got, want)
	}
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
