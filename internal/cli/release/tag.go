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
	Repo       string `glazed:"repo"`
	DryRun     bool   `glazed:"dry-run"`
	AllowDirty bool   `glazed:"allow-dirty"`
	Target     string `glazed:"from"`
	Commit     string `glazed:"commit"`
	Yes        bool   `glazed:"yes"`
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
			fields.New("allow-dirty", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Allow tagging from a dirty worktree")),
			fields.New("from", fields.TypeString, fields.WithDefault("origin/main"), fields.WithHelp("Git ref to tag by default")),
			fields.New("commit", fields.TypeString, fields.WithHelp("Explicit commit/ref to tag; overrides --from")),
			fields.New("yes", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Confirm pushing the new tag")),
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
	target := s.Target
	if s.Commit != "" {
		target = s.Commit
	}
	res, err := releasepkg.TagWithOptions(ctx, releasepkg.Options{RepoDir: s.Repo, Mode: c.mode, DryRun: s.DryRun, AllowDirty: s.AllowDirty, Target: target, Yes: s.Yes})
	if err != nil {
		return err
	}
	row := types.NewRow(
		types.MRP("repo", res.RepoDir),
		types.MRP("mode", string(res.Mode)),
		types.MRP("module", res.Module),
		types.MRP("current_tag", res.CurrentTag),
		types.MRP("tag", res.Tag),
		types.MRP("target", res.Target),
		types.MRP("commit", res.Commit),
		types.MRP("dirty", res.Dirty),
		types.MRP("existing_tag", res.ExistingTag),
		types.MRP("dry_run", s.DryRun),
		types.MRP("verified", res.Verified),
		types.MRP("plan", res.Plan),
	)
	return gp.AddRow(ctx, row)
}
