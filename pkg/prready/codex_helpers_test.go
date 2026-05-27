package prready

import (
	"testing"
	"time"
)

func TestRecentTrigger(t *testing.T) {
	now := time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)
	snap := Snapshot{Signals: []CodexSignal{{Kind: "codex-trigger", Author: "wesen", Time: now.Add(-2 * time.Minute).Format(time.RFC3339)}}}
	_, ok, age := RecentTrigger(snap, now, 10*time.Minute)
	if !ok {
		t.Fatalf("expected recent trigger")
	}
	if age != 2*time.Minute {
		t.Fatalf("age = %s", age)
	}
}

func TestClassifyTruncatedCurrentReviewAsFeedback(t *testing.T) {
	report := Classify(Snapshot{
		HeadRefOID: "abc123",
		Checks:     []Check{{Kind: "CheckRun", Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"}},
		Signals: []CodexSignal{{
			Kind:              "review",
			Author:            "chatgpt-codex-connector",
			Time:              "2026-01-01T00:00:00Z",
			CodexAuthored:     true,
			Body:              "Reviewed commit:** `abc123`",
			ThumbsUp:          1,
			CommentsTruncated: true,
		}},
	})
	if report.State != CodexFeedback {
		t.Fatalf("state = %s, want %s", report.State, CodexFeedback)
	}
}

func TestHasSatisfiedCodexSignal(t *testing.T) {
	snap := Snapshot{HeadRefOID: "abcdef", Signals: []CodexSignal{{Kind: "review", Author: "chatgpt-codex-connector", CodexAuthored: true, Time: "2026-01-01T00:00:00Z", Body: "Reviewed commit:** `abcdef`\n\nDidn't find any major issues. :+1:"}}}
	if !HasSatisfiedCodexSignal(snap) {
		t.Fatalf("expected satisfied codex signal")
	}
}

func TestHasSatisfiedCodexSignalRejectsStaleFeedback(t *testing.T) {
	snap := Snapshot{HeadRefOID: "abcdef", Signals: []CodexSignal{{Kind: "review", Author: "chatgpt-codex-connector", CodexAuthored: true, Time: "2026-01-01T00:00:00Z", Body: "Reviewed commit:** `123456`\n\nDidn't find any major issues. :+1:"}}}
	if HasSatisfiedCodexSignal(snap) {
		t.Fatalf("stale authored signal should not be treated as satisfied for current head")
	}
}
