package cli

import (
	"github.com/go-go-golems/infra-tooling/internal/cli/batch"
	"github.com/go-go-golems/infra-tooling/internal/cli/pr"
	"github.com/go-go-golems/infra-tooling/internal/cli/release"
	"github.com/spf13/cobra"
)

func NewRootCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:           "ggg",
		Short:         "Manage go-go-golems open-source repositories",
		Long:          "Manage go-go-golems pull requests, Codex reviews, release trains, validation profiles, and tags.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.PersistentFlags().Bool("with-structured-output", false, "Compatibility flag: Glazed commands already emit row-oriented structured data; use --output json/yaml/csv to select format")

	prCmd, err := pr.NewCommand()
	if err != nil {
		return nil, err
	}
	releaseCmd, err := release.NewCommand()
	if err != nil {
		return nil, err
	}
	batchCmd, err := batch.NewCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(prCmd, releaseCmd, batchCmd)
	root.AddCommand(&cobra.Command{Use: "repo", Short: "Repository dependency and validation commands (planned)"})
	root.AddCommand(&cobra.Command{Use: "train", Short: "Release-train orchestration commands (planned)"})
	return root, nil
}
