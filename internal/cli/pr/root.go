package pr

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{Use: "pr", Short: "Pull request and Codex review commands"}
	trigger, err := newCodexTriggerCommand()
	if err != nil {
		return nil, err
	}
	ready, err := newReadyCommand()
	if err != nil {
		return nil, err
	}
	comments, err := newCodexCommentsCommand()
	if err != nil {
		return nil, err
	}
	watch, err := newWatchCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(trigger, ready, comments, watch)
	return root, nil
}
