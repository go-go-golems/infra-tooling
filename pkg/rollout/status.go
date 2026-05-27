package rollout

import (
	"context"
	"os"

	"github.com/go-go-golems/infra-tooling/pkg/ghclient"
	"github.com/go-go-golems/infra-tooling/pkg/prlist"
	"github.com/go-go-golems/infra-tooling/pkg/prready"
)

type StatusResult struct {
	Repo     Repo            `json:"repo" yaml:"repo"`
	BranchOK bool            `json:"branch_ok" yaml:"branch_ok"`
	PR       string          `json:"pr" yaml:"pr"`
	Ready    *prready.Report `json:"ready,omitempty" yaml:"ready,omitempty"`
	State    string          `json:"state" yaml:"state"`
	OK       bool            `json:"ok" yaml:"ok"`
	Message  string          `json:"message" yaml:"message"`
}

func Status(ctx context.Context, cfg Config) ([]StatusResult, error) {
	branches, err := BranchStatus(cfg)
	if err != nil {
		return nil, err
	}
	var reports []prready.Report
	if cfg.PullRequest.OutputPRs != "" {
		path := resolvePath(cfg.Workspace, cfg.PullRequest.OutputPRs)
		if _, err := os.Stat(path); err == nil {
			refs, err := prlist.Load(path)
			if err != nil {
				return nil, err
			}
			client := ghclient.Client{}
			for _, ref := range refs {
				report, err := client.Readiness(ctx, ref)
				if err != nil {
					report = prready.Report{PR: ref, URL: ref.URL(), State: "error", Terminal: true}
				}
				reports = append(reports, report)
			}
		}
	}
	reportByRepo := map[string]prready.Report{}
	for _, report := range reports {
		reportByRepo[report.PR.Repo] = report
	}
	results := make([]StatusResult, 0, len(branches))
	for _, branch := range branches {
		res := StatusResult{Repo: branch.Repo, BranchOK: branch.OK, State: "local", OK: branch.OK, Message: branch.Message}
		if report, ok := reportByRepo[branch.Repo.Name]; ok {
			res.PR = report.URL
			res.Ready = &report
			res.State = string(report.State)
			res.OK = branch.OK && report.OK
		}
		results = append(results, res)
	}
	return results, nil
}
