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
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"github.com/spf13/cobra"
)

type watchCommand struct{ *cmds.CommandDescription }

type watchSettings struct {
	PR              string `glazed:"pr"`
	IntervalSeconds int    `glazed:"interval-seconds"`
	TimeoutSeconds  int    `glazed:"timeout-seconds"`
}

func newWatchCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &watchCommand{CommandDescription: cmds.NewCommandDescription(
		"watch",
		cmds.WithShort("Poll PR readiness until an actionable state"),
		cmds.WithLong(`Poll one pull request until GitHub checks and Codex signals reach an actionable state.

The command exits immediately when the PR is ready, when checks fail, when
Codex posts current-head feedback, or when the timeout expires. It is intended
as a Codex-aware alternative to waiting on opaque check timeouts.`),
		cmds.WithArguments(fields.New("pr", fields.TypeString, fields.WithHelp("PR URL or owner/repo#number"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("interval-seconds", fields.TypeInteger, fields.WithDefault(30), fields.WithHelp("Polling interval in seconds")),
			fields.New("timeout-seconds", fields.TypeInteger, fields.WithDefault(1800), fields.WithHelp("Watch timeout in seconds")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *watchCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &watchSettings{}
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
	interval := time.Duration(s.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	timeout := time.Duration(s.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}

	client := ghclient.Client{}
	start := time.Now()
	attempt := 1
	for {
		report, err := client.Readiness(ctx, ref)
		if err != nil {
			return err
		}
		elapsed := time.Since(start).Round(time.Second)
		row := types.NewRow(
			types.MRP("attempt", attempt),
			types.MRP("elapsed_seconds", int(elapsed.Seconds())),
			types.MRP("pr", ref.URL()),
			types.MRP("repository", ref.Repository()),
			types.MRP("number", ref.Number),
			types.MRP("ok", report.OK),
			types.MRP("state", string(report.State)),
			types.MRP("terminal", report.Terminal),
			types.MRP("terminal_reason", prready.TerminalReason(report)),
			types.MRP("next_action", prready.NextAction(report)),
			types.MRP("failed_check_kinds", report.FailedCheckKinds),
			types.MRP("merge_state_status", report.MergeStateStatus),
			types.MRP("review_decision", report.ReviewDecision),
			types.MRP("head_ref_oid", report.HeadRefOID),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return err
		}
		if report.OK || report.Terminal {
			if !report.OK {
				exitcode.Request(exitCodeForState(report.State))
			}
			return nil
		}
		if time.Since(start) >= timeout {
			exitcode.Request(1)
			return fmt.Errorf("timed out after %s waiting for PR readiness; last state=%s next_action=%s", elapsed, report.State, prready.NextAction(report))
		}
		attempt++
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
