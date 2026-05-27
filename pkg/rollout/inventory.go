package rollout

import (
	"bufio"
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type InventoryOptions struct {
	RequireModules []string
	Base           string
}

type Repo struct {
	Name           string   `json:"name" yaml:"name"`
	Path           string   `json:"path" yaml:"path"`
	Module         string   `json:"module" yaml:"module"`
	GlazedVersion  string   `json:"glazed_version" yaml:"glazed_version"`
	HasMakefile    bool     `json:"has_makefile" yaml:"has_makefile"`
	LintTargets    []string `json:"lint_targets" yaml:"lint_targets"`
	HasWorkflows   bool     `json:"has_workflows" yaml:"has_workflows"`
	HasLefthook    bool     `json:"has_lefthook" yaml:"has_lefthook"`
	PackageDirs    []string `json:"package_dirs" yaml:"package_dirs"`
	CurrentBranch  string   `json:"current_branch" yaml:"current_branch"`
	AheadBase      int      `json:"ahead_base" yaml:"ahead_base"`
	DirtyTracked   bool     `json:"dirty_tracked" yaml:"dirty_tracked"`
	DirtyUntracked bool     `json:"dirty_untracked" yaml:"dirty_untracked"`
}

func Inventory(root string, opts InventoryOptions) ([]Repo, error) {
	var goMods []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && shouldSkipDir(path, d.Name()) && path != root {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "go.mod" {
			goMods = append(goMods, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(goMods)
	var repos []Repo
	for _, gomod := range goMods {
		dir := filepath.Dir(gomod)
		module, requires, err := parseGoMod(gomod)
		if err != nil {
			return nil, err
		}
		if !matchesRequired(module, requires, opts.RequireModules) {
			continue
		}
		repo := InspectRepo(dir, opts.Base)
		repo.Module = module
		repo.GlazedVersion = requires["github.com/go-go-golems/glazed"]
		repos = append(repos, repo)
	}
	return repos, nil
}

func InspectRepo(dir string, base string) Repo {
	repo := Repo{Name: filepath.Base(dir), Path: dir}
	makefile := filepath.Join(dir, "Makefile")
	repo.HasMakefile = fileExists(makefile)
	if repo.HasMakefile {
		repo.LintTargets = makefileTargets(makefile, []string{"lint", "lintmax", "glazed-lint", "glazed-lint-build"})
	}
	repo.HasWorkflows = dirExists(filepath.Join(dir, ".github", "workflows"))
	repo.HasLefthook = fileExists(filepath.Join(dir, "lefthook.yml")) || fileExists(filepath.Join(dir, ".lefthook.yml"))
	repo.PackageDirs = packageDirs(dir)
	repo.CurrentBranch = gitOutput(dir, "rev-parse", "--abbrev-ref", "HEAD")
	if repo.CurrentBranch == "HEAD" {
		repo.CurrentBranch = "detached"
	}
	if base == "" {
		base = "origin/main"
	}
	repo.AheadBase = gitAhead(dir, base)
	repo.DirtyTracked, repo.DirtyUntracked = gitDirty(dir)
	return repo
}

func shouldSkipDir(path, name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", ".cache", ".bin", "dist", "build":
		return true
	}
	return strings.HasPrefix(name, ".") && name != ".github"
}

func parseGoMod(path string) (string, map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()
	module := ""
	requires := map[string]string{}
	scanner := bufio.NewScanner(f)
	inRequire := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		line = strings.Split(line, "//")[0]
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "module ") {
			module = strings.Fields(line)[1]
			continue
		}
		if line == "require (" {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if strings.HasPrefix(line, "require ") {
			fields := strings.Fields(strings.TrimPrefix(line, "require "))
			if len(fields) >= 2 {
				requires[fields[0]] = fields[1]
			}
			continue
		}
		if inRequire {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				requires[fields[0]] = fields[1]
			}
		}
	}
	return module, requires, scanner.Err()
}

func matchesRequired(module string, requires map[string]string, needles []string) bool {
	for _, needle := range needles {
		needle = strings.TrimSpace(needle)
		if needle == "" {
			continue
		}
		if module == needle {
			continue
		}
		if _, ok := requires[needle]; !ok {
			return false
		}
	}
	return true
}

var targetRE = regexp.MustCompile(`^([A-Za-z0-9_.-]+):`)

func makefileTargets(path string, wanted []string) []string {
	wantedSet := map[string]bool{}
	for _, w := range wanted {
		wantedSet[w] = true
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	found := map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		m := targetRE.FindStringSubmatch(scanner.Text())
		if m != nil && wantedSet[m[1]] {
			found[m[1]] = true
		}
	}
	out := make([]string, 0, len(found))
	for _, w := range wanted {
		if found[w] {
			out = append(out, w)
		}
	}
	return out
}

func packageDirs(root string) []string {
	set := map[string]bool{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && shouldSkipDir(path, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
			dir := filepath.Dir(path)
			rel, err := filepath.Rel(root, dir)
			if err == nil {
				if rel == "." {
					set["."] = true
				} else {
					set["./"+filepath.ToSlash(rel)] = true
				}
			}
		}
		return nil
	})
	out := make([]string, 0, len(set))
	for d := range set {
		out = append(out, d)
	}
	sort.Strings(out)
	return out
}

func gitOutput(dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitAhead(dir, base string) int {
	out := gitOutput(dir, "rev-list", "--count", base+"..HEAD")
	var n int
	_, _ = fmtSscanf(out, &n)
	return n
}

func gitDirty(dir string) (tracked bool, untracked bool) {
	out := gitOutput(dir, "status", "--porcelain")
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			untracked = true
		} else {
			tracked = true
		}
	}
	return tracked, untracked
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func fmtSscanf(s string, n *int) (int, error) {
	var sign, value int
	for _, r := range strings.TrimSpace(s) {
		if r < '0' || r > '9' {
			break
		}
		value = value*10 + int(r-'0')
		sign = 1
	}
	if sign == 0 {
		return 0, nil
	}
	*n = value
	return 1, nil
}
