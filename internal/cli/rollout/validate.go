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

type validateCommand struct{ *cmds.CommandDescription }
type validateSettings struct {
	Config string `glazed:"config"`
	DryRun bool   `glazed:"dry-run"`
}

func newValidateCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &validateCommand{CommandDescription: cmds.NewCommandDescription("validate",
		cmds.WithShort("Run rollout validation commands in each target repo"),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithFlags(fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Emit planned validation rows without running commands"))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *validateCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &validateSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	results, err := rolloutpkg.Validate(ctx, cfg, rolloutpkg.ValidationOptions{DryRun: s.DryRun})
	if err != nil {
		return err
	}
	failed := 0
	for _, r := range results {
		if !r.OK {
			failed++
		}
		if err := gp.AddRow(ctx, types.NewRow(types.MRP("repo", r.Repo.Name), types.MRP("path", r.Repo.Path), types.MRP("command", r.Command), types.MRP("run", r.Run), types.MRP("ok", r.OK), types.MRP("exit_code", r.ExitCode), types.MRP("log_path", r.LogPath), types.MRP("dry_run", r.DryRun))); err != nil {
			return err
		}
	}
	if failed > 0 {
		exitcode.Request(4)
	}
	return nil
}
