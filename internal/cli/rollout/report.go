package rollout

import (
	"context"
	"fmt"
	"os"

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

type reportCommand struct{ *cmds.CommandDescription }
type reportSettings struct {
	Config  string `glazed:"config"`
	WriteTo string `glazed:"write-to"`
}

func newReportCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &reportCommand{CommandDescription: cmds.NewCommandDescription("report",
		cmds.WithShort("Generate a Markdown rollout report"),
		cmds.WithArguments(fields.New("config", fields.TypeString, fields.WithHelp("Rollout YAML file"), fields.WithIsArgument(true))),
		cmds.WithFlags(fields.New("write-to", fields.TypeString, fields.WithHelp("Write report to this path instead of stdout"))),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *reportCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &reportSettings{}
	if err := decodeDefault(vals, s); err != nil {
		return err
	}
	cfg, err := rolloutpkg.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	report, err := rolloutpkg.MarkdownReport(cfg)
	if err != nil {
		return err
	}
	if s.WriteTo != "" {
		if err := os.WriteFile(s.WriteTo, []byte(report), 0o644); err != nil {
			return err
		}
	} else {
		fmt.Print(report)
	}
	return gp.AddRow(ctx, types.NewRow(types.MRP("config", s.Config), types.MRP("report_file", s.WriteTo), types.MRP("bytes", len(report))))
}
