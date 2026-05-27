package prready

import "strings"

func TerminalReason(report Report) string {
	if report.OK {
		return "ready"
	}
	if !report.Terminal {
		return ""
	}
	switch report.State {
	case CodexFeedback:
		return "codex_feedback"
	case FailedChecks:
		return "failed_checks"
	case MergeConflict:
		return "merge_conflict"
	default:
		return string(report.State)
	}
}

func NextAction(report Report) string {
	switch report.State {
	case Ready:
		return "merge_when_manual_review_allows"
	case WaitingChecks:
		if len(report.FailedCheckKinds) > 0 {
			return "wait_for_" + strings.Join(report.FailedCheckKinds, ",")
		}
		return "wait_for_checks"
	case WaitingCodex:
		return "wait_for_codex"
	case NoCodex:
		return "trigger_codex_review"
	case CodexFeedback:
		return "inspect_and_address_codex_feedback"
	case FailedChecks:
		return "inspect_and_fix_failed_checks"
	case MergeConflict:
		return "rebase_or_merge_main_to_resolve_conflicts"
	default:
		return "inspect_pr_readiness"
	}
}

func PendingChecks(report Report) []string {
	return findingPayloads(report, "pending checks: ")
}

func FailedChecksSummary(report Report) []string {
	return findingPayloads(report, "failing/non-success checks: ")
}

func findingPayloads(report Report, prefix string) []string {
	var out []string
	for _, f := range report.Findings {
		if payload, ok := strings.CutPrefix(f.Message, prefix); ok {
			out = append(out, strings.Split(payload, "; ")...)
		}
	}
	return out
}
