package release

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BumpMode string

const (
	Patch BumpMode = "patch"
	Minor BumpMode = "minor"
	Major BumpMode = "major"
)

type Options struct {
	RepoDir        string
	Mode           BumpMode
	DryRun         bool
	AllowDirty     bool
	Target         string
	Yes            bool
	VerifyTimeout  time.Duration
	VerifyInterval time.Duration
}

type Result struct {
	RepoDir     string
	Mode        BumpMode
	Module      string
	CurrentTag  string
	Tag         string
	Commit      string
	Target      string
	Dirty       bool
	ExistingTag bool
	Verified    bool
	Plan        []string
}

func Tag(ctx context.Context, repoDir string, mode BumpMode, dryRun bool) (Result, error) {
	return TagWithOptions(ctx, Options{RepoDir: repoDir, Mode: mode, DryRun: dryRun, Target: "origin/main"})
}

func TagWithOptions(ctx context.Context, opts Options) (Result, error) {
	if opts.RepoDir == "" {
		opts.RepoDir = "."
	}
	if opts.Target == "" {
		opts.Target = "origin/main"
	}
	if opts.VerifyTimeout <= 0 {
		opts.VerifyTimeout = 2 * time.Minute
	}
	if opts.VerifyInterval <= 0 {
		opts.VerifyInterval = 10 * time.Second
	}
	module, err := ModulePath(opts.RepoDir)
	if err != nil {
		return Result{}, err
	}
	if _, err := run(ctx, opts.RepoDir, "git", "fetch", "origin", "main", "--tags"); err != nil {
		return Result{}, err
	}
	dirty, err := Dirty(ctx, opts.RepoDir)
	if err != nil {
		return Result{}, err
	}
	if dirty && !opts.AllowDirty {
		return Result{}, fmt.Errorf("worktree is dirty; use --allow-dirty to override")
	}
	commitBytes, err := run(ctx, opts.RepoDir, "git", "rev-parse", opts.Target)
	if err != nil {
		return Result{}, err
	}
	commit := strings.TrimSpace(string(commitBytes))
	currentTag := ""
	if b, err := run(ctx, opts.RepoDir, "svu", "current"); err == nil {
		currentTag = strings.TrimSpace(string(b))
	}
	tagBytes, err := run(ctx, opts.RepoDir, "svu", string(opts.Mode))
	if err != nil {
		return Result{}, err
	}
	tag := strings.TrimSpace(string(tagBytes))
	if tag == "" {
		return Result{}, fmt.Errorf("svu %s returned empty tag", opts.Mode)
	}
	res := Result{RepoDir: opts.RepoDir, Mode: opts.Mode, Module: module, CurrentTag: currentTag, Tag: tag, Commit: commit, Target: opts.Target, Dirty: dirty, Plan: []string{"git fetch origin main --tags", "git tag " + tag + " " + commit, "git push origin refs/tags/" + tag, "GOPROXY=proxy.golang.org go list -m " + module + "@" + tag}}
	if exists, existingCommit, err := tagExists(ctx, opts.RepoDir, tag); err != nil {
		return Result{}, err
	} else if exists {
		res.ExistingTag = true
		if existingCommit != commit {
			return res, fmt.Errorf("tag %s already exists at %s, expected %s", tag, existingCommit, commit)
		}
		res.Verified = true
		return res, nil
	}
	if opts.DryRun {
		return res, nil
	}
	if !opts.Yes {
		return res, fmt.Errorf("refusing to push tag without --yes (planned tag %s at %s)", tag, commit)
	}
	if _, err := run(ctx, opts.RepoDir, "git", "tag", tag, commit); err != nil {
		return Result{}, err
	}
	if _, err := run(ctx, opts.RepoDir, "git", "push", "origin", "refs/tags/"+tag); err != nil {
		return Result{}, err
	}
	if err := VerifyProxy(ctx, opts.RepoDir, module, tag, opts.VerifyTimeout, opts.VerifyInterval); err != nil {
		return Result{}, err
	}
	res.Verified = true
	return res, nil
}

func Dirty(ctx context.Context, repoDir string) (bool, error) {
	out, err := run(ctx, repoDir, "git", "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) != "", nil
}

func VerifyProxy(ctx context.Context, repoDir, module, tag string, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for {
		_, err := runWithEnv(ctx, repoDir, []string{"GOWORK=off", "GOPROXY=proxy.golang.org"}, "go", "list", "-m", module+"@"+tag)
		if err == nil {
			return nil
		}
		last = err
		if time.Now().After(deadline) {
			return last
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

func tagExists(ctx context.Context, repoDir, tag string) (bool, string, error) {
	out, err := run(ctx, repoDir, "git", "rev-parse", "--verify", "refs/tags/"+tag)
	if err != nil {
		return false, "", nil
	}
	return true, strings.TrimSpace(string(out)), nil
}

func ModulePath(repoDir string) (string, error) {
	f, err := os.Open(filepath.Join(repoDir, "go.mod"))
	if err != nil {
		return "", err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if strings.HasPrefix(line, "module ") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], nil
			}
		}
	}
	if err := s.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module path not found in %s", filepath.Join(repoDir, "go.mod"))
}

func run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	return runWithEnv(ctx, dir, nil, name, args...)
}

func runWithEnv(ctx context.Context, dir string, env []string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}
