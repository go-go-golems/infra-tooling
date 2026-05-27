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
)

type BumpMode string

const (
	Patch BumpMode = "patch"
	Minor BumpMode = "minor"
	Major BumpMode = "major"
)

type Result struct {
	RepoDir  string
	Mode     BumpMode
	Module   string
	Tag      string
	Commit   string
	Verified bool
}

func Tag(ctx context.Context, repoDir string, mode BumpMode, dryRun bool) (Result, error) {
	if repoDir == "" {
		repoDir = "."
	}
	module, err := ModulePath(repoDir)
	if err != nil {
		return Result{}, err
	}
	if _, err := run(ctx, repoDir, "git", "fetch", "origin", "main", "--tags"); err != nil {
		return Result{}, err
	}
	commitBytes, err := run(ctx, repoDir, "git", "rev-parse", "origin/main")
	if err != nil {
		return Result{}, err
	}
	commit := strings.TrimSpace(string(commitBytes))
	if !dryRun {
		if _, err := run(ctx, repoDir, "git", "checkout", "--detach", "origin/main"); err != nil {
			return Result{}, err
		}
	}
	tagBytes, err := run(ctx, repoDir, "svu", string(mode))
	if err != nil {
		return Result{}, err
	}
	tag := strings.TrimSpace(string(tagBytes))
	if tag == "" {
		return Result{}, fmt.Errorf("svu %s returned empty tag", mode)
	}
	res := Result{RepoDir: repoDir, Mode: mode, Module: module, Tag: tag, Commit: commit}
	if dryRun {
		return res, nil
	}
	if _, err := run(ctx, repoDir, "git", "tag", tag); err != nil {
		return Result{}, err
	}
	if _, err := run(ctx, repoDir, "git", "push", "origin", "refs/tags/"+tag); err != nil {
		return Result{}, err
	}
	if _, err := runWithEnv(ctx, repoDir, []string{"GOWORK=off", "GOPROXY=proxy.golang.org"}, "go", "list", "-m", module+"@"+tag); err != nil {
		return Result{}, err
	}
	res.Verified = true
	return res, nil
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
