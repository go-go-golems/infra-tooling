package actionoutput

import (
	"encoding/json"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/go-go-golems/infra-tooling/pkg/actionstatus"
	"gopkg.in/yaml.v3"
)

func Print(w io.Writer, format string, runs []actionstatus.Run, summary actionstatus.Summary, summaryOnly bool) error {
	switch format {
	case "", "table":
		return printTable(w, runs, summary, summaryOnly)
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(struct {
			Runs    []actionstatus.Run   `json:"runs"`
			Summary actionstatus.Summary `json:"summary"`
		}{Runs: runs, Summary: summary})
	case "yaml":
		return yaml.NewEncoder(w).Encode(struct {
			Runs    []actionstatus.Run   `yaml:"runs"`
			Summary actionstatus.Summary `yaml:"summary"`
		}{Runs: runs, Summary: summary})
	default:
		return fmt.Errorf("unsupported output %q", format)
	}
}

func printTable(w io.Writer, runs []actionstatus.Run, summary actionstatus.Summary, summaryOnly bool) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if !summaryOnly {
		_, _ = fmt.Fprintln(tw, "REPO\tSHA\tWORKFLOW\tSTATUS\tCONCLUSION\tCLASSIFICATION\tURL")
		for _, r := range runs {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", r.Repo, r.SHA, r.Workflow, r.Status, r.Conclusion, r.Classification, r.URL)
		}
	}
	_, _ = fmt.Fprintln(tw, "")
	_, _ = fmt.Fprintln(tw, "SUMMARY\tSTATE\tTOTAL\tSUCCESS\tIGNORED_FAILURES\tFAILED\tPENDING\tNO_RUNS\tOTHER")
	_, _ = fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%d\n", summary.Repo, summary.State, summary.Total, summary.Success, summary.IgnoredFailures, summary.Failed, summary.Pending, summary.NoRuns, summary.Other)
	return tw.Flush()
}
