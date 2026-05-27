package rollout

import (
	"fmt"
	"os/exec"
	"strings"
)

type BranchOptions struct {
	Commit bool
	Yes    bool
}

type BranchResult struct {
	Repo           Repo   `json:"repo" yaml:"repo"`
	ExpectedBranch string `json:"expected_branch" yaml:"expected_branch"`
	ExpectedBase   string `json:"expected_base" yaml:"expected_base"`
	OK             bool   `json:"ok" yaml:"ok"`
	Message        string `json:"message" yaml:"message"`
}

func BranchStatus(cfg Config) ([]BranchResult, error) {
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	var results []BranchResult
	for _, target := range targets {
		repo := InspectRepo(target, cfg.Base)
		module, requires, _ := parseGoMod(target + "/go.mod")
		repo.Module = module
		repo.GlazedVersion = requires["github.com/go-go-golems/glazed"]
		res := BranchResult{Repo: repo, ExpectedBranch: cfg.Branch, ExpectedBase: cfg.Base}
		problems := []string{}
		if cfg.Branch != "" && repo.CurrentBranch != cfg.Branch {
			problems = append(problems, fmt.Sprintf("branch is %s, expected %s", repo.CurrentBranch, cfg.Branch))
		}
		if repo.AheadBase > 1 {
			problems = append(problems, fmt.Sprintf("%d commits ahead of %s", repo.AheadBase, cfg.Base))
		}
		if repo.DirtyTracked {
			problems = append(problems, "tracked changes present")
		}
		if len(problems) == 0 {
			res.OK = true
			res.Message = "branch state matches rollout expectations"
		} else {
			res.Message = strings.Join(problems, "; ")
		}
		results = append(results, res)
	}
	return results, nil
}

func CommitTargets(cfg Config, yes bool) ([]BranchResult, error) {
	if !yes {
		return nil, fmt.Errorf("commit requires --yes")
	}
	if cfg.Branch == "" {
		return nil, fmt.Errorf("rollout config is missing branch")
	}
	if cfg.CommitMessage == "" {
		return nil, fmt.Errorf("rollout config is missing commit_message")
	}
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	var results []BranchResult
	for _, target := range targets {
		if err := runGit(target, "checkout", "-B", cfg.Branch, cfg.Base); err != nil {
			return results, err
		}
		if err := runGit(target, "add", "Makefile", ".github/workflows"); err != nil {
			return results, err
		}
		if hasStagedChanges(target) {
			_ = runGit(target, "commit", "-m", cfg.CommitMessage)
		}
		repo := InspectRepo(target, cfg.Base)
		results = append(results, BranchResult{Repo: repo, ExpectedBranch: cfg.Branch, ExpectedBase: cfg.Base, OK: repo.CurrentBranch == cfg.Branch && repo.AheadBase <= 1, Message: "commit attempted"})
	}
	return results, nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s in %s failed: %w\n%s", strings.Join(args, " "), dir, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func hasStagedChanges(dir string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	cmd.Dir = dir
	err := cmd.Run()
	return err != nil
}
