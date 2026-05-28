package batch

import (
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/infra-tooling/internal/cli/actionoutput"
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	"github.com/go-go-golems/infra-tooling/pkg/actionstatus"
	"github.com/spf13/cobra"
)

func newActionsCommand() *cobra.Command {
	var output string
	var limit, intervalSeconds, timeoutSeconds int
	var ignores []string
	var watch, summaryOnly bool
	cmd := &cobra.Command{
		Use:   "actions <manifest.yaml>",
		Short: "Check GitHub Actions runs for many repositories",
		Long: `Check GitHub Actions runs for a YAML manifest of repositories.

Manifest format:

repos:
  - repo: go-go-golems/css-visual-diff
    branch: main
    sha: 8559422

Failures in workflows named by --ignore-workflow are reported as ignored_failure
and do not make the command fail.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manifest, err := actionstatus.LoadManifest(args[0])
			if err != nil {
				return err
			}
			if intervalSeconds <= 0 {
				intervalSeconds = 30
			}
			if timeoutSeconds <= 0 {
				timeoutSeconds = 1800
			}
			deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
			for {
				runs, summary, err := actionstatus.BatchCollect(cmd.Context(), manifest, ignores, limit)
				if err != nil {
					return err
				}
				if err := actionoutput.Print(os.Stdout, output, runs, summary, summaryOnly); err != nil {
					return err
				}
				code := actionstatus.ExitCode(summary)
				if !watch || summary.State != "pending" {
					if code != 0 {
						exitcode.Request(code)
					}
					return nil
				}
				if time.Now().After(deadline) {
					return fmt.Errorf("timed out after %ds waiting for actions", timeoutSeconds)
				}
				if err := actionstatus.Sleep(cmd.Context(), time.Duration(intervalSeconds)*time.Second); err != nil {
					return err
				}
			}
		},
	}
	cmd.Flags().StringArrayVar(&ignores, "ignore-workflow", []string{}, "Workflow name to treat as non-blocking failure; may be repeated or comma-separated")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll while matching runs are pending")
	cmd.Flags().IntVar(&intervalSeconds, "interval-seconds", 30, "Watch polling interval in seconds")
	cmd.Flags().IntVar(&timeoutSeconds, "timeout-seconds", 1800, "Watch timeout in seconds")
	cmd.Flags().IntVar(&limit, "limit", 20, "Default maximum runs to read per repository")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, yaml")
	cmd.Flags().BoolVar(&summaryOnly, "summary-only", false, "Only print summary table for table output")
	return cmd
}
