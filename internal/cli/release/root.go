package release

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{Use: "release", Short: "Tag and publish Go module releases"}
	for _, mode := range []string{"patch", "minor", "major"} {
		cmd, err := newTagCommand(mode)
		if err != nil {
			return nil, err
		}
		root.AddCommand(cmd)
	}
	root.AddCommand(newWatchCommand(), newVerifyDocsCommand(), newPreflightCommand())
	return root, nil
}
