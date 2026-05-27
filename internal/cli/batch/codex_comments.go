package batch

import (
	"context"
	"regexp"
	"sort"
	"strconv"
	"strings"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prlist"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"github.com/spf13/cobra"
)

type codexCommentsCommand struct{ *cmds.CommandDescription }

type codexCommentsSettings struct {
	File           string `glazed:"file"`
	FullBody       bool   `glazed:"full-body"`
	CurrentHead    bool   `glazed:"current-head"`
	GroupByMessage bool   `glazed:"group-by-message"`
}

type batchCodexComment struct {
	PR             string
	Repository     string
	Number         int
	SignalURL      string
	Kind           string
	Author         string
	ReviewedCommit string
	CurrentHead    bool
	Truncated      bool
	Path           string
	Line           int
	Body           string
	URL            string
	Title          string
}

func newCodexCommentsCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &codexCommentsCommand{CommandDescription: cmds.NewCommandDescription(
		"codex-comments",
		cmds.WithShort("List or group Codex-authored comments for a YAML PR list"),
		cmds.WithLong(`List Codex-authored review comments for every PR in a YAML list.

Use --group-by-message to collapse repeated feedback across rollout PRs.`),
		cmds.WithArguments(fields.New("file", fields.TypeString, fields.WithHelp("YAML PR list"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("full-body", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Emit full comment bodies instead of previews")),
			fields.New("current-head", fields.TypeBool, fields.WithDefault(true), fields.WithHelp("Only include comments that apply to the current PR head")),
			fields.New("group-by-message", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Group comments by normalized Codex message title")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *codexCommentsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &codexCommentsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	refs, err := prlist.Load(s.File)
	if err != nil {
		return err
	}
	comments, err := collectCodexComments(ctx, ghclient.Client{}, refs, s.CurrentHead)
	if err != nil {
		return err
	}
	if s.GroupByMessage {
		return emitGroupedCodexComments(ctx, gp, comments)
	}
	for _, comment := range comments {
		if err := gp.AddRow(ctx, codexCommentRow(comment, s.FullBody)); err != nil {
			return err
		}
	}
	return nil
}

func collectCodexComments(ctx context.Context, client ghclient.Client, refs []prref.Ref, currentHeadOnly bool) ([]batchCodexComment, error) {
	var out []batchCodexComment
	for _, ref := range refs {
		snap, err := client.Snapshot(ctx, ref)
		if err != nil {
			return nil, err
		}
		for _, sig := range prready.SortedSignals(snap) {
			if !sig.CodexAuthored {
				continue
			}
			reviewed := prready.ReviewedCommit(sig.Body)
			current := prready.SignalReviewedCurrentHead(sig, snap.HeadRefOID)
			if currentHeadOnly && !current {
				continue
			}
			truncated := sig.CommentsTruncated || snap.ReviewsTruncated || snap.CommentsTruncated
			if len(sig.Comments) == 0 {
				body := sig.Body
				out = append(out, batchCodexComment{PR: ref.URL(), Repository: ref.Repository(), Number: ref.Number, SignalURL: sig.URL, Kind: sig.Kind, Author: sig.Author, ReviewedCommit: reviewed, CurrentHead: current, Truncated: truncated, Body: body, URL: sig.URL, Title: codexMessageTitle(body)})
				continue
			}
			for _, c := range sig.Comments {
				out = append(out, batchCodexComment{PR: ref.URL(), Repository: ref.Repository(), Number: ref.Number, SignalURL: sig.URL, Kind: sig.Kind, Author: sig.Author, ReviewedCommit: reviewed, CurrentHead: current, Truncated: truncated, Path: c.Path, Line: c.Line, Body: c.Body, URL: c.URL, Title: codexMessageTitle(c.Body)})
			}
		}
	}
	return out, nil
}

func emitGroupedCodexComments(ctx context.Context, gp middlewares.Processor, comments []batchCodexComment) error {
	groups := map[string][]batchCodexComment{}
	for _, c := range comments {
		groups[c.Title] = append(groups[c.Title], c)
	}
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		items := groups[key]
		sort.Slice(items, func(i, j int) bool {
			if items[i].Repository == items[j].Repository {
				return items[i].Number < items[j].Number
			}
			return items[i].Repository < items[j].Repository
		})
		prs := make([]string, 0, len(items))
		locations := make([]string, 0, len(items))
		urls := make([]string, 0, len(items))
		seenPR := map[string]bool{}
		for _, item := range items {
			prName := item.Repository + "#" + fmtInt(item.Number)
			if !seenPR[prName] {
				prs = append(prs, prName)
				seenPR[prName] = true
			}
			if item.Path != "" {
				locations = append(locations, item.Repository+":"+item.Path+":"+fmtInt(item.Line))
			}
			urls = append(urls, item.URL)
		}
		row := types.NewRow(
			types.MRP("title", key),
			types.MRP("count", len(items)),
			types.MRP("prs", strings.Join(prs, ",")),
			types.MRP("locations", strings.Join(locations, "; ")),
			types.MRP("urls", strings.Join(urls, "; ")),
			types.MRP("sample_body", bodyPreview(items[0].Body, false)),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func codexCommentRow(c batchCodexComment, fullBody bool) types.Row {
	return types.NewRow(
		types.MRP("pr", c.PR),
		types.MRP("repository", c.Repository),
		types.MRP("number", c.Number),
		types.MRP("title", c.Title),
		types.MRP("signal_url", c.SignalURL),
		types.MRP("kind", c.Kind),
		types.MRP("author", c.Author),
		types.MRP("reviewed_commit", c.ReviewedCommit),
		types.MRP("current_head", c.CurrentHead),
		types.MRP("truncated", c.Truncated),
		types.MRP("path", c.Path),
		types.MRP("line", c.Line),
		types.MRP("body", bodyPreview(c.Body, fullBody)),
		types.MRP("url", c.URL),
	)
}

var boldTitleRE = regexp.MustCompile(`\*\*([^*]+)\*\*`)
var badgeRE = regexp.MustCompile(`<[^>]+>|!\[[^]]*\]\([^)]*\)`)

func codexMessageTitle(body string) string {
	body = strings.TrimSpace(body)
	if m := boldTitleRE.FindStringSubmatch(body); len(m) > 1 {
		return normalizeCodexTitle(m[1])
	}
	for _, line := range strings.Split(body, "\n") {
		line = normalizeCodexTitle(line)
		if line != "" {
			return line
		}
	}
	return "(empty Codex message)"
}

func normalizeCodexTitle(s string) string {
	s = badgeRE.ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "**", " ")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.Join(strings.Fields(s), " ")
	return strings.TrimSpace(s)
}

func bodyPreview(body string, full bool) string {
	if full || len(body) <= 240 {
		return body
	}
	return body[:240]
}

func fmtInt(v int) string {
	return strconv.Itoa(v)
}
