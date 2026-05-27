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

type branchCommand struct{ *cmds.CommandDescription }
type branchSettings struct {
	Config string `glazed:"config"`
	Commit bool   `glazed:"commit"`
	Yes    bool   `glazed:"yes"`
}

func newBranchCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &branchCommand{CommandDescription: cmds.NewCommandDescription("branch",
		cmds.WithShort("Inspect or commit rollout branches"),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("commit", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Create/reset rollout branches and commit configured paths")),
			fields.New("yes", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Confirm branch commit mutation")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *branchCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &branchSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	var results []rolloutpkg.BranchResult
	if s.Commit {
		results, err = rolloutpkg.CommitTargets(cfg, s.Yes)
	} else {
		results, err = rolloutpkg.BranchStatus(cfg)
	}
	if err != nil {
		return err
	}
	bad := 0
	for _, r := range results {
		if !r.OK {
			bad++
		}
		if err := gp.AddRow(ctx, types.NewRow(types.MRP("repo", r.Repo.Name), types.MRP("path", r.Repo.Path), types.MRP("branch", r.Repo.CurrentBranch), types.MRP("expected_branch", r.ExpectedBranch), types.MRP("base", r.ExpectedBase), types.MRP("ahead_base", r.Repo.AheadBase), types.MRP("dirty_tracked", r.Repo.DirtyTracked), types.MRP("dirty_untracked", r.Repo.DirtyUntracked), types.MRP("ok", r.OK), types.MRP("message", r.Message))); err != nil {
			return err
		}
	}
	if bad > 0 {
		exitcode.Request(1)
	}
	return nil
}
