package rollout

import (
	"context"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds/schema"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	rolloutpkg "github.com/go-go-golems/infra-tooling/pkg/rollout"
)

func csv(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func decodeDefault(vals *values.Values, s any) error {
	return vals.DecodeSectionInto(schema.DefaultSlug, s)
}

func addRepoRow(ctx context.Context, gp middlewares.Processor, repo rolloutpkg.Repo) error {
	return gp.AddRow(ctx, types.NewRow(
		types.MRP("repo", repo.Name),
		types.MRP("path", repo.Path),
		types.MRP("module", repo.Module),
		types.MRP("glazed_version", repo.GlazedVersion),
		types.MRP("has_makefile", repo.HasMakefile),
		types.MRP("lint_targets", repo.LintTargets),
		types.MRP("has_workflows", repo.HasWorkflows),
		types.MRP("has_lefthook", repo.HasLefthook),
		types.MRP("package_dirs", repo.PackageDirs),
		types.MRP("current_branch", repo.CurrentBranch),
		types.MRP("ahead_base", repo.AheadBase),
		types.MRP("dirty_tracked", repo.DirtyTracked),
		types.MRP("dirty_untracked", repo.DirtyUntracked),
	))
}
