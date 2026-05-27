package rollout

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRunsCommandsAndWritesLogs(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/repo\n")
	cfg := Config{Workspace: root}
	cfg.Selection.Include = []string{"repo"}
	cfg.Validation.Commands = []ValidationCommand{{Name: "hello", Run: "echo hello"}}
	cfg.Validation.LogDir = filepath.Join(root, "logs")

	results, err := Validate(context.Background(), cfg, ValidationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || !results[0].OK || results[0].ExitCode != 0 {
		t.Fatalf("unexpected results: %#v", results)
	}
	b, err := os.ReadFile(results[0].LogPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "hello") {
		t.Fatalf("log did not contain command output: %s", b)
	}
}

func TestValidateContinueOnError(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/repo\n")
	cfg := Config{Workspace: root}
	cfg.Selection.Include = []string{"repo"}
	cfg.Validation.ContinueOnError = true
	cfg.Validation.Commands = []ValidationCommand{{Name: "fail", Run: "exit 7"}, {Name: "ok", Run: "echo ok"}}
	cfg.Validation.LogDir = filepath.Join(root, "logs")

	results, err := Validate(context.Background(), cfg, ValidationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || results[0].ExitCode != 7 || results[0].OK || !results[1].OK {
		t.Fatalf("unexpected results: %#v", results)
	}
}
