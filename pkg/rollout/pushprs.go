package rollout

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type PushPROptions struct {
	DryRun       bool
	Yes          bool
	NoVerifyPush bool
	Reason       string
}

type PushPRResult struct {
	Repo    Repo   `json:"repo" yaml:"repo"`
	Action  string `json:"action" yaml:"action"`
	URL     string `json:"url" yaml:"url"`
	OK      bool   `json:"ok" yaml:"ok"`
	Message string `json:"message" yaml:"message"`
}

type prOutputFile struct {
	PRs []string `yaml:"prs"`
}

func PushPRs(ctx context.Context, cfg Config, opts PushPROptions) ([]PushPRResult, error) {
	if !opts.DryRun && !opts.Yes {
		return nil, fmt.Errorf("push-prs requires --yes unless --dry-run is set")
	}
	noVerify := opts.NoVerifyPush || cfg.PullRequest.NoVerifyPush
	if noVerify && strings.TrimSpace(opts.Reason) == "" && !opts.DryRun {
		return nil, fmt.Errorf("--no-verify-push requires --reason")
	}
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	var results []PushPRResult
	var urls []string
	for _, target := range targets {
		repo := InspectRepo(target, cfg.Base)
		res := PushPRResult{Repo: repo}
		if cfg.Branch == "" || repo.CurrentBranch != cfg.Branch {
			res.Action = "skipped_branch_mismatch"
			res.Message = fmt.Sprintf("current branch %s does not match rollout branch %s", repo.CurrentBranch, cfg.Branch)
			results = append(results, res)
			continue
		}
		if repo.AheadBase > 1 {
			res.Action = "skipped_unfocused_branch"
			res.Message = fmt.Sprintf("branch is %d commits ahead of %s", repo.AheadBase, cfg.Base)
			results = append(results, res)
			continue
		}
		if opts.DryRun {
			res.Action = "would_push_and_open_pr"
			res.OK = true
			results = append(results, res)
			continue
		}
		args := []string{"push", "-u", "origin", cfg.Branch}
		if noVerify {
			args = append([]string{"push", "--no-verify", "-u", "origin", cfg.Branch})
		}
		if err := runGit(target, args...); err != nil {
			res.Action = "push_failed"
			res.Message = err.Error()
			results = append(results, res)
			continue
		}
		url, err := createPR(ctx, target, cfg)
		if err != nil {
			res.Action = "pr_failed"
			res.Message = err.Error()
			results = append(results, res)
			continue
		}
		res.Action = "opened"
		res.URL = url
		res.OK = true
		urls = append(urls, url)
		results = append(results, res)
	}
	if len(urls) > 0 && cfg.PullRequest.OutputPRs != "" {
		path := resolvePath(cfg.Workspace, cfg.PullRequest.OutputPRs)
		b, err := yaml.Marshal(prOutputFile{PRs: urls})
		if err != nil {
			return results, err
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return results, err
		}
		if err := os.WriteFile(path, b, 0o644); err != nil {
			return results, err
		}
	}
	return results, nil
}

func createPR(ctx context.Context, dir string, cfg Config) (string, error) {
	args := []string{"pr", "create", "--base", strings.TrimPrefix(cfg.Base, "origin/"), "--head", cfg.Branch, "--title", cfg.PullRequest.Title}
	if cfg.PullRequest.Title == "" {
		args = []string{"pr", "create", "--base", strings.TrimPrefix(cfg.Base, "origin/"), "--head", cfg.Branch, "--title", cfg.CommitMessage}
	}
	if cfg.PullRequest.BodyFile != "" {
		args = append(args, "--body-file", resolvePath(cfg.Workspace, cfg.PullRequest.BodyFile))
	} else {
		args = append(args, "--body", "Automated rollout PR created by ggg rollout push-prs.")
	}
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gh %s failed: %w\n%s", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}
