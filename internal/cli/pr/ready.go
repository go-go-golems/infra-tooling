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
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"github.com/spf13/cobra"
)

type readyCommand struct{ *cmds.CommandDescription }
type readySettings struct {
	PR       string `glazed:"pr"`
	Findings bool   `glazed:"findings"`
}

func newReadyCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &readyCommand{CommandDescription: cmds.NewCommandDescription(
		"ready",
		cmds.WithShort("Check whether a PR is ready to merge"),
		cmds.WithLong("Check GitHub status checks and Codex review signals for one pull request. Use --findings to emit one row per finding."),
		cmds.WithArguments(fields.New("pr", fields.TypeString, fields.WithHelp("PR URL or owner/repo#number"), fields.WithIsArgument(true))),
		cmds.WithFlags(fields.New("findings", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Emit one row per finding instead of one summary row"))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *readyCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &readySettings{}
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
	report, err := ghclient.Client{}.Readiness(ctx, ref)
	if err != nil {
		return err
	}
	if s.Findings {
		for i, f := range report.Findings {
			row := types.NewRow(types.MRP("pr", ref.URL()), types.MRP("state", string(report.State)), types.MRP("ok", f.OK), types.MRP("terminal", report.Terminal), types.MRP("finding_index", i), types.MRP("finding_kind", f.Kind), types.MRP("message", f.Message))
			if err := gp.AddRow(ctx, row); err != nil {
				return err
			}
		}
		if !report.OK {
			exitcode.Request(exitCodeForState(report.State))
		}
		return nil
	}
	row := types.NewRow(
		types.MRP("pr", ref.URL()), types.MRP("repository", ref.Repository()), types.MRP("number", ref.Number),
		types.MRP("ok", report.OK), types.MRP("state", string(report.State)), types.MRP("terminal", report.Terminal),
		types.MRP("terminal_reason", prready.TerminalReason(report)), types.MRP("next_action", prready.NextAction(report)),
		types.MRP("failed_check_kinds", report.FailedCheckKinds), types.MRP("pending_checks", prready.PendingChecks(report)), types.MRP("failed_checks", prready.FailedChecksSummary(report)),
		types.MRP("merge_state_status", report.MergeStateStatus), types.MRP("review_decision", report.ReviewDecision), types.MRP("head_ref_oid", report.HeadRefOID),
	)
	if err := gp.AddRow(ctx, row); err != nil {
		return err
	}
	if !report.OK {
		exitcode.Request(exitCodeForState(report.State))
	}
	return nil
}

func exitCodeForState(state prready.State) int {
	switch state {
	case prready.Ready:
		return 0
	case prready.CodexFeedback:
		return 3
	case prready.FailedChecks:
		return 4
	case prready.MergeConflict:
		return 6
	default:
		return 1
	}
}
