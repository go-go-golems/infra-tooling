package prready

import (
	"sort"
	"strings"
	"time"
)

func SortedSignals(s Snapshot) []CodexSignal {
	out := append([]CodexSignal(nil), s.Signals...)
	sort.Slice(out, func(i, j int) bool { return out[i].Time < out[j].Time })
	return out
}

func LatestSignal(s Snapshot) (CodexSignal, bool) {
	signals := SortedSignals(s)
	if len(signals) == 0 {
		return CodexSignal{}, false
	}
	return signals[len(signals)-1], true
}

func LatestAuthoredSignal(s Snapshot) (CodexSignal, bool) {
	signals := SortedSignals(s)
	for i := len(signals) - 1; i >= 0; i-- {
		if signals[i].CodexAuthored {
			return signals[i], true
		}
	}
	return CodexSignal{}, false
}

func SignalReviewedCurrentHead(signal CodexSignal, head string) bool {
	reviewed := ReviewedCommit(signal.Body)
	head = strings.ToLower(head)
	return reviewed == "" || strings.HasPrefix(head, reviewed)
}

func HasCurrentAuthoredFeedback(s Snapshot) bool {
	signal, ok := LatestAuthoredSignal(s)
	if !ok || !SignalReviewedCurrentHead(signal, s.HeadRefOID) {
		return false
	}
	return signal.CommentsTruncated || len(signal.Comments) > 0 || !BodyIsBenign(signal.Body)
}

func RecentTrigger(s Snapshot, now time.Time, window time.Duration) (CodexSignal, bool, time.Duration) {
	if window <= 0 {
		return CodexSignal{}, false, 0
	}
	latest, ok := LatestSignal(s)
	if !ok || latest.Kind != "codex-trigger" {
		return CodexSignal{}, false, 0
	}
	parsed, err := time.Parse(time.RFC3339, latest.Time)
	if err != nil {
		return latest, false, 0
	}
	age := now.Sub(parsed)
	return latest, age >= 0 && age < window, age
}
