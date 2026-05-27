package release

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
	releasepkg "github.com/go-go-golems/infra-tooling/pkg/release"
	"github.com/spf13/cobra"
)

type tagCommand struct {
	*cmds.CommandDescription
	mode releasepkg.BumpMode
}

type tagSettings struct {
	Repo   string `glazed:"repo"`
	DryRun bool   `glazed:"dry-run"`
}

func newTagCommand(mode string) (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	name := "tag-" + mode
	cmd := &tagCommand{mode: releasepkg.BumpMode(mode), CommandDescription: cmds.NewCommandDescription(
		name,
		cmds.WithShort("Create, push, and verify the next "+mode+" release tag"),
		cmds.WithLong("Fetch origin/main and tags, compute the next "+mode+" version with svu, create and push the tag, then verify the module through proxy.golang.org."),
		cmds.WithFlags(
			fields.New("repo", fields.TypeString, fields.WithDefault("."), fields.WithHelp("Repository directory")),
			fields.New("dry-run", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Compute the next tag without creating or pushing it")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *tagCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &tagSettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	res, err := releasepkg.Tag(ctx, s.Repo, c.mode, s.DryRun)
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("repo", res.RepoDir),
		types.MRP("mode", string(res.Mode)),
		types.MRP("module", res.Module),
		types.MRP("tag", res.Tag),
		types.MRP("commit", res.Commit),
		types.MRP("dry_run", s.DryRun),
		types.MRP("verified", res.Verified),
	)
	return gp.AddRow(ctx, row)
}
