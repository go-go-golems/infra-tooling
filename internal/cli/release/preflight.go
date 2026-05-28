package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type preflightSettings struct {
	Repo    string
	Output  string
	Strict  bool
	Verbose bool
}

type preflightFinding struct {
	Severity string `json:"severity" yaml:"severity"`
	Code     string `json:"code" yaml:"code"`
	Message  string `json:"message" yaml:"message"`
	File     string `json:"file,omitempty" yaml:"file,omitempty"`
	Hint     string `json:"hint,omitempty" yaml:"hint,omitempty"`
}

type preflightResult struct {
	OK       bool               `json:"ok" yaml:"ok"`
	Repo     string             `json:"repo" yaml:"repo"`
	Findings []preflightFinding `json:"findings" yaml:"findings"`
}

func newPreflightCommand() *cobra.Command {
	s := &preflightSettings{Repo: ".", Output: "table"}
	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Check for common release-tag failures before pushing a tag",
		RunE: func(cmd *cobra.Command, args []string) error {
			res := runPreflight(s)
			if err := writePreflightResult(res, s.Output); err != nil {
				return err
			}
			if !res.OK {
				return fmt.Errorf("release preflight failed for %s", res.Repo)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&s.Repo, "repo", s.Repo, "Repository directory")
	cmd.Flags().StringVar(&s.Output, "output", s.Output, "Output format: table, json, yaml")
	cmd.Flags().BoolVar(&s.Strict, "strict", false, "Treat warnings as failures")
	cmd.Flags().BoolVar(&s.Verbose, "verbose", false, "Include passing informational findings")
	return cmd
}

func runPreflight(s *preflightSettings) preflightResult {
	repo, err := filepath.Abs(s.Repo)
	if err != nil {
		repo = s.Repo
	}
	res := preflightResult{Repo: repo, OK: true}
	add := func(sev, code, msg, file, hint string) {
		res.Findings = append(res.Findings, preflightFinding{Severity: sev, Code: code, Message: msg, File: file, Hint: hint})
		if sev == "error" || (sev == "warning" && s.Strict) {
			res.OK = false
		}
	}
	goreleaserPath := firstExisting(repo, ".goreleaser.yaml", ".goreleaser.yml")
	if goreleaserPath == "" {
		add("warning", "missing_goreleaser", "No .goreleaser.yaml/.yml file found", "", "Skip if this repository intentionally does not use GoReleaser.")
	} else {
		checkGoReleaserPreflight(repo, goreleaserPath, add)
	}
	checkGenerateFrontendPreflight(repo, goreleaserPath, add)
	checkDocsctlWorkflowPreflight(repo, add)
	if len(res.Findings) == 0 && s.Verbose {
		add("info", "preflight_clean", "No release preflight findings", "", "")
	}
	return res
}

func checkGoReleaserPreflight(repo, path string, add func(string, string, string, string, string)) {
	bodyBytes, err := os.ReadFile(path)
	if err != nil {
		add("error", "read_goreleaser", err.Error(), path, "Ensure the GoReleaser config is readable.")
		return
	}
	body := string(bodyBytes)
	if regexp.MustCompile(`\bXXX\b`).MatchString(body) {
		add("error", "goreleaser_placeholder", "GoReleaser config still contains scaffold placeholder XXX", path, "Replace project_name, binary, main, descriptions, and homepage placeholders before tagging.")
	}
	for _, mainPath := range goreleaserMainPaths(body) {
		clean := strings.TrimPrefix(mainPath, "./")
		if _, err := os.Stat(filepath.Join(repo, clean)); err != nil {
			add("error", "goreleaser_missing_main", "GoReleaser main path does not exist: "+mainPath, path, "Point main to an existing command directory such as ./cmd/<binary>.")
		}
	}
	if strings.Contains(body, "CGO_ENABLED=0") && moduleMentions(repo, "tree-sitter") {
		add("error", "cgo_disabled_with_tree_sitter", "GoReleaser disables CGO but go.mod references tree-sitter packages", path, "Enable CGO for release builds and configure cross-compilers where needed.")
	}
}

func goreleaserMainPaths(body string) []string {
	var paths []string
	re := regexp.MustCompile(`(?m)^\s*(?:-\s*)?main:\s*['"]?([^'"\s#]+)`)
	for _, m := range re.FindAllStringSubmatch(body, -1) {
		paths = append(paths, strings.TrimSpace(m[1]))
	}
	return paths
}

func checkGenerateFrontendPreflight(repo, goreleaserPath string, add func(string, string, string, string, string)) {
	if goreleaserPath == "" {
		return
	}
	bodyBytes, _ := os.ReadFile(goreleaserPath)
	body := string(bodyBytes)
	if !strings.Contains(body, "go generate ./...") {
		return
	}
	frontendDirs := findFrontendPackageDirs(repo)
	if len(frontendDirs) == 0 {
		return
	}
	releaseWorkflow := firstExisting(repo, ".github/workflows/release.yaml", ".github/workflows/release.yml")
	workflowBody := ""
	if releaseWorkflow != "" {
		b, _ := os.ReadFile(releaseWorkflow)
		workflowBody = string(b)
	}
	if !strings.Contains(workflowBody, "pnpm/action-setup") && !strings.Contains(workflowBody, "corepack") {
		add("warning", "generate_frontend_without_pnpm_setup", "GoReleaser runs go generate and frontend package(s) exist, but release workflow does not set up pnpm", releaseWorkflow, "Add pnpm/action-setup before GoReleaser or remove frontend generation from release hooks.")
	}
	for _, dir := range frontendDirs {
		rel, _ := filepath.Rel(repo, dir)
		if !strings.Contains(workflowBody, "pnpm --dir "+filepath.ToSlash(rel)+" install") && !strings.Contains(workflowBody, "working-directory: "+filepath.ToSlash(rel)) {
			add("warning", "generate_frontend_without_install", "Frontend package may need dependencies before go generate: "+filepath.ToSlash(rel), releaseWorkflow, "Install dependencies with `pnpm --dir "+filepath.ToSlash(rel)+" install --frozen-lockfile` before GoReleaser.")
		}
	}
}

func checkDocsctlWorkflowPreflight(repo string, add func(string, string, string, string, string)) {
	workflow := firstExisting(repo, ".github/workflows/release.yaml", ".github/workflows/release.yml")
	if workflow == "" {
		return
	}
	b, _ := os.ReadFile(workflow)
	body := string(b)
	if !strings.Contains(body, "publish-docs") || !strings.Contains(body, "publish-docsctl.yml") {
		return
	}
	if strings.Contains(body, "id-token: write") && !regexp.MustCompile(`(?s)publish-docs:\s.*?permissions:\s.*?id-token:\s*write`).MatchString(body) {
		add("warning", "oidc_not_job_scoped", "Release workflow has id-token: write but not clearly scoped to publish-docs job", workflow, "Keep id-token: write on the publish-docs job, not workflow-wide.")
	}
	for _, key := range []string{"package_name", "package_version", "export_command", "vault_role", "vault_token_role"} {
		if !strings.Contains(body, key+":") {
			add("error", "docsctl_missing_"+strings.ReplaceAll(key, "_", "_"), "publish-docs job is missing `"+key+"`", workflow, "Complete the reusable docsctl workflow inputs before tagging.")
		}
	}
}

func findFrontendPackageDirs(repo string) []string {
	var dirs []string
	_ = filepath.WalkDir(repo, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if base == ".git" || base == "node_modules" || base == "dist" {
			return filepath.SkipDir
		}
		if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
			if _, err := os.Stat(filepath.Join(path, "pnpm-lock.yaml")); err == nil {
				dirs = append(dirs, path)
			}
		}
		return nil
	})
	return dirs
}

func moduleMentions(repo, needle string) bool {
	for _, name := range []string{"go.mod", "go.sum"} {
		b, err := os.ReadFile(filepath.Join(repo, name))
		if err == nil && strings.Contains(string(b), needle) {
			return true
		}
	}
	return false
}

func firstExisting(repo string, names ...string) string {
	for _, name := range names {
		path := filepath.Join(repo, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func writePreflightResult(res preflightResult, output string) error {
	switch output {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(res)
	case "yaml":
		b, err := yaml.Marshal(res)
		if err != nil {
			return err
		}
		_, err = os.Stdout.Write(b)
		return err
	default:
		fmt.Printf("ok\trepo\tfindings\n")
		fmt.Printf("%t\t%s\t%d\n", res.OK, res.Repo, len(res.Findings))
		for _, f := range res.Findings {
			fmt.Printf("%s\t%s\t%s\t%s\n", f.Severity, f.Code, f.File, f.Message)
			if f.Hint != "" {
				fmt.Printf("hint\t%s\n", f.Hint)
			}
		}
		return nil
	}
}
