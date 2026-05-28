package actionstatus

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Options struct {
	Repo            string
	Branch          string
	SHA             string
	Limit           int
	IgnoreWorkflows []string
}

type Run struct {
	Repo           string `json:"repo" yaml:"repo"`
	Branch         string `json:"branch" yaml:"branch"`
	SHA            string `json:"sha" yaml:"sha"`
	Workflow       string `json:"workflow" yaml:"workflow"`
	DisplayTitle   string `json:"display_title" yaml:"display_title"`
	Status         string `json:"status" yaml:"status"`
	Conclusion     string `json:"conclusion" yaml:"conclusion"`
	Classification string `json:"classification" yaml:"classification"`
	Ignored        bool   `json:"ignored" yaml:"ignored"`
	URL            string `json:"url" yaml:"url"`
	CreatedAt      string `json:"created_at" yaml:"created_at"`
}

type Summary struct {
	Repo            string `json:"repo" yaml:"repo"`
	OK              bool   `json:"ok" yaml:"ok"`
	State           string `json:"state" yaml:"state"`
	Total           int    `json:"total" yaml:"total"`
	Success         int    `json:"success" yaml:"success"`
	IgnoredFailures int    `json:"ignored_failures" yaml:"ignored_failures"`
	Failed          int    `json:"failed" yaml:"failed"`
	Pending         int    `json:"pending" yaml:"pending"`
	NoRuns          int    `json:"no_runs" yaml:"no_runs"`
	Other           int    `json:"other" yaml:"other"`
}

type Result struct {
	Runs    []Run   `json:"runs" yaml:"runs"`
	Summary Summary `json:"summary" yaml:"summary"`
}

func Collect(ctx context.Context, opts Options) (Result, error) {
	if opts.Repo == "" {
		return Result{}, fmt.Errorf("repo is required")
	}
	if opts.Limit <= 0 {
		opts.Limit = 20
	}
	shaFilter := opts.SHA
	if shaFilter != "" {
		resolved, err := resolveCommit(ctx, opts.Repo, shaFilter)
		if err != nil {
			return Result{}, err
		}
		shaFilter = resolved
	}
	args := []string{"run", "list", "-R", opts.Repo, "--limit", fmt.Sprintf("%d", opts.Limit), "--json", "databaseId,displayTitle,headSha,status,conclusion,workflowName,createdAt,url"}
	if opts.Branch != "" {
		args = append(args, "--branch", opts.Branch)
	}
	if shaFilter != "" {
		args = append(args, "--commit", shaFilter)
	}
	out, err := run(ctx, "gh", args...)
	if err != nil {
		return Result{}, err
	}
	var ghRuns []ghRun
	if err := json.Unmarshal(out, &ghRuns); err != nil {
		return Result{}, err
	}
	ignore := ignoreSet(opts.IgnoreWorkflows)
	res := Result{Summary: Summary{Repo: opts.Repo}}
	for _, gr := range ghRuns {
		sha := gr.HeadSHA
		classification, ignored := Classify(gr.Status, gr.Conclusion, gr.WorkflowName, ignore)
		r := Run{Repo: opts.Repo, Branch: opts.Branch, SHA: shortSHA(sha), Workflow: gr.WorkflowName, DisplayTitle: gr.DisplayTitle, Status: gr.Status, Conclusion: gr.Conclusion, Classification: classification, Ignored: ignored, URL: gr.URL, CreatedAt: gr.CreatedAt}
		res.Runs = append(res.Runs, r)
		res.Summary.Total++
		switch classification {
		case "ok":
			res.Summary.Success++
		case "ignored_failure":
			res.Summary.IgnoredFailures++
		case "failed":
			res.Summary.Failed++
		case "pending":
			res.Summary.Pending++
		default:
			res.Summary.Other++
		}
	}
	if res.Summary.Total == 0 {
		res.Summary.NoRuns = 1
	}
	res.Summary.State = State(res.Summary)
	res.Summary.OK = res.Summary.State == "ok"
	return res, nil
}

func Classify(status, conclusion, workflow string, ignored map[string]bool) (string, bool) {
	status = strings.ToLower(status)
	conclusion = strings.ToLower(conclusion)
	if status == "queued" || status == "in_progress" || status == "waiting" || status == "requested" || status == "pending" {
		return "pending", false
	}
	if status != "completed" {
		return "pending", false
	}
	switch conclusion {
	case "success", "skipped", "cancelled", "neutral", "":
		return "ok", false
	case "failure", "timed_out", "action_required", "startup_failure":
		if ignored[workflow] || ignored[strings.ToLower(workflow)] {
			return "ignored_failure", true
		}
		return "failed", false
	default:
		return "other", false
	}
}

func State(s Summary) string {
	if s.Failed > 0 {
		return "failed"
	}
	if s.Pending > 0 || s.NoRuns > 0 {
		if s.Total == 0 && s.NoRuns > 0 {
			return "no_runs"
		}
		return "pending"
	}
	if s.Total == 0 {
		return "no_runs"
	}
	return "ok"
}

func ExitCode(s Summary) int {
	switch s.State {
	case "failed":
		return 1
	case "pending", "no_runs":
		return 2
	default:
		return 0
	}
}

type Manifest struct {
	Repos []ManifestRepo `yaml:"repos"`
}

type ManifestRepo struct {
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch"`
	SHA    string `yaml:"sha"`
	Limit  int    `yaml:"limit"`
}

func LoadManifest(path string) (Manifest, error) {
	var m Manifest
	body, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := yaml.Unmarshal(body, &m); err != nil {
		return m, err
	}
	if len(m.Repos) == 0 {
		return m, fmt.Errorf("manifest has no repos")
	}
	return m, nil
}

func BatchCollect(ctx context.Context, manifest Manifest, ignore []string, defaultLimit int) ([]Run, Summary, error) {
	var runs []Run
	summary := Summary{Repo: "summary"}
	for _, repo := range manifest.Repos {
		limit := repo.Limit
		if limit == 0 {
			limit = defaultLimit
		}
		res, err := Collect(ctx, Options{Repo: repo.Repo, Branch: repo.Branch, SHA: repo.SHA, Limit: limit, IgnoreWorkflows: ignore})
		if err != nil {
			return nil, summary, err
		}
		if res.Summary.NoRuns > 0 {
			runs = append(runs, Run{Repo: repo.Repo, Branch: repo.Branch, SHA: repo.SHA, Workflow: "(no matching runs)", Classification: "no_runs"})
		} else {
			runs = append(runs, res.Runs...)
		}
		summary.Total += res.Summary.Total
		summary.Success += res.Summary.Success
		summary.IgnoredFailures += res.Summary.IgnoredFailures
		summary.Failed += res.Summary.Failed
		summary.Pending += res.Summary.Pending
		summary.NoRuns += res.Summary.NoRuns
		summary.Other += res.Summary.Other
	}
	summary.State = State(summary)
	summary.OK = summary.State == "ok"
	return runs, summary, nil
}

func resolveCommit(ctx context.Context, repo, sha string) (string, error) {
	if len(sha) == 40 {
		return sha, nil
	}
	out, err := run(ctx, "gh", "api", "repos/"+repo+"/commits/"+sha, "--jq", ".sha")
	if err != nil {
		return "", err
	}
	resolved := strings.TrimSpace(string(out))
	if resolved == "" {
		return "", fmt.Errorf("could not resolve commit %q in %s", sha, repo)
	}
	return resolved, nil
}

func run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

type ghRun struct {
	HeadSHA      string `json:"headSha"`
	WorkflowName string `json:"workflowName"`
	DisplayTitle string `json:"displayTitle"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	URL          string `json:"url"`
	CreatedAt    string `json:"createdAt"`
}

func ignoreSet(values []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				m[part] = true
				m[strings.ToLower(part)] = true
			}
		}
	}
	return m
}

func shortSHA(s string) string {
	if len(s) > 7 {
		return s[:7]
	}
	return s
}

func Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
