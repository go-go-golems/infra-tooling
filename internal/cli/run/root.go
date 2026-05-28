package run

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{Use: "run", Short: "GitHub Actions run status helpers"}
	root.AddCommand(newStatusCommand())
	return root, nil
}
