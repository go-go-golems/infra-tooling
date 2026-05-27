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

type planCommand struct{ *cmds.CommandDescription }

type planSettings struct {
	Config  string `glazed:"config"`
	Profile string `glazed:"profile"`
}

func newPlanCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &planCommand{CommandDescription: cmds.NewCommandDescription("plan",
		cmds.WithShort("Plan profile-specific rollout file changes without mutating repositories"),
		cmds.WithLong("Inspect rollout targets and emit one row per profile-specific operation that is present, needed, or worth manual inspection."),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithFlags(fields.New("profile", fields.TypeString, fields.WithDefault(rolloutpkg.ProfileGlazedLint), fields.WithHelp("Rollout profile to plan"))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *planCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &planSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	ops, err := rolloutpkg.Plan(cfg, rolloutpkg.PlanOptions{Profile: s.Profile})
	if err != nil {
		return err
	}
	needed := 0
	for _, op := range ops {
		if op.Status == "needed" {
			needed++
		}
		if err := gp.AddRow(ctx, types.NewRow(
			types.MRP("repo", op.Repo.Name),
			types.MRP("path", op.Repo.Path),
			types.MRP("profile", op.Profile),
			types.MRP("file", op.File),
			types.MRP("kind", op.Kind),
			types.MRP("status", op.Status),
			types.MRP("description", op.Description),
			types.MRP("detail", op.Detail),
		)); err != nil {
			return err
		}
	}
	if needed > 0 {
		exitcode.Request(1)
	}
	return nil
}
