package rollout

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
	rolloutpkg "github.com/go-go-golems/infra-tooling/pkg/rollout"
	"github.com/spf13/cobra"
)

type initCommand struct{ *cmds.CommandDescription }
type initSettings struct {
	ID            string `glazed:"id"`
	Name          string `glazed:"name"`
	Workspace     string `glazed:"workspace"`
	Include       string `glazed:"include"`
	RequireModule string `glazed:"require-module"`
	Branch        string `glazed:"branch"`
	Base          string `glazed:"base"`
	CommitMessage string `glazed:"commit-message"`
	Validation    string `glazed:"validation"`
	LogDir        string `glazed:"log-dir"`
	WriteTo       string `glazed:"write-to"`
}

func newInitCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &initCommand{CommandDescription: cmds.NewCommandDescription("init",
		cmds.WithShort("Create a rollout YAML file"),
		cmds.WithFlags(
			fields.New("id", fields.TypeString, fields.WithHelp("Ticket or rollout id")),
			fields.New("name", fields.TypeString, fields.WithHelp("Rollout name")),
			fields.New("workspace", fields.TypeString, fields.WithDefault("."), fields.WithHelp("Workspace root")),
			fields.New("include", fields.TypeString, fields.WithHelp("Comma-separated repo names or paths")),
			fields.New("require-module", fields.TypeString, fields.WithHelp("Comma-separated go.mod requirements")),
			fields.New("branch", fields.TypeString, fields.WithDefault("rollout/change"), fields.WithHelp("Rollout branch name")),
			fields.New("base", fields.TypeString, fields.WithDefault("origin/main"), fields.WithHelp("Base ref")),
			fields.New("commit-message", fields.TypeString, fields.WithDefault("Apply rollout changes"), fields.WithHelp("Commit message")),
			fields.New("validation", fields.TypeString, fields.WithHelp("Validation command to run in each repo")),
			fields.New("log-dir", fields.TypeString, fields.WithDefault(".ggg-rollout-logs"), fields.WithHelp("Validation log directory, relative to workspace unless absolute")),
			fields.New("write-to", fields.TypeString, fields.WithHelp("Path to write rollout YAML")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *initCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &initSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	if s.WriteTo == "" {
		return fmt.Errorf("--write-to is required")
	}
	cfg := rolloutpkg.Config{ID: s.ID, Name: s.Name, Workspace: s.Workspace, Branch: s.Branch, Base: s.Base, CommitMessage: s.CommitMessage}
	cfg.Selection.Include = csv(s.Include)
	cfg.Selection.RequireGoModContains = csv(s.RequireModule)
	if s.Validation != "" {
		cfg.Validation.Commands = []rolloutpkg.ValidationCommand{{Name: "validation", Run: s.Validation}}
	}
	cfg.Validation.ContinueOnError = true
	cfg.Validation.LogDir = s.LogDir
	cfg.PullRequest.Title = s.CommitMessage
	if err := rolloutpkg.SaveConfig(s.WriteTo, cfg); err != nil {
		return err
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("config_file", s.WriteTo), types.MRP("workspace", s.Workspace), types.MRP("targets", cfg.Selection.Include), types.MRP("validation_commands", len(cfg.Validation.Commands))))
}
