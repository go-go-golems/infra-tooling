package ghclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

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

const codexStatusQuery = `query($owner:String!,$repo:String!,$number:Int!){ repository(owner:$owner,name:$repo){ pullRequest(number:$number){ reviews(last:50){nodes{author{login} submittedAt url reactionGroups{content users(first:20){totalCount}}}} comments(last:50){nodes{author{login} body createdAt url reactionGroups{content users(first:20){totalCount}}}}}}}`

func (c Client) CodexStatus(ctx context.Context, ref prref.Ref) (CodexStatus, error) {
	if err := ref.Validate(); err != nil {
		return CodexStatus{}, err
	}
	out, err := run(ctx, "gh", "api", "graphql", "-f", "owner="+ref.Owner, "-f", "repo="+ref.Repo, "-F", fmt.Sprintf("number=%d", ref.Number), "-f", "query="+codexStatusQuery)
	if err != nil {
		return CodexStatus{}, err
	}
	var decoded graphQLResponse
	if err := json.Unmarshal(out, &decoded); err != nil {
		return CodexStatus{}, err
	}
	pr := decoded.Data.Repository.PullRequest
	var latest signal
	for _, n := range pr.Reviews.Nodes {
		login := n.Author.Login
		if isCodex(login) {
			s := signal{Kind: "review", Author: login, URL: n.URL, Time: n.SubmittedAt, Eyes: reactionCount(n.ReactionGroups, "EYES"), ThumbsUp: reactionCount(n.ReactionGroups, "THUMBS_UP")}
			if s.Time >= latest.Time {
				latest = s
			}
		}
	}
	for _, n := range pr.Comments.Nodes {
		login := n.Author.Login
		trigger := strings.TrimSpace(n.Body) == "@codex review"
		if isCodex(login) || trigger {
			kind := "comment"
			if trigger && !isCodex(login) {
				kind = "codex-trigger"
			}
			s := signal{Kind: kind, Author: login, URL: n.URL, Time: n.CreatedAt, Eyes: reactionCount(n.ReactionGroups, "EYES"), ThumbsUp: reactionCount(n.ReactionGroups, "THUMBS_UP")}
			if s.Time >= latest.Time {
				latest = s
			}
		}
	}
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

type signal struct {
	Kind     string
	Author   string
	URL      string
	Time     string
	Eyes     int
	ThumbsUp int
}

type graphQLResponse struct {
	Data struct {
		Repository struct {
			PullRequest struct {
				Reviews  struct{ Nodes []reviewNode }  `json:"reviews"`
				Comments struct{ Nodes []commentNode } `json:"comments"`
			} `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}

type reviewNode struct {
	Author         author          `json:"author"`
	SubmittedAt    string          `json:"submittedAt"`
	URL            string          `json:"url"`
	ReactionGroups []reactionGroup `json:"reactionGroups"`
}

type commentNode struct {
	Author         author          `json:"author"`
	Body           string          `json:"body"`
	CreatedAt      string          `json:"createdAt"`
	URL            string          `json:"url"`
	ReactionGroups []reactionGroup `json:"reactionGroups"`
}

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
