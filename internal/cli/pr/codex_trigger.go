package pr

import (
	"context"
	"fmt"
	"time"

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

type codexTriggerCommand struct{ *cmds.CommandDescription }

type codexTriggerSettings struct {
	PR                  string `glazed:"pr"`
	File                string `glazed:"file"`
	Force               bool   `glazed:"force"`
	DryRun              bool   `glazed:"dry-run"`
	Yes                 bool   `glazed:"yes"`
	RecentTriggerWindow string `glazed:"recent-trigger-window"`
}

func newCodexTriggerCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &codexTriggerCommand{CommandDescription: cmds.NewCommandDescription(
		"codex-trigger",
		cmds.WithShort("Trigger Codex review for one or more PRs"),
		cmds.WithLong(`Trigger Codex review by posting the standard '@codex review' comment.

By default the command first checks the latest Codex signal and skips PRs that
already have an EYES reaction, which indicates that a Codex run may still be in
progress. Use --force to post another trigger anyway. Use --file with a YAML PR
list:

prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2
`),
		cmds.WithArguments(fields.New("pr", fields.TypeString, fields.WithHelp("PR URL or owner/repo#number"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("file", fields.TypeString, fields.WithHelp("YAML file containing prs entries")),
			fields.New("force", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Trigger even if a Codex EYES reaction indicates a run is already in progress")),
			fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Show what would happen without posting comments")),
			fields.New("yes", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Confirm mutating operation without prompting; currently informational for non-interactive use")),
			fields.New("recent-trigger-window", fields.TypeString, fields.WithDefault("10m"), fields.WithHelp("Skip duplicate @codex review triggers newer than this duration unless --force is set")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *codexTriggerCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &codexTriggerSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	window, err := time.ParseDuration(s.RecentTriggerWindow)
	if err != nil {
		return fmt.Errorf("invalid recent-trigger-window: %w", err)
	}
	refs, err := refsFromSettings(s)
	if err != nil {
		return err
	}
	client := ghclient.Client{}
	for _, ref := range refs {
		snap, err := client.Snapshot(ctx, ref)
		if err != nil {
			return err
		}
		latest, _ := prready.LatestSignal(snap)
		status := ghclient.CodexStatus{SignalURL: latest.URL, SignalKind: latest.Kind, Author: latest.Author, Eyes: latest.Eyes, ThumbsUp: latest.ThumbsUp, Running: latest.Eyes > 0}
		action := "triggered"
		url := ""
		currentFeedback := prready.HasCurrentAuthoredFeedback(snap)
		_, recentTrigger, triggerAge := prready.RecentTrigger(snap, time.Now(), window)
		if s.DryRun {
			action = "would_trigger"
		} else if status.Running && !s.Force {
			action = "skipped_running"
		} else if currentFeedback && !s.Force {
			action = "skipped_current_feedback"
		} else if recentTrigger && !s.Force {
			action = "skipped_recent_trigger"
		} else {
			url, err = client.TriggerCodex(ctx, ref)
			if err != nil {
				return err
			}
		}
		row := types.NewRow(
			types.MRP("pr", ref.URL()),
			types.MRP("repository", ref.Repository()),
			types.MRP("number", ref.Number),
			types.MRP("action", action),
			types.MRP("force", s.Force),
			types.MRP("dry_run", s.DryRun),
			types.MRP("codex_running", status.Running),
			types.MRP("current_feedback", currentFeedback),
			types.MRP("recent_trigger", recentTrigger),
			types.MRP("trigger_age_seconds", int(triggerAge.Seconds())),
			types.MRP("recent_trigger_window", window.String()),
			types.MRP("eyes", status.Eyes),
			types.MRP("thumbs_up", status.ThumbsUp),
			types.MRP("signal_url", status.SignalURL),
			types.MRP("trigger_url", url),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

func refsFromSettings(s *codexTriggerSettings) ([]prref.Ref, error) {
	if s.File != "" {
		return prlist.Load(s.File)
	}
	if s.PR == "" {
		return nil, fmt.Errorf("provide a PR argument or --file prs.yaml")
	}
	ref, err := prref.Parse(s.PR)
	if err != nil {
		return nil, err
	}
	return []prref.Ref{ref}, nil
}
