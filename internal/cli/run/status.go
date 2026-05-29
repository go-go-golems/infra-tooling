package run

import (
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/infra-tooling/internal/cli/actionoutput"
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	"github.com/go-go-golems/infra-tooling/pkg/actionstatus"
	"github.com/spf13/cobra"
)

func newStatusCommand() *cobra.Command {
	var repo, branch, sha, output string
	var limit, intervalSeconds, timeoutSeconds int
	var ignores []string
	var watch, summaryOnly bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check GitHub Actions runs for one repository",
		Long: `Check GitHub Actions runs for one repository/branch/commit.

Failures in workflows named by --ignore-workflow are reported as ignored_failure
and do not make the command fail. Use this for known-noisy workflows such as
Secret Scanning while still catching real CI/lint/release failures.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if repo == "" {
				return fmt.Errorf("--repo is required")
			}
			if intervalSeconds <= 0 {
				intervalSeconds = 30
			}
			if timeoutSeconds <= 0 {
				timeoutSeconds = 1800
			}
			deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
			for {
				res, err := actionstatus.Collect(cmd.Context(), actionstatus.Options{Repo: repo, Branch: branch, SHA: sha, Limit: limit, IgnoreWorkflows: ignores})
				if err != nil {
					return err
				}
				if err := actionoutput.Print(os.Stdout, output, res.Runs, res.Summary, summaryOnly); err != nil {
					return err
				}
				code := actionstatus.ExitCode(res.Summary)
				if !watch || res.Summary.State != "pending" {
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
	cmd.Flags().StringVar(&repo, "repo", "", "GitHub repository owner/name")
	cmd.Flags().StringVar(&branch, "branch", "main", "Branch to inspect")
	cmd.Flags().StringVar(&sha, "sha", "", "Optional head SHA prefix to filter runs")
	cmd.Flags().IntVar(&limit, "limit", 20, "Maximum runs to read from gh")
	cmd.Flags().StringArrayVar(&ignores, "ignore-workflow", []string{}, "Workflow name to treat as non-blocking failure; may be repeated or comma-separated")
	cmd.Flags().BoolVar(&watch, "watch", false, "Poll while matching runs are pending")
	cmd.Flags().IntVar(&intervalSeconds, "interval-seconds", 30, "Watch polling interval in seconds")
	cmd.Flags().IntVar(&timeoutSeconds, "timeout-seconds", 1800, "Watch timeout in seconds")
	cmd.Flags().StringVar(&output, "output", "table", "Output format: table, json, yaml")
	cmd.Flags().BoolVar(&summaryOnly, "summary-only", false, "Only print summary table for table output")
	return cmd
}
