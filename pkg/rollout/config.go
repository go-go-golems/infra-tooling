package rollout

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ID            string        `yaml:"id"`
	Name          string        `yaml:"name"`
	Workspace     string        `yaml:"workspace"`
	Branch        string        `yaml:"branch"`
	Base          string        `yaml:"base"`
	CommitMessage string        `yaml:"commit_message"`
	Selection     Selection     `yaml:"selection"`
	Validation    Validation    `yaml:"validation"`
	PullRequest   PullRequest   `yaml:"pull_request"`
	Readiness     Readiness     `yaml:"readiness"`
	Release       ReleaseConfig `yaml:"release"`
}

type Selection struct {
	RequireGoModContains []string       `yaml:"require_go_mod_contains"`
	Include              []string       `yaml:"include"`
	Exclude              []ExcludedRepo `yaml:"exclude"`
}

type ExcludedRepo struct {
	Repo   string `yaml:"repo"`
	Reason string `yaml:"reason"`
}

type Validation struct {
	Commands        []ValidationCommand `yaml:"commands"`
	ContinueOnError bool                `yaml:"continue_on_error"`
	LogDir          string              `yaml:"log_dir"`
}

type ValidationCommand struct {
	Name string `yaml:"name"`
	Run  string `yaml:"run"`
}

type PullRequest struct {
	Title        string `yaml:"title"`
	BodyFile     string `yaml:"body_file"`
	OutputPRs    string `yaml:"output_prs"`
	NoVerifyPush bool   `yaml:"no_verify_push"`
}

type Readiness struct {
	TriggerCodex bool `yaml:"trigger_codex"`
	Watch        bool `yaml:"watch"`
}

type ReleaseConfig struct {
	Mode                  string `yaml:"mode"`
	RequireManualApproval bool   `yaml:"require_manual_approval"`
}

func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return Config{}, err
	}
	if cfg.Workspace == "" {
		return Config{}, fmt.Errorf("rollout config %s is missing workspace", path)
	}
	if cfg.Base == "" {
		cfg.Base = "origin/main"
	}
	return cfg, nil
}

func SaveConfig(path string, cfg Config) error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

func (c Config) ResolveTargets() ([]string, error) {
	workspace := c.Workspace
	if workspace == "" {
		return nil, fmt.Errorf("workspace is required")
	}
	excluded := map[string]bool{}
	for _, e := range c.Selection.Exclude {
		excluded[strings.TrimSpace(e.Repo)] = true
	}
	var targets []string
	for _, name := range c.Selection.Include {
		name = strings.TrimSpace(name)
		if name == "" || excluded[name] {
			continue
		}
		if filepath.IsAbs(name) {
			targets = append(targets, filepath.Clean(name))
		} else {
			targets = append(targets, filepath.Join(workspace, name))
		}
	}
	if len(targets) > 0 {
		return targets, nil
	}
	repos, err := Inventory(workspace, InventoryOptions{RequireModules: c.Selection.RequireGoModContains, Base: c.Base})
	if err != nil {
		return nil, err
	}
	for _, repo := range repos {
		if excluded[repo.Name] || excluded[repo.Module] {
			continue
		}
		targets = append(targets, repo.Path)
	}
	return targets, nil
}
