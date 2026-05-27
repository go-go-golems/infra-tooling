package prlist

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-go-golems/infra-tooling/pkg/prref"
	"gopkg.in/yaml.v3"
)

type File struct {
	PRs []Entry `yaml:"prs"`
}

type Entry struct {
	Raw    string `yaml:"-"`
	Repo   string `yaml:"repo"`
	Number int    `yaml:"number"`
	URL    string `yaml:"url"`
	Ref    string `yaml:"ref"`
}

func (e *Entry) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		e.Raw = value.Value
		return nil
	case yaml.MappingNode:
		type entry Entry
		var out entry
		if err := value.Decode(&out); err != nil {
			return err
		}
		*e = Entry(out)
		return nil
	default:
		return fmt.Errorf("PR entry must be a string or mapping")
	}
}

func Load(path string) ([]prref.Ref, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := yaml.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	refs := make([]prref.Ref, 0, len(f.PRs))
	for i, entry := range f.PRs {
		ref, err := entry.RefValue()
		if err != nil {
			return nil, fmt.Errorf("prs[%d]: %w", i, err)
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func (e Entry) RefValue() (prref.Ref, error) {
	if strings.TrimSpace(e.Raw) != "" {
		return prref.Parse(e.Raw)
	}
	if strings.TrimSpace(e.URL) != "" {
		return prref.Parse(e.URL)
	}
	if strings.TrimSpace(e.Ref) != "" {
		return prref.Parse(e.Ref)
	}
	if strings.TrimSpace(e.Repo) != "" && e.Number > 0 {
		parts := strings.SplitN(e.Repo, "/", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
			return prref.Ref{}, fmt.Errorf("repo must be owner/name, got %q", e.Repo)
		}
		return prref.Ref{Owner: parts[0], Repo: parts[1], Number: e.Number}, nil
	}
	return prref.Ref{}, fmt.Errorf("PR entry needs string ref, url/ref, or repo+number")
}
