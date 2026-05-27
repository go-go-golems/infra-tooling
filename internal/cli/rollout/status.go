package rollout

import (
	"context"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	rolloutpkg "github.com/go-go-golems/infra-tooling/pkg/rollout"
	"github.com/spf13/cobra"
)

type statusCommand struct{ *cmds.CommandDescription }
type statusSettings struct {
	Config string `glazed:"config"`
}

func newStatusCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &statusCommand{CommandDescription: cmds.NewCommandDescription("status",
		cmds.WithShort("Show local branch and PR readiness rollout status"),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *statusCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &statusSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	results, err := rolloutpkg.Status(ctx, cfg)
	if err != nil {
		return err
	}
	bad := 0
	for _, r := range results {
		if !r.OK {
			bad++
		}
		prState, terminal := "", false
		if r.Ready != nil {
			prState = string(r.Ready.State)
			terminal = r.Ready.Terminal
		}
		if err := gp.AddRow(ctx, types.NewRow(types.MRP("repo", r.Repo.Name), types.MRP("path", r.Repo.Path), types.MRP("branch", r.Repo.CurrentBranch), types.MRP("ahead_base", r.Repo.AheadBase), types.MRP("branch_ok", r.BranchOK), types.MRP("pr", r.PR), types.MRP("pr_state", prState), types.MRP("terminal", terminal), types.MRP("ok", r.OK), types.MRP("state", r.State), types.MRP("message", r.Message))); err != nil {
			return err
		}
	}
	if bad > 0 {
		exitcode.Request(1)
	}
	return nil
}
