package rollout

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const ProfileGlazedLint = "glazed-lint"

type PlanOptions struct {
	Profile string
}

type PlanOperation struct {
	Repo        Repo   `json:"repo" yaml:"repo"`
	Profile     string `json:"profile" yaml:"profile"`
	File        string `json:"file" yaml:"file"`
	Kind        string `json:"kind" yaml:"kind"`
	Status      string `json:"status" yaml:"status"`
	Description string `json:"description" yaml:"description"`
	Detail      string `json:"detail" yaml:"detail"`
}

func Plan(cfg Config, opts PlanOptions) ([]PlanOperation, error) {
	profile := opts.Profile
	if profile == "" {
		profile = ProfileGlazedLint
	}
	switch profile {
	case ProfileGlazedLint:
		return PlanGlazedLint(cfg)
	default:
		return nil, fmt.Errorf("unsupported rollout profile %q", profile)
	}
}

func PlanGlazedLint(cfg Config) ([]PlanOperation, error) {
	targets, err := cfg.ResolveTargets()
	if err != nil {
		return nil, err
	}
	var ops []PlanOperation
	for _, target := range targets {
		repo := InspectRepo(target, cfg.Base)
		module, requires, _ := parseGoMod(filepath.Join(target, "go.mod"))
		repo.Module = module
		repo.GlazedVersion = requires["github.com/go-go-golems/glazed"]
		repoOps, err := planGlazedLintRepo(repo)
		if err != nil {
			return ops, err
		}
		ops = append(ops, repoOps...)
	}
	return ops, nil
}

func planGlazedLintRepo(repo Repo) ([]PlanOperation, error) {
	makefilePath := filepath.Join(repo.Path, "Makefile")
	var ops []PlanOperation
	if !fileExists(makefilePath) {
		return []PlanOperation{planOp(repo, "Makefile", "makefile", "needed", "Create Makefile before Glazed lint wiring can be planned", "Makefile is missing")}, nil
	}
	b, err := os.ReadFile(makefilePath)
	if err != nil {
		return nil, err
	}
	mk := string(b)
	selfHostedGlazed := repo.Module == "github.com/go-go-golems/glazed"
	requiredVars := []string{"GLAZED_LINT_BIN", "GLAZED_LINT_FLAGS", "GLAZED_LINT_DIRS"}
	if !selfHostedGlazed {
		requiredVars = append([]string{"GLAZED_LINT_BIN", "GLAZED_LINT_PKG", "GLAZED_VERSION", "GLAZED_LINT_TOOL_VERSION", "GLAZED_LINT_FLAGS", "GLAZED_LINT_DIRS"})
	}
	for _, name := range requiredVars {
		status := "present"
		detail := name + " is defined"
		if !makeVarDefined(mk, name) {
			status = "needed"
			detail = name + " is missing"
		}
		ops = append(ops, planOp(repo, "Makefile", "makefile-variable", status, "Ensure "+name+" is defined", detail))
	}

	buildBody, buildOK := makeTargetBody(mk, "glazed-lint-build")
	if !buildOK {
		ops = append(ops, planOp(repo, "Makefile", "makefile-target", "needed", "Add glazed-lint-build target", "target is missing"))
	} else {
		ops = append(ops, planOp(repo, "Makefile", "makefile-target", "present", "glazed-lint-build target exists", "target is present"))
		if selfHostedGlazed {
			ops = append(ops, checkContains(repo, "Makefile", "tool-build-local", buildBody, "go build -o $(GLAZED_LINT_BIN) ./cmd/tools/glazed-lint", "Build glazed-lint from the local Glazed checkout"))
		} else {
			ops = append(ops, checkContains(repo, "Makefile", "tool-install-workspace", buildBody, "GOWORK=off go install", "Build glazed-lint outside ambient workspaces"))
			ops = append(ops, checkContains(repo, "Makefile", "tool-install-version", buildBody, "$(GLAZED_LINT_TOOL_VERSION)", "Use an explicit fallback/tool version for glazed-lint"))
			if strings.Contains(buildBody, "@latest") {
				ops = append(ops, planOp(repo, "Makefile", "tool-install-version", "needed", "Remove @latest glazed-lint install fallback", "target still contains @latest"))
			} else {
				ops = append(ops, planOp(repo, "Makefile", "tool-install-version", "present", "No @latest glazed-lint install fallback", "target does not contain @latest"))
			}
		}
	}

	lintBody, lintOK := makeTargetBody(mk, "glazed-lint")
	if !lintOK {
		ops = append(ops, planOp(repo, "Makefile", "makefile-target", "needed", "Add glazed-lint target", "target is missing"))
	} else {
		ops = append(ops, planOp(repo, "Makefile", "makefile-target", "present", "glazed-lint target exists", "target is present"))
		ops = append(ops, checkContains(repo, "Makefile", "vet-workspace", lintBody, "GOWORK=off go vet", "Run standalone glazed-lint outside ambient workspaces"))
		ops = append(ops, checkContains(repo, "Makefile", "vet-dirs", lintBody, "$(GLAZED_LINT_DIRS)", "Use GLAZED_LINT_DIRS in standalone glazed-lint target"))
		ops = append(ops, checkContains(repo, "Makefile", "vet-flags", lintBody, "$(GLAZED_LINT_FLAGS)", "Use GLAZED_LINT_FLAGS in standalone glazed-lint target"))
	}

	for _, target := range []string{"lint", "lintmax"} {
		body, ok := makeTargetBody(mk, target)
		if !ok {
			ops = append(ops, planOp(repo, "Makefile", "makefile-target", "warning", "Inspect missing "+target+" target", "target is missing"))
			continue
		}
		if strings.Contains(body, "$(GLAZED_LINT_BIN)") || strings.Contains(body, "glazed-lint") {
			ops = append(ops, planOp(repo, "Makefile", "lint-integration", "present", "Integrate Glazed lint into "+target, "target references glazed-lint"))
			if strings.Contains(body, "$(GLAZED_LINT_BIN)") {
				ops = append(ops, checkContains(repo, "Makefile", "vet-workspace", body, "GOWORK=off go vet", "Run Glazed vettool outside ambient workspaces in "+target))
				ops = append(ops, checkContains(repo, "Makefile", "vet-dirs", body, "$(GLAZED_LINT_DIRS)", "Use GLAZED_LINT_DIRS in "+target))
			}
		} else {
			ops = append(ops, planOp(repo, "Makefile", "lint-integration", "needed", "Integrate Glazed lint into "+target, "target does not reference glazed-lint"))
		}
	}

	ops = append(ops, planWorkflow(repo)...)
	return ops, nil
}

func planWorkflow(repo Repo) []PlanOperation {
	workflowDir := filepath.Join(repo.Path, ".github", "workflows")
	if !dirExists(workflowDir) {
		return []PlanOperation{planOp(repo, ".github/workflows", "ci-workflow", "warning", "Inspect missing GitHub workflow directory", "workflow directory is missing")}
	}
	matches, _ := filepath.Glob(filepath.Join(workflowDir, "*.yml"))
	matches2, _ := filepath.Glob(filepath.Join(workflowDir, "*.yaml"))
	matches = append(matches, matches2...)
	sort.Strings(matches)
	if len(matches) == 0 {
		return []PlanOperation{planOp(repo, ".github/workflows", "ci-workflow", "warning", "Inspect missing GitHub workflow files", "no workflow yaml files found")}
	}
	for _, path := range matches {
		b, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		if strings.Contains(string(b), "make glazed-lint") {
			rel, _ := filepath.Rel(repo.Path, path)
			return []PlanOperation{planOp(repo, filepath.ToSlash(rel), "ci-workflow", "present", "Run Glazed CLI policy linters in CI", "workflow runs make glazed-lint")}
		}
	}
	return []PlanOperation{planOp(repo, ".github/workflows", "ci-workflow", "needed", "Add CI step to run make glazed-lint", "no workflow runs make glazed-lint")}
}

func planOp(repo Repo, file, kind, status, description, detail string) PlanOperation {
	return PlanOperation{Repo: repo, Profile: ProfileGlazedLint, File: file, Kind: kind, Status: status, Description: description, Detail: detail}
}

func checkContains(repo Repo, file, kind, body, needle, description string) PlanOperation {
	if strings.Contains(body, needle) {
		return planOp(repo, file, kind, "present", description, "found "+needle)
	}
	return planOp(repo, file, kind, "needed", description, "missing "+needle)
}

func makeVarDefined(makefile, name string) bool {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(name) + `\s*(?:\?|:)?=`)
	return re.MatchString(makefile)
}

func makeTargetBody(makefile, target string) (string, bool) {
	lines := strings.Split(makefile, "\n")
	re := regexp.MustCompile(`^` + regexp.QuoteMeta(target) + `\s*:`)
	anyTarget := regexp.MustCompile(`^[A-Za-z0-9_.-]+\s*:`)
	start := -1
	for i, line := range lines {
		if re.MatchString(line) {
			start = i
			break
		}
	}
	if start == -1 {
		return "", false
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "\t") && !strings.HasPrefix(line, " ") && anyTarget.MatchString(line) {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n"), true
}
