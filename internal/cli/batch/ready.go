package batch

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
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
	Until               string `glazed:"until"`
	SummaryOnly         bool   `glazed:"summary-only"`
	MarkdownReport      bool   `glazed:"markdown-report"`
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

Watch mode repeats until the configured stop condition. The default --until=actionable
stops on all-ready, terminal failures, Codex feedback, merge conflicts, or partial readiness.`),
		cmds.WithArguments(fields.New("file", fields.TypeString, fields.WithHelp("YAML PR list"), fields.WithIsArgument(true))),
		cmds.WithFlags(
			fields.New("watch", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Poll until the batch has an actionable state")),
			fields.New("interval-seconds", fields.TypeInteger, fields.WithDefault(30), fields.WithHelp("Watch polling interval in seconds")),
			fields.New("timeout-seconds", fields.TypeInteger, fields.WithDefault(1800), fields.WithHelp("Watch timeout in seconds")),
			fields.New("trigger-missing-codex", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Post @codex review for PRs with no Codex signal")),
			fields.New("until", fields.TypeChoice, fields.WithDefault("actionable"), fields.WithChoices("actionable", "all-ready", "terminal", "first-ready"), fields.WithHelp("Watch stop condition")),
			fields.New("summary-only", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Emit grouped summary rows instead of detailed per-PR rows")),
			fields.New("markdown-report", fields.TypeBool, fields.WithDefault(false), fields.WithHelp("Print a copy/paste-ready Markdown readiness report")),
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
	until := s.Until
	if until == "" {
		until = "actionable"
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
		reports, summary, err := collectBatchReports(ctx, client, refs, s.TriggerMissingCodex)
		if err != nil {
			return err
		}
		if s.MarkdownReport {
			_, _ = fmt.Fprint(os.Stdout, markdownBatchReport(reports, summary))
		} else if err := emitBatchRows(ctx, gp, reports, summary, attempt, s.SummaryOnly); err != nil {
			return err
		}
		code := summary.exitCode()
		if !s.Watch || shouldStopWatch(summary, until) {
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

type batchPRReport struct {
	Ref        prref.Ref
	Report     prready.Report
	TriggerURL string
}

func collectBatchReports(ctx context.Context, client ghclient.Client, refs []prref.Ref, triggerMissing bool) ([]batchPRReport, batchSummary, error) {
	summary := batchSummary{}
	reports := make([]batchPRReport, 0, len(refs))
	for _, ref := range refs {
		report, err := client.Readiness(ctx, ref)
		triggerURL := ""
		if err != nil {
			summary.Errors++
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
		reports = append(reports, batchPRReport{Ref: ref, Report: report, TriggerURL: triggerURL})
	}
	summary.State = summary.state()
	return reports, summary, nil
}

func emitBatchRows(ctx context.Context, gp middlewares.Processor, reports []batchPRReport, summary batchSummary, attempt int, summaryOnly bool) error {
	if summaryOnly {
		if err := emitSummaryRows(ctx, gp, reports, summary, attempt); err != nil {
			return err
		}
	} else {
		for _, r := range reports {
			if err := gp.AddRow(ctx, detailedBatchRow(r, attempt)); err != nil {
				return err
			}
		}
	}
	return gp.AddRow(ctx, summaryRow(summary, attempt))
}

func detailedBatchRow(r batchPRReport, attempt int) types.Row {
	report := r.Report
	ref := r.Ref
	return types.NewRow(
		types.MRP("attempt", attempt),
		types.MRP("pr", ref.URL()),
		types.MRP("repository", ref.Repository()),
		types.MRP("number", ref.Number),
		types.MRP("ok", report.OK),
		types.MRP("state", string(report.State)),
		types.MRP("terminal", report.Terminal),
		types.MRP("terminal_reason", prready.TerminalReason(report)),
		types.MRP("next_action", prready.NextAction(report)),
		types.MRP("failed_check_kinds", report.FailedCheckKinds),
		types.MRP("pending_checks", prready.PendingChecks(report)),
		types.MRP("failed_checks", prready.FailedChecksSummary(report)),
		types.MRP("merge_state_status", report.MergeStateStatus),
		types.MRP("head_ref_oid", report.HeadRefOID),
		types.MRP("trigger_url", r.TriggerURL),
	)
}

func emitSummaryRows(ctx context.Context, gp middlewares.Processor, reports []batchPRReport, summary batchSummary, attempt int) error {
	groups := map[string][]batchPRReport{}
	for _, r := range reports {
		groups[batchCategory(r.Report)] = append(groups[batchCategory(r.Report)], r)
	}
	keys := []string{"ready", "codex_feedback", "failed_checks", "merge_conflict", "waiting_checks", "waiting_codex", "no_codex", "other"}
	for _, key := range keys {
		for _, r := range groups[key] {
			if err := gp.AddRow(ctx, summaryOnlyRow(r, key, attempt)); err != nil {
				return err
			}
		}
	}
	_ = summary
	return nil
}

func summaryOnlyRow(r batchPRReport, category string, attempt int) types.Row {
	report := r.Report
	ref := r.Ref
	return types.NewRow(
		types.MRP("attempt", attempt),
		types.MRP("category", category),
		types.MRP("repository", ref.Repository()),
		types.MRP("number", ref.Number),
		types.MRP("pr", ref.URL()),
		types.MRP("state", string(report.State)),
		types.MRP("next_action", prready.NextAction(report)),
		types.MRP("pending_checks", strings.Join(prready.PendingChecks(report), "; ")),
		types.MRP("failed_checks", strings.Join(prready.FailedChecksSummary(report), "; ")),
		types.MRP("merge_state_status", report.MergeStateStatus),
	)
}

func summaryRow(summary batchSummary, attempt int) types.Row {
	return types.NewRow(
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
}

func batchCategory(report prready.Report) string {
	switch report.State {
	case prready.Ready:
		return "ready"
	case prready.CodexFeedback:
		return "codex_feedback"
	case prready.FailedChecks:
		return "failed_checks"
	case prready.MergeConflict:
		return "merge_conflict"
	case prready.WaitingChecks:
		return "waiting_checks"
	case prready.WaitingCodex:
		return "waiting_codex"
	case prready.NoCodex:
		return "no_codex"
	default:
		return "other"
	}
}

func markdownBatchReport(reports []batchPRReport, summary batchSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## Batch readiness\n\n")
	fmt.Fprintf(&b, "- Ready: %d\n", summary.Ready)
	fmt.Fprintf(&b, "- Not ready: %d\n", summary.NotReady)
	fmt.Fprintf(&b, "- Codex feedback: %d\n", summary.CodexFeedback)
	fmt.Fprintf(&b, "- Failed checks: %d\n", summary.FailedChecks)
	fmt.Fprintf(&b, "- Merge conflicts: %d\n", summary.MergeConflicts)
	fmt.Fprintf(&b, "- Errors: %d\n", summary.Errors)
	fmt.Fprintf(&b, "- State: `%s`\n\n", summary.State)
	groups := map[string][]batchPRReport{}
	for _, r := range reports {
		groups[batchCategory(r.Report)] = append(groups[batchCategory(r.Report)], r)
	}
	sections := []struct{ key, title string }{{"ready", "Ready"}, {"codex_feedback", "Codex feedback"}, {"failed_checks", "Failed checks"}, {"merge_conflict", "Merge conflicts"}, {"waiting_checks", "Waiting checks"}, {"waiting_codex", "Waiting on Codex"}, {"no_codex", "Missing Codex"}, {"other", "Other"}}
	for _, section := range sections {
		items := groups[section.key]
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Ref.Repository() < items[j].Ref.Repository() })
		fmt.Fprintf(&b, "### %s\n\n", section.title)
		for _, r := range items {
			detail := markdownDetail(r.Report)
			if detail != "" {
				detail = " — " + detail
			}
			fmt.Fprintf(&b, "- [%s#%d](%s): `%s`, next: `%s`%s\n", r.Ref.Repository(), r.Ref.Number, r.Ref.URL(), r.Report.State, prready.NextAction(r.Report), detail)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func markdownDetail(report prready.Report) string {
	if pending := prready.PendingChecks(report); len(pending) > 0 {
		return "pending " + strings.Join(pending, "; ")
	}
	if failed := prready.FailedChecksSummary(report); len(failed) > 0 {
		return "failed " + strings.Join(failed, "; ")
	}
	return prready.TerminalReason(report)
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

func shouldStopWatch(s batchSummary, until string) bool {
	state := s.state()
	terminalBlocker := state == "error" || state == "codex_feedback" || state == "failed_checks" || state == "merge_conflict"
	switch until {
	case "all-ready":
		return state == "ready" || terminalBlocker
	case "terminal":
		return state == "ready" || terminalBlocker
	case "first-ready":
		return s.Ready > 0 || terminalBlocker
	default:
		return state == "ready" || terminalBlocker || state == "partial_ready"
	}
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
