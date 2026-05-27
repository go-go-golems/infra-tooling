package prready

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/go-go-golems/infra-tooling/pkg/prref"
)

var (
	benignBodyRE = regexp.MustCompile(`(?i)^\s*(?:approved|looks good|lgtm|no issues found|✅|👍|:\+1:|:thumbsup:|thumbs up|nit:)?\s*$`)
	satisfiedRE  = regexp.MustCompile(`(?is)(didn'?t find (?:any )?major issues|no major issues|looks good|lgtm).*(?:👍|:\+1:|:thumbsup:|thumbs up)`)
	reviewedRE   = regexp.MustCompile("Reviewed commit:\\*\\*\\s*`([0-9a-fA-F]+)`")
)

type State string

const (
	Ready         State = "ready"
	WaitingChecks State = "waiting_checks"
	WaitingCodex  State = "waiting_codex"
	NoCodex       State = "no_codex"
	FailedChecks  State = "failed_checks"
	CodexFeedback State = "codex_feedback"
	NotReady      State = "not_ready"
)

type Report struct {
	OK               bool      `json:"ok"`
	State            State     `json:"state"`
	Terminal         bool      `json:"terminal"`
	FailedCheckKinds []string  `json:"failedCheckKinds"`
	PR               prref.Ref `json:"pr"`
	URL              string    `json:"url"`
	MergeStateStatus string    `json:"mergeStateStatus"`
	ReviewDecision   string    `json:"reviewDecision"`
	HeadRefOID       string    `json:"headRefOid"`
	Findings         []Finding `json:"findings"`
}

type Finding struct {
	OK      bool   `json:"ok"`
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

type Snapshot struct {
	PR               prref.Ref
	URL              string
	MergeStateStatus string
	ReviewDecision   string
	HeadRefOID       string
	Checks           []Check
	Signals          []CodexSignal
}

type Check struct {
	Name       string
	Kind       string
	Status     string
	Conclusion string
	State      string
	URL        string
}

type CodexSignal struct {
	Kind          string
	Author        string
	URL           string
	Time          string
	Body          string
	CodexAuthored bool
	Eyes          int
	ThumbsUp      int
	Comments      []ReviewComment
}

type ReviewComment struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Body string `json:"body"`
	URL  string `json:"url"`
}

func Classify(s Snapshot) Report {
	findings := append(checkFindings(s.Checks), codexFindings(s)...)
	state, terminal, kinds := classifyFindings(findings)
	return Report{OK: state == Ready, State: state, Terminal: terminal, FailedCheckKinds: kinds, PR: s.PR, URL: s.URL, MergeStateStatus: s.MergeStateStatus, ReviewDecision: s.ReviewDecision, HeadRefOID: s.HeadRefOID, Findings: findings}
}

func checkFindings(checks []Check) []Finding {
	if len(checks) == 0 {
		return []Finding{{OK: false, Kind: "checks", Message: "no status checks found; actions may not have run"}}
	}
	var pending, bad []string
	for _, c := range checks {
		name := c.Name
		if name == "" {
			name = "<unnamed check>"
		}
		switch c.Kind {
		case "CheckRun":
			if c.Status != "COMPLETED" {
				pending = append(pending, name+": status="+c.Status)
			} else if c.Conclusion != "SUCCESS" && c.Conclusion != "SKIPPED" && c.Conclusion != "NEUTRAL" {
				bad = append(bad, name+": conclusion="+c.Conclusion)
			}
		case "StatusContext":
			if c.State != "SUCCESS" {
				bad = append(bad, name+": state="+c.State)
			}
		}
	}
	var out []Finding
	if len(pending) > 0 {
		out = append(out, Finding{OK: false, Kind: "checks", Message: "pending checks: " + strings.Join(pending, "; ")})
	}
	if len(bad) > 0 {
		out = append(out, Finding{OK: false, Kind: "checks", Message: "failing/non-success checks: " + strings.Join(bad, "; ")})
	}
	if len(pending) == 0 && len(bad) == 0 {
		out = append(out, Finding{OK: true, Kind: "checks", Message: "all status checks completed successfully"})
	}
	return out
}

func codexFindings(s Snapshot) []Finding {
	if len(s.Signals) == 0 {
		return []Finding{{OK: false, Kind: "codex", Message: "no Codex-authored review/comment signal found"}}
	}
	signals := append([]CodexSignal(nil), s.Signals...)
	sort.Slice(signals, func(i, j int) bool { return signals[i].Time < signals[j].Time })
	latest := signals[len(signals)-1]
	out := []Finding{{OK: true, Kind: "codex", Message: "latest Codex signal (" + latest.Kind + ") by " + latest.Author + ": " + latest.URL}}
	var latestAuthored *CodexSignal
	for i := range signals {
		if signals[i].CodexAuthored {
			latestAuthored = &signals[i]
		}
	}
	if latestAuthored != nil {
		out = append(out, Finding{OK: true, Kind: "codex", Message: "latest Codex-authored signal (" + latestAuthored.Kind + ") by " + latestAuthored.Author + ": " + latestAuthored.URL})
		reviewed := ReviewedCommit(latestAuthored.Body)
		if reviewed != "" && !strings.HasPrefix(strings.ToLower(s.HeadRefOID), reviewed) {
			out = append(out, Finding{OK: true, Kind: "codex", Message: "latest Codex-authored findings are stale for an older head commit"})
		} else if len(latestAuthored.Comments) > 0 {
			out = append(out, Finding{OK: false, Kind: "codex", Message: "latest Codex-authored review has code review comment(s): " + FormatComments(latestAuthored.Comments)})
		} else if !BodyIsBenign(latestAuthored.Body) {
			out = append(out, Finding{OK: false, Kind: "codex", Message: "latest Codex-authored body contains substantive comments: " + preview(latestAuthored.Body)})
		} else {
			out = append(out, Finding{OK: true, Kind: "codex", Message: "latest Codex-authored body is empty/benign/satisfied and has no code review comments"})
		}
	}
	if latest.ThumbsUp <= 0 && !(latest.CodexAuthored && BodyIsSatisfied(latest.Body)) {
		out = append(out, Finding{OK: false, Kind: "codex", Message: "latest Codex signal has no thumbs-up reaction or satisfied thumbs-up body"})
	} else {
		out = append(out, Finding{OK: true, Kind: "codex", Message: "latest Codex signal is satisfied"})
	}
	if latest.Eyes > 0 {
		out = append(out, Finding{OK: false, Kind: "codex", Message: "latest Codex signal has eyes reaction(s), review may still be running"})
	} else {
		out = append(out, Finding{OK: true, Kind: "codex", Message: "latest Codex signal has no eyes reaction"})
	}
	return out
}

func classifyFindings(findings []Finding) (State, bool, []string) {
	allOK := true
	var failed []string
	for _, f := range findings {
		if !f.OK {
			allOK = false
			failed = append(failed, f.Message)
		}
	}
	if allOK {
		return Ready, true, nil
	}
	kinds := failedCheckKinds(failed)
	for _, msg := range failed {
		if strings.Contains(msg, "code review comment") || strings.Contains(msg, "substantive comments") {
			return CodexFeedback, true, kinds
		}
	}
	for _, msg := range failed {
		if strings.Contains(msg, "failing/non-success checks:") {
			return FailedChecks, true, kinds
		}
	}
	for _, msg := range failed {
		if strings.Contains(msg, "pending checks:") {
			return WaitingChecks, false, kinds
		}
	}
	for _, msg := range failed {
		if strings.Contains(msg, "no Codex-authored") {
			return NoCodex, false, kinds
		}
	}
	for _, msg := range failed {
		if strings.Contains(msg, "eyes reaction") || strings.Contains(msg, "no thumbs-up") {
			return WaitingCodex, false, kinds
		}
	}
	return NotReady, false, kinds
}

func failedCheckKinds(failed []string) []string {
	set := map[string]bool{}
	for _, msg := range failed {
		lower := strings.ToLower(msg)
		if strings.Contains(lower, "pending checks:") {
			set["pending_checks"] = true
		}
		if strings.Contains(lower, "failing/non-success checks:") {
			set["checks"] = true
		}
		if strings.Contains(lower, "test") {
			set["test"] = true
		}
		if strings.Contains(lower, "lint") {
			set["lint"] = true
		}
		if strings.Contains(lower, "govuln") || strings.Contains(lower, "vulnerability") {
			set["govulncheck"] = true
		}
		if strings.Contains(lower, "gosec") || strings.Contains(lower, "security scan") {
			set["gosec"] = true
		}
		if strings.Contains(lower, "dependency review") {
			set["dependency_review"] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func BodyIsSatisfied(body string) bool { return satisfiedRE.MatchString(body) }
func BodyIsBenign(body string) bool {
	b := strings.TrimSpace(body)
	return b == "" || benignBodyRE.MatchString(b) || BodyIsSatisfied(b)
}
func ReviewedCommit(body string) string {
	m := reviewedRE.FindStringSubmatch(body)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

func FormatComments(comments []ReviewComment) string {
	parts := make([]string, 0, len(comments))
	max := len(comments)
	if max > 3 {
		max = 3
	}
	for _, c := range comments[:max] {
		loc := c.Path
		if c.Line > 0 {
			loc += ":" + strconv.Itoa(c.Line)
		}
		parts = append(parts, loc+": "+preview(c.Body)+" ("+c.URL+")")
	}
	if len(comments) > 3 {
		parts = append(parts, "...")
	}
	return strings.Join(parts, "; ")
}
func preview(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 240 {
		return s[:240]
	}
	return s
}
