package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type watchSettings struct {
	Repo         string
	Workflow     string
	Tag          string
	Interval     time.Duration
	Timeout      time.Duration
	StartTimeout time.Duration
	Output       string
	NoStream     bool
	VerifyDocs   bool
	Package      string
	BaseURL      string
}

type releaseRun struct {
	DatabaseID   int    `json:"databaseId" yaml:"database_id"`
	Status       string `json:"status" yaml:"status"`
	Conclusion   string `json:"conclusion" yaml:"conclusion"`
	HeadBranch   string `json:"headBranch" yaml:"head_branch"`
	HeadSha      string `json:"headSha" yaml:"head_sha"`
	Event        string `json:"event" yaml:"event"`
	DisplayTitle string `json:"displayTitle" yaml:"display_title"`
	CreatedAt    string `json:"createdAt" yaml:"created_at"`
	URL          string `json:"url" yaml:"url"`
}

type watchResult struct {
	OK               bool              `json:"ok" yaml:"ok"`
	Repository       string            `json:"repository" yaml:"repository"`
	Workflow         string            `json:"workflow" yaml:"workflow"`
	Tag              string            `json:"tag" yaml:"tag"`
	Run              releaseRun        `json:"run" yaml:"run"`
	Docs             *verifyDocsResult `json:"docs,omitempty" yaml:"docs,omitempty"`
	Error            string            `json:"error,omitempty" yaml:"error,omitempty"`
	FailedLogCommand string            `json:"failed_log_command,omitempty" yaml:"failed_log_command,omitempty"`
	DispatchCommand  string            `json:"dispatch_command,omitempty" yaml:"dispatch_command,omitempty"`
	ElapsedSecs      int               `json:"elapsed_seconds" yaml:"elapsed_seconds"`
}

func newWatchCommand() *cobra.Command {
	s := &watchSettings{Workflow: "release.yaml", Interval: 30 * time.Second, Timeout: 30 * time.Minute, StartTimeout: 2 * time.Minute, Output: "table", BaseURL: "https://docs.yolo.scapegoat.dev"}
	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Watch a tag-triggered workflow and optionally verify docs",
		RunE: func(cmd *cobra.Command, args []string) error {
			res := watchRelease(cmd.Context(), s)
			if err := writeWatchResult(res, s.Output); err != nil {
				return err
			}
			if !res.OK {
				return fmt.Errorf("release watch failed for %s %s", s.Repo, s.Tag)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&s.Repo, "repo", "", "GitHub repository owner/name")
	cmd.Flags().StringVar(&s.Workflow, "workflow", s.Workflow, "Workflow file name, for example release.yaml or publish-docs.yaml")
	cmd.Flags().StringVar(&s.Tag, "tag", "", "Tag/head branch to watch")
	cmd.Flags().DurationVar(&s.Interval, "interval", s.Interval, "Polling interval")
	cmd.Flags().DurationVar(&s.Timeout, "timeout", s.Timeout, "Overall timeout")
	cmd.Flags().DurationVar(&s.StartTimeout, "start-timeout", s.StartTimeout, "How long to wait for a matching workflow run to appear before returning a dispatch hint")
	cmd.Flags().StringVar(&s.Output, "output", s.Output, "Output format: table, json, yaml")
	cmd.Flags().BoolVar(&s.NoStream, "no-stream", false, "Poll without running gh run watch")
	cmd.Flags().BoolVar(&s.VerifyDocs, "verify-docs", false, "Verify docs browser after the release succeeds")
	cmd.Flags().StringVar(&s.Package, "package", "", "Docs package name; defaults to repository basename")
	cmd.Flags().StringVar(&s.BaseURL, "base-url", s.BaseURL, "Docs browser base URL")
	_ = cmd.MarkFlagRequired("repo")
	_ = cmd.MarkFlagRequired("tag")
	return cmd
}

func watchRelease(ctx context.Context, s *watchSettings) watchResult {
	started := time.Now()
	res := watchResult{Repository: s.Repo, Workflow: s.Workflow, Tag: s.Tag}
	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()
	run, err := waitForReleaseRun(ctx, s)
	if err != nil {
		res.Error = err.Error()
		res.DispatchCommand = fmt.Sprintf("gh workflow run %s --repo %s --ref %s", s.Workflow, s.Repo, s.Tag)
		res.ElapsedSecs = int(time.Since(started).Seconds())
		return res
	}
	res.Run = run
	if !s.NoStream {
		_ = streamGhRunWatch(ctx, s.Repo, run.DatabaseID, s.Interval)
		run, err = getReleaseRun(ctx, s.Repo, run.DatabaseID)
	} else {
		run, err = waitForReleaseRunCompletion(ctx, s.Repo, run.DatabaseID, s.Interval)
	}
	if err != nil {
		res.Error = err.Error()
		res.ElapsedSecs = int(time.Since(started).Seconds())
		return res
	}
	res.Run = run
	res.OK = run.Status == "completed" && run.Conclusion == "success"
	if run.Status == "completed" && run.Conclusion != "success" {
		res.FailedLogCommand = fmt.Sprintf("gh run view %d --repo %s --log-failed", run.DatabaseID, s.Repo)
	}
	if res.OK && s.VerifyDocs {
		pkg := s.Package
		if pkg == "" {
			_, pkg, _ = strings.Cut(s.Repo, "/")
		}
		docs := verifyDocs(ctx, &verifyDocsSettings{Package: pkg, Version: s.Tag, BaseURL: s.BaseURL, MinSections: 1, Timeout: 30 * time.Second})
		res.Docs = &docs
		res.OK = docs.OK
		if !docs.OK {
			res.Error = docs.Error
		}
	}
	res.ElapsedSecs = int(time.Since(started).Seconds())
	return res
}

func waitForReleaseRun(ctx context.Context, s *watchSettings) (releaseRun, error) {
	startTimeout := s.StartTimeout
	if startTimeout <= 0 || startTimeout > s.Timeout {
		startTimeout = s.Timeout
	}
	startCtx, cancel := context.WithTimeout(ctx, startTimeout)
	defer cancel()
	for {
		runs, err := listReleaseRuns(ctx, s.Repo, s.Workflow, s.Tag)
		if err != nil {
			return releaseRun{}, err
		}
		if len(runs) > 0 {
			return runs[0], nil
		}
		select {
		case <-ctx.Done():
			return releaseRun{}, ctx.Err()
		case <-startCtx.Done():
			return releaseRun{}, fmt.Errorf("no workflow run found for %s on %s within %s; the workflow may be manual-only or missing a tag trigger", s.Tag, s.Workflow, startTimeout)
		case <-time.After(s.Interval):
		}
	}
}

func listReleaseRuns(ctx context.Context, repo, workflow, tag string) ([]releaseRun, error) {
	out, err := runGh(ctx, "run", "list", "--repo", repo, "--workflow", workflow, "--branch", tag, "--limit", "5", "--json", "databaseId,status,conclusion,headBranch,headSha,event,displayTitle,createdAt,url")
	if err != nil {
		return nil, err
	}
	var runs []releaseRun
	if err := json.Unmarshal(out, &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

func getReleaseRun(ctx context.Context, repo string, id int) (releaseRun, error) {
	out, err := runGh(ctx, "run", "view", strconv.Itoa(id), "--repo", repo, "--json", "databaseId,status,conclusion,headBranch,headSha,event,displayTitle,createdAt,url")
	if err != nil {
		return releaseRun{}, err
	}
	var run releaseRun
	if err := json.Unmarshal(out, &run); err != nil {
		return releaseRun{}, err
	}
	return run, nil
}

func waitForReleaseRunCompletion(ctx context.Context, repo string, id int, interval time.Duration) (releaseRun, error) {
	for {
		run, err := getReleaseRun(ctx, repo, id)
		if err != nil {
			return releaseRun{}, err
		}
		if run.Status == "completed" {
			return run, nil
		}
		select {
		case <-ctx.Done():
			return releaseRun{}, ctx.Err()
		case <-time.After(interval):
		}
	}
}

func streamGhRunWatch(ctx context.Context, repo string, id int, interval time.Duration) error {
	cmd := exec.CommandContext(ctx, "gh", "run", "watch", strconv.Itoa(id), "--repo", repo, "--interval", strconv.Itoa(max(1, int(interval.Seconds()))), "--exit-status")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGh(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh %s failed: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

func writeWatchResult(res watchResult, output string) error {
	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	case "yaml":
		b, err := yaml.Marshal(res)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	default:
		fmt.Printf("ok\trepo\ttag\tconclusion\trun_url\n")
		fmt.Printf("%t\t%s\t%s\t%s\t%s\n", res.OK, res.Repository, res.Tag, res.Run.Conclusion, res.Run.URL)
		if res.Docs != nil {
			fmt.Printf("docs\t%s\t%s\t%d\t%s\n", res.Docs.Package, res.Docs.Version, res.Docs.SectionCount, res.Docs.URL)
		}
		if res.FailedLogCommand != "" {
			fmt.Fprintf(os.Stderr, "failed logs: %s\n", res.FailedLogCommand)
		}
		if res.DispatchCommand != "" {
			fmt.Fprintf(os.Stderr, "dispatch manually: %s\n", res.DispatchCommand)
		}
		if res.Error != "" {
			fmt.Fprintf(os.Stderr, "error: %s\n", res.Error)
		}
		return nil
	}
}
