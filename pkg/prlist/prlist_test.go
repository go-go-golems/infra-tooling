package prlist

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-go-golems/infra-tooling/pkg/prref"
)

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prs.yaml")
	content := `prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2
  - ref: go-go-golems/go-minitrace#11
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	want := []prref.Ref{{Owner: "go-go-golems", Repo: "discord-bot", Number: 9}, {Owner: "go-go-golems", Repo: "goja-git", Number: 2}, {Owner: "go-go-golems", Repo: "go-minitrace", Number: 11}}
	if len(got) != len(want) {
		t.Fatalf("len=%d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d]=%#v want %#v", i, got[i], want[i])
		}
	}
}
