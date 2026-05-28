# Changelog

## 2026-05-28

- Initial workspace created


## 2026-05-28

Created INFRA-004 ticket, initial rollout implementation guide, batch planner script, and generated batch artifacts from INFRA-003 inventory.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/analysis/01-rollout-analysis-and-implementation-guide.md — Initial rollout guide


## 2026-05-28

Added SQLite rollout tracker, CLI update verbs, auto-refresh dashboard, initialized current B2 PR progress, and started dashboard in tmux session infra004-dashboard.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/scripts/02-rollout-tracker.py — Progress tracker implementation


## 2026-05-28

Advanced B2 logcopter-only wave: merged oak-git-db, go-go-agent-action, go-go-app-arc-agi, and salad with merge commits; verified or classified main actions; recorded voyage blocked and barbar skipped edge cases in SQLite tracker.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/reference/01-diary.md — Detailed Step 3 rollout diary
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Current rollout progress state


## 2026-05-28

Completed B2 safe-mechanical triage: all remaining B2 repos are now blocked/skipped in tracker pending manual decisions about stdlib logging conflicts, archived/external modules, or placeholder module identity.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/ttmp/2026/05/28/INFRA-004--batch-infra-003-follow-up-rollout-across-go-go-golems-repos/sources/05-rollout-progress.sqlite — Updated B2 triage states


## 2026-05-28

Updated ggg action status behavior so no_runs is terminal/non-blocking under --watch, exits 0, reports ok=true, and no longer keeps batch action watches pending; installed updated ggg to ~/.local/bin.

### Related Files

- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/internal/cli/batch/actions.go — Batch action watch loop now only waits on pending
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/internal/cli/run/status.go — Single-repo watch loop now only waits on pending
- /home/manuel/workspaces/2026-05-24/add-js-providers/infra-tooling/pkg/actionstatus/actionstatus.go — State and exit-code semantics for no_runs
