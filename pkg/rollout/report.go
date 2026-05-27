package rollout

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func MarkdownReport(cfg Config) (string, error) {
	repos, err := reposForConfig(cfg)
	if err != nil {
		return "", err
	}
	branches, _ := BranchStatus(cfg)
	branchByRepo := map[string]BranchResult{}
	for _, b := range branches {
		branchByRepo[b.Repo.Name] = b
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Rollout report: %s\n\n", cfg.Name)
	fmt.Fprintf(&buf, "- ID: `%s`\n", cfg.ID)
	fmt.Fprintf(&buf, "- Workspace: `%s`\n", cfg.Workspace)
	fmt.Fprintf(&buf, "- Branch: `%s`\n", cfg.Branch)
	fmt.Fprintf(&buf, "- Base: `%s`\n\n", cfg.Base)
	fmt.Fprintf(&buf, "## Targets\n\n")
	fmt.Fprintf(&buf, "| Repo | Module | Glazed | Branch | Ahead | Dirty | Makefile targets |\n")
	fmt.Fprintf(&buf, "| --- | --- | --- | --- | ---: | --- | --- |\n")
	for _, repo := range repos {
		dirty := "clean"
		if repo.DirtyTracked || repo.DirtyUntracked {
			dirty = fmt.Sprintf("tracked=%t untracked=%t", repo.DirtyTracked, repo.DirtyUntracked)
		}
		fmt.Fprintf(&buf, "| `%s` | `%s` | `%s` | `%s` | %d | %s | `%s` |\n", repo.Name, repo.Module, repo.GlazedVersion, repo.CurrentBranch, repo.AheadBase, dirty, strings.Join(repo.LintTargets, ", "))
	}
	fmt.Fprintf(&buf, "\n## Branch checks\n\n")
	for _, br := range branches {
		mark := "✅"
		if !br.OK {
			mark = "⚠️"
		}
		fmt.Fprintf(&buf, "- %s `%s`: %s\n", mark, br.Repo.Name, br.Message)
	}
	fmt.Fprintf(&buf, "\n## Validation commands\n\n")
	if len(cfg.Validation.Commands) == 0 {
		fmt.Fprintf(&buf, "No validation commands configured.\n")
	} else {
		for _, cmd := range cfg.Validation.Commands {
			fmt.Fprintf(&buf, "- `%s`: `%s`\n", cmd.Name, cmd.Run)
		}
		fmt.Fprintf(&buf, "\nLogs: `%s`\n", cfg.Validation.LogDir)
	}
	fmt.Fprintf(&buf, "\n## Pull requests\n\n")
	if cfg.PullRequest.OutputPRs == "" {
		fmt.Fprintf(&buf, "No PR output file configured.\n")
	} else if b, err := os.ReadFile(resolvePath(cfg.Workspace, cfg.PullRequest.OutputPRs)); err == nil {
		fmt.Fprintf(&buf, "PR list from `%s`:\n\n```yaml\n%s\n```\n", cfg.PullRequest.OutputPRs, strings.TrimSpace(string(b)))
	} else {
		fmt.Fprintf(&buf, "PR output file configured but not present yet: `%s`\n", cfg.PullRequest.OutputPRs)
	}
	fmt.Fprintf(&buf, "\n## Next steps\n\n")
	fmt.Fprintf(&buf, "1. Run validation and inspect failed logs.\n")
	fmt.Fprintf(&buf, "2. Verify each branch is focused and based on `%s`.\n", cfg.Base)
	fmt.Fprintf(&buf, "3. Push/open PRs and run `ggg batch ready` on the generated PR list.\n")
	return buf.String(), nil
}

func reposForConfig(cfg Config) ([]Repo, error) {
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	repos := make([]Repo, 0, len(targets))
	for _, target := range targets {
		repo := InspectRepo(target, cfg.Base)
		module, requires, _ := parseGoMod(filepath.Join(target, "go.mod"))
		repo.Module = module
		repo.GlazedVersion = requires["github.com/go-go-golems/glazed"]
		repos = append(repos, repo)
	}
	return repos, nil
}

func resolvePath(base, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
