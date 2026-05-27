package batch

import (
	"context"
	"fmt"
	"time"

	glazedcli "github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/fields"
	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/settings"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/go-go-golems/infra-tooling/internal/exitcode"
	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prlist"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"github.com/spf13/cobra"
)

type readyCommand struct{ *cmds.CommandDescription }

type readySettings struct {
	File                string `glazed:"file"`
	Watch               bool   `glazed:"watch"`
	IntervalSeconds     int    `glazed:"interval-seconds"`
	TimeoutSeconds      int    `glazed:"timeout-seconds"`
	TriggerMissingCodex bool   `glazed:"trigger-missing-codex"`
}

func newReadyCommand() (*cobra.Command, error) {
	glazedSection, err := settings.NewGlazedSchema()
	if err != nil {
		return nil, err
	}
	commandSettingsSection, err := glazedcli.NewCommandSettingsSection()
	if err != nil {
		return nil, err
	}
	cmd := &readyCommand{CommandDescription: cmds.NewCommandDescription(
		"ready",
		cmds.WithShort("Check readiness for a YAML list of PRs"),
		cmds.WithLong(`Check readiness for every PR in a YAML file.

Input format:

prs:
  - https://github.com/go-go-golems/discord-bot/pull/9
  - repo: go-go-golems/goja-git
    number: 2

Watch mode repeats while all PRs are still waiting. It stops on all-ready,
terminal failures, Codex feedback, or partial readiness.`),
		cmds.WithArguments(fields.New("file", fields.TypeString, fields.WithHelp("YAML PR list"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("watch", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Poll until the batch has an actionable state")),
			fields.New("interval-seconds", fields.TypeInteger, fields.WithDefault(30), fields.WithHelp("Watch polling interval in seconds")),
			fields.New("timeout-seconds", fields.TypeInteger, fields.WithDefault(1800), fields.WithHelp("Watch timeout in seconds")),
			fields.New("trigger-missing-codex", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Post @codex review for PRs with no Codex signal")),
		),
		cmds.WithSections(glazedSection, commandSettingsSection),
	)}
	return glazedcli.BuildCobraCommandFromCommand(cmd, glazedcli.WithParserConfig(glazedcli.CobraParserConfig{ShortHelpSections: []string{schema.DefaultSlug, settings.GlazedSlug}, MiddlewaresFunc: glazedcli.CobraCommandDefaultMiddlewares}))
}

func (c *readyCommand) RunIntoGlazeProcessor(ctx context.Context, vals *values.Values, gp middlewares.Processor) error {
	s := &readySettings{}
	if err := vals.DecodeSectionInto(schema.DefaultSlug, s); err != nil {
		return err
	}
	if s.File == "" {
		return fmt.Errorf("provide a YAML PR list")
	}
	refs, err := prlist.Load(s.File)
	if err != nil {
		return err
	}
	if len(refs) == 0 {
		return fmt.Errorf("PR list is empty")
	}
	interval := time.Duration(s.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 30 * time.Second
	}
	timeout := time.Duration(s.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Minute
	}
	start := time.Now()
	attempt := 1
	client := ghclient.Client{}
	for {
		summary, err := runOnce(ctx, client, refs, s.TriggerMissingCodex, attempt, gp)
		if err != nil {
			return err
		}
		code := summary.exitCode()
		if !s.Watch || code == 0 || code == 2 || code == 3 || code == 4 || code == 5 || code == 6 {
			if code != 0 {
				exitcode.Request(code)
			}
			return nil
		}
		if time.Since(start) >= timeout {
			return fmt.Errorf("timed out after %s waiting for batch readiness", time.Since(start).Round(time.Second))
		}
		attempt++
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

type batchSummary struct {
	Ready          int
	NotReady       int
	CodexFeedback  int
	FailedChecks   int
	MergeConflicts int
	Errors         int
	State          string
}

func runOnce(ctx context.Context, client ghclient.Client, refs []prref.Ref, triggerMissing bool, attempt int, gp middlewares.Processor) (batchSummary, error) {
	summary := batchSummary{}
	for _, ref := range refs {
		report, err := client.Readiness(ctx, ref)
		triggerURL := ""
		if err != nil {
			summary.Errors++
			summary.NotReady++
			report = prready.Report{PR: ref, URL: ref.URL(), State: "error", Terminal: true}
		} else if report.State == prready.NoCodex && triggerMissing {
			triggerURL, err = client.TriggerCodex(ctx, ref)
			if err != nil {
				summary.Errors++
			} else {
				report.State = "codex_triggered"
			}
		}
		if report.State == prready.Ready {
			summary.Ready++
		} else {
			summary.NotReady++
		}
		switch report.State {
		case prready.CodexFeedback:
			summary.CodexFeedback++
		case prready.FailedChecks:
			summary.FailedChecks++
		case prready.MergeConflict:
			summary.MergeConflicts++
		}
		row := types.NewRow(
			types.MRP("attempt", attempt),
			types.MRP("pr", ref.URL()),
			types.MRP("repository", ref.Repository()),
			types.MRP("number", ref.Number),
			types.MRP("ok", report.OK),
			types.MRP("state", string(report.State)),
			types.MRP("terminal", report.Terminal),
			types.MRP("failed_check_kinds", report.FailedCheckKinds),
			types.MRP("trigger_url", triggerURL),
		)
		if err := gp.AddRow(ctx, row); err != nil {
			return summary, err
		}
	}
	summary.State = summary.state()
	row := types.NewRow(
		types.MRP("attempt", attempt),
		types.MRP("pr", ""),
		types.MRP("repository", "summary"),
		types.MRP("number", 0),
		types.MRP("ok", summary.NotReady == 0),
		types.MRP("state", summary.State),
		types.MRP("ready", summary.Ready),
		types.MRP("not_ready", summary.NotReady),
		types.MRP("codex_feedback", summary.CodexFeedback),
		types.MRP("failed_checks", summary.FailedChecks),
		types.MRP("merge_conflicts", summary.MergeConflicts),
		types.MRP("errors", summary.Errors),
	)
	return summary, gp.AddRow(ctx, row)
}

func (s batchSummary) state() string {
	if s.NotReady == 0 {
		return "ready"
	}
	if s.Errors > 0 {
		return "error"
	}
	if s.MergeConflicts > 0 {
		return "merge_conflict"
	}
	if s.CodexFeedback > 0 {
		return "codex_feedback"
	}
	if s.FailedChecks > 0 {
		return "failed_checks"
	}
	if s.Ready > 0 {
		return "partial_ready"
	}
	return "waiting"
}

func (s batchSummary) exitCode() int {
	switch s.state() {
	case "ready":
		return 0
	case "error":
		return 2
	case "codex_feedback":
		return 3
	case "failed_checks":
		return 4
	case "partial_ready":
		return 5
	case "merge_conflict":
		return 6
	default:
		return 1
	}
}
