package ghclient

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
)

const readinessQuery = `query($owner: String!, $repo: String!, $number: Int!) { repository(owner: $owner, name: $repo) { pullRequest(number: $number) { url number title mergeStateStatus reviewDecision headRefOid statusCheckRollup { contexts(first: 100) { nodes { __typename ... on CheckRun { name status conclusion detailsUrl } ... on StatusContext { context state targetUrl } } } } reviews(last: 100) { nodes { author { login } state body submittedAt url reactionGroups { content users(first: 20) { totalCount } } comments(first: 100) { nodes { path line body url } } } } comments(last: 100) { nodes { author { login } body createdAt url reactionGroups { content users(first: 20) { totalCount } } } } } } }`

func (c Client) Readiness(ctx context.Context, ref prref.Ref) (prready.Report, error) {
	if err := ref.Validate(); err != nil {
		return prready.Report{}, err
	}
	out, err := run(ctx, "gh", "api", "graphql", "-f", "owner="+ref.Owner, "-f", "repo="+ref.Repo, "-F", fmt.Sprintf("number=%d", ref.Number), "-f", "query="+readinessQuery)
	if err != nil {
		return prready.Report{}, err
	}
	var decoded readinessResponse
	if err := json.Unmarshal(out, &decoded); err != nil {
		return prready.Report{}, err
	}
	pr := decoded.Data.Repository.PullRequest
	snap := prready.Snapshot{PR: ref, URL: pr.URL, MergeStateStatus: pr.MergeStateStatus, ReviewDecision: pr.ReviewDecision, HeadRefOID: pr.HeadRefOID}
	for _, n := range pr.StatusCheckRollup.Contexts.Nodes {
		check := prready.Check{Kind: n.TypeName, Status: n.Status, Conclusion: n.Conclusion, State: n.State}
		if n.TypeName == "CheckRun" {
			check.Name = n.Name
			check.URL = n.DetailsURL
		} else {
			check.Name = n.Context
			check.URL = n.TargetURL
		}
		snap.Checks = append(snap.Checks, check)
	}
	for _, n := range pr.Reviews.Nodes {
		login := n.Author.Login
		if !isCodex(login) {
			continue
		}
		sig := prready.CodexSignal{Kind: "review", Author: login, URL: n.URL, Time: n.SubmittedAt, Body: n.Body, CodexAuthored: true, Eyes: reactionCount(n.ReactionGroups, "EYES"), ThumbsUp: reactionCount(n.ReactionGroups, "THUMBS_UP")}
		for _, c := range n.Comments.Nodes {
			if strings.TrimSpace(c.Body) != "" {
				sig.Comments = append(sig.Comments, prready.ReviewComment{Path: c.Path, Line: c.Line, Body: c.Body, URL: c.URL})
			}
		}
		snap.Signals = append(snap.Signals, sig)
	}
	for _, n := range pr.Comments.Nodes {
		login := n.Author.Login
		trigger := isCodexTrigger(n.Body)
		if !isCodex(login) && !trigger {
			continue
		}
		kind := "comment"
		authored := isCodex(login)
		if trigger && !authored {
			kind = "codex-trigger"
		}
		snap.Signals = append(snap.Signals, prready.CodexSignal{Kind: kind, Author: login, URL: n.URL, Time: n.CreatedAt, Body: n.Body, CodexAuthored: authored, Eyes: reactionCount(n.ReactionGroups, "EYES"), ThumbsUp: reactionCount(n.ReactionGroups, "THUMBS_UP")})
	}
	return prready.Classify(snap), nil
}

var codexTriggerRE = regexp.MustCompile(`(?im)^\s*@codex\s+review\s*$`)

func isCodexTrigger(body string) bool { return codexTriggerRE.MatchString(body) }

type readinessResponse struct {
	Data struct {
		Repository struct {
			PullRequest readinessPR `json:"pullRequest"`
		} `json:"repository"`
	} `json:"data"`
}
type readinessPR struct {
	URL               string `json:"url"`
	MergeStateStatus  string `json:"mergeStateStatus"`
	ReviewDecision    string `json:"reviewDecision"`
	HeadRefOID        string `json:"headRefOid"`
	StatusCheckRollup struct {
		Contexts struct {
			Nodes []checkNode `json:"nodes"`
		} `json:"contexts"`
	} `json:"statusCheckRollup"`
	Reviews struct {
		Nodes []readinessReviewNode `json:"nodes"`
	} `json:"reviews"`
	Comments struct {
		Nodes []readinessCommentNode `json:"nodes"`
	} `json:"comments"`
}
type checkNode struct {
	TypeName   string `json:"__typename"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	DetailsURL string `json:"detailsUrl"`
	Context    string `json:"context"`
	State      string `json:"state"`
	TargetURL  string `json:"targetUrl"`
}
type readinessReviewNode struct {
	Author         author          `json:"author"`
	Body           string          `json:"body"`
	SubmittedAt    string          `json:"submittedAt"`
	URL            string          `json:"url"`
	ReactionGroups []reactionGroup `json:"reactionGroups"`
	Comments       struct {
		Nodes []readinessReviewComment `json:"nodes"`
	} `json:"comments"`
}
type readinessReviewComment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Body string `json:"body"`
	URL  string `json:"url"`
}
type readinessCommentNode struct {
	Author         author          `json:"author"`
	Body           string          `json:"body"`
	CreatedAt      string          `json:"createdAt"`
	URL            string          `json:"url"`
	ReactionGroups []reactionGroup `json:"reactionGroups"`
}
