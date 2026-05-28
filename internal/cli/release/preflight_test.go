package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunPreflightDetectsGoReleaserPlaceholders(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "go.mod", "module example.com/tool\n")
	writeFile(t, repo, ".goreleaser.yaml", "version: 2\nproject_name: XXX\nbuilds:\n  - main: ./cmd/XXX\n    binary: XXX\n")

	res := runPreflight(&preflightSettings{Repo: repo})
	if res.OK {
		t.Fatalf("expected preflight to fail: %#v", res.Findings)
	}
	assertFinding(t, res.Findings, "goreleaser_placeholder")
	assertFinding(t, res.Findings, "goreleaser_missing_main")
}

func TestRunPreflightDetectsCGODisabledTreeSitter(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "go.mod", "module example.com/tool\nrequire github.com/tree-sitter/tree-sitter-javascript v0.25.0\n")
	writeFile(t, repo, "cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeFile(t, repo, ".goreleaser.yaml", "version: 2\nbuilds:\n  - env:\n      - CGO_ENABLED=0\n    main: ./cmd/tool\n")

	res := runPreflight(&preflightSettings{Repo: repo})
	if res.OK {
		t.Fatalf("expected preflight to fail: %#v", res.Findings)
	}
	assertFinding(t, res.Findings, "cgo_disabled_with_tree_sitter")
}

func TestRunPreflightWarnsForFrontendGenerateWithoutInstall(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "go.mod", "module example.com/tool\n")
	writeFile(t, repo, "cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeFile(t, repo, "web/review-site/package.json", "{\"scripts\":{\"build\":\"tsc -b\"}}\n")
	writeFile(t, repo, "web/review-site/pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	writeFile(t, repo, ".github/workflows/release.yaml", "jobs:\n  release:\n    steps: []\n")
	writeFile(t, repo, ".goreleaser.yaml", "version: 2\nbefore:\n  hooks:\n    - go generate ./...\nbuilds:\n  - main: ./cmd/tool\n")

	res := runPreflight(&preflightSettings{Repo: repo})
	if !res.OK {
		t.Fatalf("warnings should not fail non-strict preflight: %#v", res.Findings)
	}
	assertFinding(t, res.Findings, "generate_frontend_without_pnpm_setup")
	assertFinding(t, res.Findings, "generate_frontend_without_install")

	strict := runPreflight(&preflightSettings{Repo: repo, Strict: true})
	if strict.OK {
		t.Fatalf("strict preflight should fail on warnings: %#v", strict.Findings)
	}
}

func TestRunPreflightPassesFixedCSSVisualDiffShape(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, repo, "go.mod", "module example.com/tool\n")
	writeFile(t, repo, "cmd/tool/main.go", "package main\nfunc main() {}\n")
	writeFile(t, repo, "web/review-site/package.json", "{\"scripts\":{\"build\":\"tsc -b\"}}\n")
	writeFile(t, repo, "web/review-site/pnpm-lock.yaml", "lockfileVersion: '9.0'\n")
	writeFile(t, repo, ".github/workflows/release.yaml", "jobs:\n  release:\n    steps:\n      - uses: pnpm/action-setup@v4\n      - run: pnpm --dir web/review-site install --frozen-lockfile\n")
	writeFile(t, repo, ".goreleaser.yaml", "version: 2\nbefore:\n  hooks:\n    - go generate ./...\nbuilds:\n  - main: ./cmd/tool\n")

	res := runPreflight(&preflightSettings{Repo: repo})
	if !res.OK || len(res.Findings) != 0 {
		t.Fatalf("expected clean preflight, got ok=%v findings=%#v", res.OK, res.Findings)
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

func assertFinding(t *testing.T, findings []preflightFinding, code string) {
	t.Helper()
	for _, f := range findings {
		if f.Code == code {
			return
		}
	}
	t.Fatalf("missing finding %q in %#v", code, findings)
}
