package prref

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Ref struct {
	Owner  string `json:"owner" yaml:"owner"`
	Repo   string `json:"repo" yaml:"repo"`
	Number int    `json:"number" yaml:"number"`
}

func (r Ref) Repository() string {
	if r.Owner == "" || r.Repo == "" {
		return ""
	}
	return r.Owner + "/" + r.Repo
}

func (r Ref) String() string {
	if r.Repository() == "" || r.Number == 0 {
		return ""
	}
	return fmt.Sprintf("%s#%d", r.Repository(), r.Number)
}

func (r Ref) URL() string {
	if r.Repository() == "" || r.Number == 0 {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/pull/%d", r.Repository(), r.Number)
}

func (r Ref) Validate() error {
	if strings.TrimSpace(r.Owner) == "" {
		return fmt.Errorf("PR owner is required")
	}
	if strings.TrimSpace(r.Repo) == "" {
		return fmt.Errorf("PR repo is required")
	}
	if r.Number <= 0 {
		return fmt.Errorf("PR number must be positive")
	}
	return nil
}

var (
	urlRE   = regexp.MustCompile(`^https?://github\.com/([^/]+)/([^/]+)/pull/(\d+)(?:[/?#].*)?$`)
	shortRE = regexp.MustCompile(`^([^/]+)/([^#]+)#(\d+)$`)
)

func Parse(s string) (Ref, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Ref{}, fmt.Errorf("PR reference is empty")
	}
	if m := urlRE.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[3])
		return Ref{Owner: m[1], Repo: m[2], Number: n}, nil
	}
	if m := shortRE.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[3])
		return Ref{Owner: m[1], Repo: m[2], Number: n}, nil
	}
	return Ref{}, fmt.Errorf("could not parse PR reference %q; expected GitHub PR URL or owner/repo#number", s)
}
