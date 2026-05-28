package batch

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{Use: "batch", Short: "Batch pull request readiness commands"}
	ready, err := newReadyCommand()
	if err != nil {
		return nil, err
	}
	comments, err := newCodexCommentsCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(ready, comments, newActionsCommand())
	return root, nil
}
