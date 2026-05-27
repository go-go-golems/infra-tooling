package ghclient

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
)

type Client struct{}

type CodexStatus struct {
	SignalURL  string
	SignalKind string
	Author     string
	Eyes       int
	ThumbsUp   int
	Running    bool
}

func (c Client) CodexStatus(ctx context.Context, ref prref.Ref) (CodexStatus, error) {
	snap, err := c.Snapshot(ctx, ref)
	if err != nil {
		return CodexStatus{}, err
	}
	latest, _ := prready.LatestSignal(snap)
	return CodexStatus{SignalURL: latest.URL, SignalKind: latest.Kind, Author: latest.Author, Eyes: latest.Eyes, ThumbsUp: latest.ThumbsUp, Running: latest.Eyes > 0}, nil
}

func (c Client) TriggerCodex(ctx context.Context, ref prref.Ref) (string, error) {
	out, err := run(ctx, "gh", "pr", "comment", ref.URL(), "--body", "@codex review")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
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

var codexRE = regexp.MustCompile(`(?i)(^|[-_])(codex|openai-codex|chatgpt)([-_]|$)|codex|openai`)

func isCodex(login string) bool { return codexRE.MatchString(login) }

type author struct {
	Login string `json:"login"`
}

type reactionGroup struct {
	Content string `json:"content"`
	Users   struct {
		TotalCount int `json:"totalCount"`
	} `json:"users"`
}

func reactionCount(groups []reactionGroup, content string) int {
	for _, g := range groups {
		if g.Content == content {
			return g.Users.TotalCount
		}
	}
	return 0
}
