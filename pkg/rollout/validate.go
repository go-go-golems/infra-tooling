package rollout

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ValidationOptions struct {
	DryRun bool
}

type ValidationResult struct {
	Repo     Repo   `json:"repo" yaml:"repo"`
	Command  string `json:"command" yaml:"command"`
	Run      string `json:"run" yaml:"run"`
	ExitCode int    `json:"exit_code" yaml:"exit_code"`
	LogPath  string `json:"log_path" yaml:"log_path"`
	OK       bool   `json:"ok" yaml:"ok"`
	DryRun   bool   `json:"dry_run" yaml:"dry_run"`
}

func Validate(ctx context.Context, cfg Config, opts ValidationOptions) ([]ValidationResult, error) {
	if len(cfg.Validation.Commands) == 0 {
		return nil, fmt.Errorf("rollout config has no validation commands")
	}
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	logDir := cfg.Validation.LogDir
	if logDir == "" {
		logDir = filepath.Join(cfg.Workspace, ".ggg-rollout-logs")
	}
	if !filepath.IsAbs(logDir) {
		logDir = filepath.Join(cfg.Workspace, logDir)
	}
	var results []ValidationResult
	for _, target := range targets {
		repo := InspectRepo(target, cfg.Base)
		module, requires, _ := parseGoMod(filepath.Join(target, "go.mod"))
		repo.Module = module
		repo.GlazedVersion = requires["github.com/go-go-golems/glazed"]
		for _, command := range cfg.Validation.Commands {
			result := ValidationResult{Repo: repo, Command: command.Name, Run: command.Run, DryRun: opts.DryRun}
			if command.Name == "" {
				result.Command = sanitizeCommandName(command.Run)
			}
			result.LogPath = filepath.Join(logDir, repo.Name+"-"+result.Command+".log")
			if opts.DryRun {
				result.OK = true
				results = append(results, result)
				continue
			}
			if err := os.MkdirAll(filepath.Dir(result.LogPath), 0o755); err != nil {
				return results, err
			}
			code, runErr := runShell(ctx, target, command.Run, result.LogPath)
			result.ExitCode = code
			result.OK = runErr == nil && code == 0
			results = append(results, result)
			if !result.OK && !cfg.Validation.ContinueOnError {
				return results, fmt.Errorf("validation failed for %s: %s (exit %d, log %s)", repo.Name, command.Run, code, result.LogPath)
			}
		}
	}
	return results, nil
}

func runShell(ctx context.Context, dir, command, logPath string) (int, error) {
	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	cmd.Dir = dir
	log, err := os.Create(logPath)
	if err != nil {
		return 1, err
	}
	defer log.Close()
	_, _ = fmt.Fprintf(log, "$ cd %s\n$ %s\n# started: %s\n\n", dir, command, time.Now().Format(time.RFC3339))
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	_, _ = fmt.Fprintf(log, "\n# finished: %s\n", time.Now().Format(time.RFC3339))
	if err == nil {
		return 0, nil
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), err
	}
	return 1, err
}

func sanitizeCommandName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "command"
	}
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else if b.Len() > 0 && b.String()[b.Len()-1] != '-' {
			b.WriteByte('-')
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "command"
	}
	if len(out) > 40 {
		out = out[:40]
	}
	return out
}
