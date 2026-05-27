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

type pushPRsCommand struct{ *cmds.CommandDescription }
type pushPRsSettings struct {
	Config       string `glazed:"config"`
	DryRun       bool   `glazed:"dry-run"`
	Yes          bool   `glazed:"yes"`
	NoVerifyPush bool   `glazed:"no-verify-push"`
	Reason       string `glazed:"reason"`
}

func newPushPRsCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &pushPRsCommand{CommandDescription: cmds.NewCommandDescription("push-prs",
		cmds.WithShort("Push rollout branches and open PRs"),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Show planned push/PR actions")),
			fields.New("yes", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Confirm push and PR creation")),
			fields.New("no-verify-push", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Use git push --no-verify")),
			fields.New("reason", fields.TypeString, fields.WithHelp("Reason for --no-verify-push")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *pushPRsCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &pushPRsSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	results, err := rolloutpkg.PushPRs(ctx, cfg, rolloutpkg.PushPROptions{DryRun: s.DryRun, Yes: s.Yes, NoVerifyPush: s.NoVerifyPush, Reason: s.Reason})
	if err != nil {
		return err
	}
	bad := 0
	for _, r := range results {
		if !r.OK {
			bad++
		}
		if err := gp.AddRow(ctx, types.NewRow(types.MRP("repo", r.Repo.Name), types.MRP("branch", r.Repo.CurrentBranch), types.MRP("ahead_base", r.Repo.AheadBase), types.MRP("action", r.Action), types.MRP("ok", r.OK), types.MRP("url", r.URL), types.MRP("message", r.Message))); err != nil {
			return err
		}
	}
	if bad > 0 {
		exitcode.Request(1)
	}
	return nil
}
