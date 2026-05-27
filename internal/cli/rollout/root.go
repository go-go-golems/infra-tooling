package rollout

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{Use: "rollout", Short: "Plan and operate multi-repository rollouts"}
	for _, factory := range []func() (*cobra.Command, error){
		newInventoryCommand,
		newInitCommand,
		newValidateCommand,
		newBranchCommand,
		newPushPRsCommand,
		newReportCommand,
	} {
		sub, err := factory()
		if err != nil {
			return nil, err
		}
		cmd.AddCommand(sub)
	}
	return cmd, nil
}
