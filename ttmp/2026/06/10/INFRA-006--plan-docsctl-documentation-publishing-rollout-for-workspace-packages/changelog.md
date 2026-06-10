# Changelog

## 2026-06-10

- Initial workspace created


## 2026-06-10

Implemented docsctl publishing rollout workflows, help-export fixes, ggg validation improvements, and Vault publisher roles (commits: infra-tooling f465744, terraform 87eeed8, package commits recorded in diary).

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/design-doc/01-workspace-docsctl-publishing-rollout-implementation-guide.md — Intern-facing implementation guide
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/reference/01-diary.md — Detailed chronological implementation diary


## 2026-06-10

Validated INFRA-006 with docmgr doctor and uploaded the guide/diary/tasks/changelog bundle to reMarkable at /ai/2026/06/10/INFRA-006.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/reference/01-diary.md — Records doctor and reMarkable upload evidence


## 2026-06-10

Pushed rollout branches and opened PRs for all package repos, infra-tooling, and Terraform; recorded docmgr pre-push release-snapshot asset failure.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/reference/01-diary.md — PR and push evidence


## 2026-06-10

Removed workflow_dispatch from all separate docs publishing workflows so docsctl publish jobs are tag-push-only and match Vault claim bindings.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/reference/01-diary.md — Records review-response workflow trigger cleanup


## 2026-06-10

Updated docmgr release-coupled publish-docs job to run only on v* tag pushes and grant job-level OIDC permissions; pushed commit c4e0268 to PR #41.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/docmgr/.github/workflows/release.yml — Docmgr docs job gating and OIDC permissions


## 2026-06-10

Updated llm-proxy PR #3 to define server flags through a Glazed serve command and added make glazed-lint to CI; pushed commits f344257 and 9378925.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/llm-proxy/.github/workflows/push.yml — CI glazed-lint check
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/llm-proxy/Makefile — glazed-lint targets
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/llm-proxy/cmd/llm-proxy-server/main.go — Glazed serve command implementation


## 2026-06-10

Addressed logcopter PR #3 review comments by restoring exported doc.FS, embedding tutorials as well as topics, and loading Glazed help from the exported FS; pushed commit 602c6ca.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/logcopter/pkg/doc/doc.go — Preserves public embedded docs filesystem and docsctl help loading


## 2026-06-10

Verified rollout branches were pushed; found docmgr PR #41 merged before late gating fix, so opened follow-up PR #42 with the missing release-tag docs job gating.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/reference/01-diary.md — Records branch/merge verification and docmgr follow-up PR


## 2026-06-10

Addressed infra-tooling PR #16 review by rewriting full docsctl --output-path arguments during validation; regenerated plan now includes goja-bleve validate_ok; pushed commit 30d75be.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/internal/cli/rollout/docsctl.go — Nested output-path rewrite fix
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/internal/cli/rollout/docsctl_test.go — Regression tests for nested and quoted output paths
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/infra-tooling/ttmp/2026/06/10/INFRA-006--plan-docsctl-documentation-publishing-rollout-for-workspace-packages/sources/02-ggg-docsctl-plan.yaml — Regenerated validation plan including goja-bleve


## 2026-06-10

Analyzed and fixed logcopter PR #3 CI failures: restored legacy generator flag parsing for normal runs, kept Cobra help export, upgraded jsonparser, and bumped Go directive to 1.25.11; pushed commits 4286c3c and bc3e60b.

### Related Files

- /home/manuel/workspaces/2026-06-10/add-docs-deploy/logcopter/cmd/logcopter-gen/main.go — Dual legacy generator and Cobra help-export parsing
- /home/manuel/workspaces/2026-06-10/add-docs-deploy/logcopter/go.mod — Patched Go directive and jsonparser version for CI security checks

