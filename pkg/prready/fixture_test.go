package prready

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestClassifySnapshotFixtures(t *testing.T) {
	tests := []struct {
		fixture  string
		state    State
		terminal bool
	}{
		{"ready.json", Ready, true},
		{"failed_checks.json", FailedChecks, true},
		{"codex_feedback_current_head.json", CodexFeedback, true},
		{"waiting_codex_running.json", WaitingCodex, false},
		{"stale_codex_feedback_waiting.json", WaitingCodex, false},
		{"truncated_current_head_feedback.json", CodexFeedback, true},
	}
	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			snap := loadSnapshotFixture(t, tt.fixture)
			report := Classify(snap)
			if report.State != tt.state {
				t.Fatalf("state = %s, want %s; findings=%#v", report.State, tt.state, report.Findings)
			}
			if report.Terminal != tt.terminal {
				t.Fatalf("terminal = %v, want %v", report.Terminal, tt.terminal)
			}
		})
	}
}

func loadSnapshotFixture(t *testing.T, name string) Snapshot {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	var snap Snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		t.Fatal(err)
	}
	return snap
}
