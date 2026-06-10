package rollout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type docsctlSettings struct {
	workspace      string
	include        []string
	exclude        []string
	packages       []string
	commands       []string
	exportCommands []string
	output         string
	timeout        time.Duration
	version        string
}

type docsctlCandidate struct {
	Repo          string `json:"repo" yaml:"repo"`
	Path          string `json:"path" yaml:"path"`
	PackageName   string `json:"package_name" yaml:"package_name"`
	CmdDir        string `json:"cmd_dir" yaml:"cmd_dir"`
	Workflow      string `json:"workflow" yaml:"workflow"`
	ExportCommand string `json:"export_command" yaml:"export_command"`
	SQLitePath    string `json:"sqlite_path" yaml:"sqlite_path"`
	VaultRole     string `json:"vault_role" yaml:"vault_role"`
	Status        string `json:"status,omitempty" yaml:"status,omitempty"`
	ExitCode      int    `json:"exit_code,omitempty" yaml:"exit_code,omitempty"`
	SQLiteSize    int64  `json:"sqlite_size,omitempty" yaml:"sqlite_size,omitempty"`
	Note          string `json:"note,omitempty" yaml:"note,omitempty"`
}

type docsctlPlan struct {
	Profile      string             `json:"profile" yaml:"profile"`
	Workspace    string             `json:"workspace" yaml:"workspace"`
	GeneratedAt  string             `json:"generated_at" yaml:"generated_at"`
	Repositories []docsctlCandidate `json:"repositories" yaml:"repositories"`
}

func newDocsctlCommand() (*cobra.Command, error) {
	s := &docsctlSettings{timeout: 90 * time.Second, version: "v0.0.0-local"}
	cmd := &cobra.Command{Use: "docsctl", Short: "Discover and plan docsctl publishing rollouts"}
	addDocsctlFlags := func(c *cobra.Command) {
		c.Flags().StringVar(&s.workspace, "workspace", "", "Workspace containing repositories")
		c.Flags().StringSliceVar(&s.include, "include", nil, "Repository names to include")
		c.Flags().StringSliceVar(&s.exclude, "exclude", nil, "Repository names to exclude")
		c.Flags().StringSliceVar(&s.packages, "package", nil, "Package override repo=package; repeatable")
		c.Flags().StringSliceVar(&s.commands, "cmd", nil, "Command selection repo=./cmd/name; repeatable")
		c.Flags().StringSliceVar(&s.exportCommands, "export-command", nil, "Export command override repo='shell command'; repeatable")
		c.Flags().StringVar(&s.output, "output", "table", "Output format: table, json, yaml")
		c.Flags().DurationVar(&s.timeout, "timeout", 90*time.Second, "Per-command validation timeout")
		c.Flags().StringVar(&s.version, "version", "v0.0.0-local", "Version string to pass to docsctl validate")
	}
	inventory := &cobra.Command{Use: "inventory", Short: "Inventory candidate docsctl help exporters", RunE: func(cmd *cobra.Command, args []string) error {
		candidates, err := docsctlInventory(s.workspace, s.include, s.exclude, s.packages, s.commands, s.exportCommands)
		if err != nil {
			return err
		}
		return writeDocsctlCandidates(candidates, s.output)
	}}
	validate := &cobra.Command{Use: "validate", Short: "Validate candidate docsctl help SQLite exports", RunE: func(cmd *cobra.Command, args []string) error {
		candidates, err := docsctlInventory(s.workspace, s.include, s.exclude, s.packages, s.commands, s.exportCommands)
		if err != nil {
			return err
		}
		validated := validateDocsctlCandidates(cmd.Context(), candidates, s.timeout, s.version)
		return writeDocsctlCandidates(validated, s.output)
	}}
	plan := &cobra.Command{Use: "plan", Short: "Emit a docsctl rollout plan for validated candidates", RunE: func(cmd *cobra.Command, args []string) error {
		candidates, err := docsctlInventory(s.workspace, s.include, s.exclude, s.packages, s.commands, s.exportCommands)
		if err != nil {
			return err
		}
		validated := validateDocsctlCandidates(cmd.Context(), candidates, s.timeout, s.version)
		selected := make([]docsctlCandidate, 0, len(validated))
		for _, c := range validated {
			if c.Status == "validate_ok" {
				selected = append(selected, c)
			}
		}
		p := docsctlPlan{Profile: "docsctl", Workspace: s.workspace, GeneratedAt: time.Now().Format(time.RFC3339), Repositories: selected}
		return writeDocsctlPlan(p, s.output)
	}}
	for _, c := range []*cobra.Command{inventory, validate, plan} {
		addDocsctlFlags(c)
		cmd.AddCommand(c)
	}
	return cmd, nil
}

func docsctlInventory(workspace string, include, exclude, packages, commands, exportCommands []string) ([]docsctlCandidate, error) {
	if workspace == "" {
		return nil, fmt.Errorf("--workspace is required")
	}
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return nil, err
	}
	includeSet := stringSet(include)
	excludeSet := stringSet(exclude)
	packageOverrides := assignmentMap(packages)
	commandOverrides := assignmentMap(commands)
	exportCommandOverrides := assignmentMap(exportCommands)
	var candidates []docsctlCandidate
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		repo := entry.Name()
		if len(includeSet) > 0 && !includeSet[repo] || excludeSet[repo] {
			continue
		}
		repoPath := filepath.Join(workspace, repo)
		if _, err := os.Stat(filepath.Join(repoPath, "go.mod")); err != nil {
			continue
		}
		moduleName := moduleNameFromGoMod(filepath.Join(repoPath, "go.mod"))
		mains, _ := filepath.Glob(filepath.Join(repoPath, "cmd", "*", "main.go"))
		for _, main := range mains {
			cmdDir := "./" + filepath.ToSlash(strings.TrimSuffix(strings.TrimPrefix(main, repoPath+string(os.PathSeparator)), string(os.PathSeparator)+"main.go"))
			if selected, ok := commandOverrides[repo]; ok && selected != cmdDir {
				continue
			}
			packageName := repo
			if moduleName != "" {
				packageName = filepath.Base(moduleName)
			}
			if override, ok := packageOverrides[repo]; ok {
				packageName = override
			}
			workflow := detectDocsWorkflow(repoPath)
			exportCommand := fmt.Sprintf("GOWORK=off go run %s help export --format sqlite --output-path .docsctl/help.sqlite", cmdDir)
			if override, ok := exportCommandOverrides[repo]; ok {
				exportCommand = override
			}
			candidates = append(candidates, docsctlCandidate{Repo: repo, Path: repoPath, PackageName: packageName, CmdDir: cmdDir, Workflow: workflow, ExportCommand: exportCommand, SQLitePath: ".docsctl/help.sqlite", VaultRole: "docsctl-" + packageName + "-publisher"})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Repo == candidates[j].Repo {
			return candidates[i].CmdDir < candidates[j].CmdDir
		}
		return candidates[i].Repo < candidates[j].Repo
	})
	return candidates, nil
}

func validateDocsctlCandidates(ctx context.Context, candidates []docsctlCandidate, timeout time.Duration, version string) []docsctlCandidate {
	out := make([]docsctlCandidate, len(candidates))
	copy(out, candidates)
	for i := range out {
		out[i] = validateDocsctlCandidate(ctx, out[i], timeout, version)
	}
	return out
}

func validateDocsctlCandidate(ctx context.Context, c docsctlCandidate, timeout time.Duration, version string) docsctlCandidate {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tmp, err := os.MkdirTemp("", "ggg-docsctl-*")
	if err != nil {
		c.Status, c.Note = "error", err.Error()
		return c
	}
	defer os.RemoveAll(tmp)
	sqlitePath := filepath.Join(tmp, "help.sqlite")
	exportCommand := c.ExportCommand
	if exportCommand == "" {
		exportCommand = fmt.Sprintf("GOWORK=off go run %s help export --format sqlite --output-path %s", c.CmdDir, shellQuote(sqlitePath))
	} else {
		exportCommand = strings.ReplaceAll(exportCommand, ".docsctl/help.sqlite", shellQuote(sqlitePath))
	}
	stdout, stderr, code := runShellInRepo(ctx, c.Path, exportCommand)
	c.ExitCode = code
	if code != 0 {
		c.Status, c.Note = "export_failed", trimNote(stderr+stdout)
		return c
	}
	info, err := os.Stat(sqlitePath)
	if err != nil || info.Size() == 0 {
		c.Status, c.Note = "export_failed", "SQLite export missing or empty"
		return c
	}
	c.SQLiteSize = info.Size()
	stdout, stderr, code = runInRepo(ctx, c.Path, "docsctl", "validate", "--file", sqlitePath, "--package", c.PackageName, "--version", version)
	c.ExitCode = code
	if code != 0 {
		c.Status, c.Note = "validate_failed", trimNote(stderr+stdout)
		return c
	}
	c.Status, c.Note = "validate_ok", trimNote(stdout)
	return c
}

func runShellInRepo(ctx context.Context, dir, command string) (string, string, int) {
	return runInRepo(ctx, dir, "bash", "-lc", command)
}

func runInRepo(ctx context.Context, dir, name string, args ...string) (string, string, int) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return stdout.String(), stderr.String(), 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return stdout.String(), stderr.String(), exitErr.ExitCode()
	}
	return stdout.String(), stderr.String(), 127
}

func writeDocsctlCandidates(candidates []docsctlCandidate, output string) error {
	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(candidates)
	case "yaml":
		b, err := yaml.Marshal(candidates)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	default:
		fmt.Printf("repo\tcmd\tworkflow\tstatus\tsize\tnote\n")
		for _, c := range candidates {
			fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\n", c.Repo, c.CmdDir, c.Workflow, c.Status, c.SQLiteSize, c.Note)
		}
		return nil
	}
}

func writeDocsctlPlan(plan docsctlPlan, output string) error {
	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}
	b, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(b)
	return err
}

func detectDocsWorkflow(repoPath string) string {
	for _, name := range []string{
		".github/workflows/publish-docs.yaml",
		".github/workflows/publish-docs.yml",
		".github/workflows/release.yaml",
		".github/workflows/release.yml",
	} {
		if _, err := os.Stat(filepath.Join(repoPath, name)); err == nil {
			return name
		}
	}
	return ""
}

func moduleNameFromGoMod(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "module" {
			return fields[1]
		}
	}
	return ""
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func stringSet(values []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				m[part] = true
			}
		}
	}
	return m
}

func assignmentMap(values []string) map[string]string {
	m := map[string]string{}
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
			if ok && strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
				m[strings.TrimSpace(key)] = strings.TrimSpace(value)
			}
		}
	}
	return m
}

func trimNote(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 240 {
		return s[:240]
	}
	return s
}
