package rollout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInventoryFiltersGlazedRepos(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "with-glazed", "go.mod"), "module example.com/with\n\nrequire github.com/go-go-golems/glazed v1.2.3\n")
	writeFile(t, filepath.Join(root, "with-glazed", "Makefile"), "lint:\n\tgo test ./...\n\nglazed-lint:\n\tgo vet ./...\n")
	writeFile(t, filepath.Join(root, "with-glazed", "cmd", "main.go"), "package main\nfunc main(){}\n")
	writeFile(t, filepath.Join(root, "without", "go.mod"), "module example.com/without\n")

	repos, err := Inventory(root, InventoryOptions{RequireModules: []string{"github.com/go-go-golems/glazed"}, Base: "origin/main"})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repo, got %d: %#v", len(repos), repos)
	}
	repo := repos[0]
	if repo.Name != "with-glazed" || repo.Module != "example.com/with" || repo.GlazedVersion != "v1.2.3" {
		t.Fatalf("unexpected repo metadata: %#v", repo)
	}
	if !repo.HasMakefile {
		t.Fatalf("expected makefile")
	}
	if got := repo.LintTargets; len(got) != 2 || got[0] != "lint" || got[1] != "glazed-lint" {
		t.Fatalf("unexpected targets: %#v", got)
	}
	if len(repo.PackageDirs) != 1 || repo.PackageDirs[0] != "./cmd" {
		t.Fatalf("unexpected package dirs: %#v", repo.PackageDirs)
	}
}

func TestConfigResolveTargetsUsesInclude(t *testing.T) {
	root := t.TempDir()
	cfg := Config{Workspace: root}
	cfg.Selection.Include = []string{"repo-a", filepath.Join(root, "repo-b")}
	targets, err := cfg.ResolveTargets()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(root, "repo-a"), filepath.Join(root, "repo-b")}
	for i := range want {
		if targets[i] != want[i] {
			t.Fatalf("target[%d] = %q, want %q", i, targets[i], want[i])
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
