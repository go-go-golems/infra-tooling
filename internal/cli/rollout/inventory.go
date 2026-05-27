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
	rolloutpkg "github.com/go-go-golems/infra-tooling/pkg/rollout"
	"github.com/spf13/cobra"
)

type inventoryCommand struct{ *cmds.CommandDescription }
type inventorySettings struct {
	Root          string `glazed:"root"`
	RequireModule string `glazed:"require-module"`
	Base          string `glazed:"base"`
}

func newInventoryCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &inventoryCommand{CommandDescription: cmds.NewCommandDescription("inventory",
		cmds.WithShort("Inventory repositories under a workspace"),
		cmds.WithFlags(
			fields.New("root", fields.TypeString, fields.WithDefault("."), fields.WithHelp("Workspace root to scan")),
			fields.New("require-module", fields.TypeString, fields.WithHelp("Comma-separated module paths that go.mod must require")),
			fields.New("base", fields.TypeString, fields.WithDefault("origin/main"), fields.WithHelp("Git base ref used for ahead counts")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *inventoryCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &inventorySettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	repos, err := rolloutpkg.Inventory(s.Root, rolloutpkg.InventoryOptions{RequireModules: csv(s.RequireModule), Base: s.Base})
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if err := addRepoRow(ctx, gp, repo); err != nil {
			return err
		}
	}
	return nil
}
