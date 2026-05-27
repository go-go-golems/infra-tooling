package prready

import (
	"testing"

	"github.com/go-go-golems/infra-tooling/pkg/prref"
)

func TestClassifyCodexInlineFeedback(t *testing.T) {
	report := Classify(Snapshot{
		PR: prref.Ref{Owner: "o", Repo: "r", Number: 1}, HeadRefOID: "abc123",
		Checks:  []Check{{Kind: "CheckRun", Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"}},
		Signals: []CodexSignal{{Kind: "review", Author: "chatgpt-codex-connector", Time: "2026-01-01T00:00:00Z", CodexAuthored: true, Body: "Reviewed commit:** `abc123`", ThumbsUp: 1, Comments: []ReviewComment{{Path: "x.go", Line: 10, Body: "fix this", URL: "u"}}}},
	})
	if report.State != CodexFeedback || !report.Terminal {
		t.Fatalf("state=%s terminal=%v", report.State, report.Terminal)
	}
}

func TestClassifyStaleCodexFeedbackWaitsForNewCodex(t *testing.T) {
	report := Classify(Snapshot{
		PR: prref.Ref{Owner: "o", Repo: "r", Number: 1}, HeadRefOID: "def456",
		Checks: []Check{{Kind: "CheckRun", Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"}},
		Signals: []CodexSignal{
			{Kind: "review", Author: "chatgpt-codex-connector", Time: "2026-01-01T00:00:00Z", CodexAuthored: true, Body: "Reviewed commit:** `abc123`", ThumbsUp: 0, Comments: []ReviewComment{{Path: "x.go", Line: 10, Body: "fix this", URL: "u"}}},
			{Kind: "codex-trigger", Author: "wesen", Time: "2026-01-01T00:01:00Z"},
		},
	})
	if report.State != WaitingCodex || report.Terminal {
		t.Fatalf("state=%s terminal=%v", report.State, report.Terminal)
	}
}

func TestClassifyReady(t *testing.T) {
	report := Classify(Snapshot{
		PR: prref.Ref{Owner: "o", Repo: "r", Number: 1}, HeadRefOID: "abc123",
		Checks:  []Check{{Kind: "CheckRun", Name: "test", Status: "COMPLETED", Conclusion: "SUCCESS"}},
		Signals: []CodexSignal{{Kind: "review", Author: "chatgpt-codex-connector", Time: "2026-01-01T00:00:00Z", CodexAuthored: true, Body: "Didn't find any major issues. :+1:"}},
	})
	if !report.OK || report.State != Ready {
		t.Fatalf("state=%s ok=%v findings=%#v", report.State, report.OK, report.Findings)
	}
}
