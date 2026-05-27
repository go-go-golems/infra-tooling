package pr

import (
	"context"
	"fmt"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"github.com/spf13/cobra"
)

type codexCommentsCommand struct{ *cmds.CommandDescription }
type codexCommentsSettings struct {
	PR       string `glazed:"pr"`
	FullBody bool   `glazed:"full-body"`
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
		cmds.WithShort("List Codex-authored review comments for a PR"),
		cmds.WithLong("Emit structured rows for Codex-authored review bodies and inline review comments, including stale/current-head status."),
		cmds.WithArguments(fields.New("pr", fields.TypeString, fields.WithHelp("PR URL or owner/repo#number"), fields.WithIsArgument(true))),
		cmds.WithFlags(fields.New("full-body", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Emit full comment bodies instead of previews"))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *codexCommentsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &codexCommentsSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if s.PR == "" {
		return fmt.Errorf("provide a PR argument")
	}
	ref, err := prref.Parse(s.PR)
	if err != nil {
		return err
	}
	snap, err := ghclient.Client{}.Snapshot(ctx, ref)
	if err != nil {
		return err
	}
	for _, sig := range prready.SortedSignals(snap) {
		if !sig.CodexAuthored {
			continue
		}
		reviewed := prready.ReviewedCommit(sig.Body)
		current := prready.SignalReviewedCurrentHead(sig, snap.HeadRefOID)
		truncated := sig.CommentsTruncated || snap.ReviewsTruncated || snap.CommentsTruncated
		if len(sig.Comments) == 0 {
			row := types.NewRow(types.MRP("pr", ref.URL()), types.MRP("signal_url", sig.URL), types.MRP("kind", sig.Kind), types.MRP("author", sig.Author), types.MRP("reviewed_commit", reviewed), types.MRP("current_head", current), types.MRP("truncated", truncated), types.MRP("path", ""), types.MRP("line", 0), types.MRP("body", bodyForOutput(sig.Body, s.FullBody)), types.MRP("url", sig.URL))
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
			continue
		}
		for _, comment := range sig.Comments {
			row := types.NewRow(types.MRP("pr", ref.URL()), types.MRP("signal_url", sig.URL), types.MRP("kind", sig.Kind), types.MRP("author", sig.Author), types.MRP("reviewed_commit", reviewed), types.MRP("current_head", current), types.MRP("truncated", truncated), types.MRP("path", comment.Path), types.MRP("line", comment.Line), types.MRP("body", bodyForOutput(comment.Body, s.FullBody)), types.MRP("url", comment.URL))
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
	}
	return nil
}

func bodyForOutput(body string, full bool) string {
	if full {
		return body
	}
	if len(body) > 240 {
		return body[:240]
	}
	return body
}
