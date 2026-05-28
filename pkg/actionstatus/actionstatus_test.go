package actionstatus

import "testing"

func TestClassify(t *testing.T) {
	ignore := map[string]bool{"Secret Scanning": true, "secret scanning": true}
	tests := []struct {
		name       string
		status     string
		conclusion string
		workflow   string
		want       string
		ignored    bool
	}{
		{name: "success", status: "completed", conclusion: "success", workflow: "test", want: "ok"},
		{name: "skipped", status: "completed", conclusion: "skipped", workflow: "Open GitOps PR", want: "ok"},
		{name: "pending", status: "in_progress", workflow: "test", want: "pending"},
		{name: "failed", status: "completed", conclusion: "failure", workflow: "lint", want: "failed"},
		{name: "cancelled", status: "completed", conclusion: "cancelled", workflow: "release", want: "failed"},
		{name: "ignored failure", status: "completed", conclusion: "failure", workflow: "Secret Scanning", want: "ignored_failure", ignored: true},
		{name: "startup failure", status: "completed", conclusion: "startup_failure", workflow: "release.yml", want: "failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ignored := Classify(tt.status, tt.conclusion, tt.workflow, ignore)
			if got != tt.want || ignored != tt.ignored {
				t.Fatalf("Classify() = (%q,%v), want (%q,%v)", got, ignored, tt.want, tt.ignored)
			}
		})
	}
}

func TestStateAndExitCode(t *testing.T) {
	tests := []struct {
		summary Summary
		state   string
		code    int
	}{
		{Summary{Total: 2, Success: 1, IgnoredFailures: 1}, "ok", 0},
		{Summary{Total: 2, Failed: 1, Pending: 1}, "failed", 1},
		{Summary{Total: 1, Other: 1}, "failed", 1},
		{Summary{Total: 2, Pending: 1, Success: 1}, "pending", 2},
		{Summary{NoRuns: 1}, "no_runs", 2},
		{Summary{Total: 2, Success: 2, NoRuns: 1}, "pending", 2},
		{Summary{}, "no_runs", 2},
	}
	for _, tt := range tests {
		s := tt.summary
		s.State = State(s)
		if s.State != tt.state {
			t.Fatalf("State(%+v) = %q, want %q", tt.summary, s.State, tt.state)
		}
		if got := ExitCode(s); got != tt.code {
			t.Fatalf("ExitCode(%+v) = %d, want %d", s, got, tt.code)
		}
	}
}
