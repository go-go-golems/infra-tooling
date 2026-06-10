# Tasks

## TODO

- [ ] Push branches and open/update PRs where repository remotes allow it.
- [ ] After PR merges, tag releases and verify production docs URLs.
- [ ] Expand minimal embedded docs for `llm-proxy`, `logcopter`, and `chat-overlay` in future follow-up PRs.

## DONE

- [x] Create ticket workspace and initial design/diary documents.
- [x] Initial inventory script created and run against workspace packages.
- [x] Keep a frequent diary with commands, failures, commits, PR/push evidence, and validation notes.
- [x] Improve `ggg rollout docsctl` where it reduces manual rollout risk (module-derived package names, export command overrides, command-faithful validation).
- [x] Add or repair local Glazed help export support for packages whose CLIs could not emit SQLite help: `docmgr`, `llm-proxy`, `logcopter`, and `chat-overlay`.
- [x] Add `.github/workflows/publish-docs.yaml` or keep release-coupled `publish-docs` jobs for every rollout package: `devctl`, `docmgr`, `llm-proxy`, `logcopter`, `chat-overlay`, `remarquee`, `scraper`, `sessionstream`, and `vm-system`.
- [x] Keep existing docsctl setup for `goja-bleve` and verify it as a baseline.
- [x] Add Vault/Terraform docsctl publisher roles for missing packages under `/home/manuel/code/wesen/terraform/vault/github-actions/envs/k3s` with numeric repository IDs and exact workflow refs.
- [x] Run `terraform fmt`, `terraform plan`, `terraform apply`, and record a clean post-apply plan.
- [x] For each package, validate: `GOWORK=off go run <cmd> help export --format sqlite --output-path .docsctl/help.sqlite`, `test -s`, and `docsctl validate --package <package> --version v0.0.0-local` via `ggg rollout docsctl plan`.
- [x] Run package unit/smoke tests appropriate to touched repos.
- [x] Commit changes at focused intervals per repo and infra/tooling area for package repositories.
- [x] Write the intern-facing analysis/design/implementation guide with file references, diagrams, rollout tables, pseudocode, and validation instructions.
- [x] Validate docmgr ticket hygiene with `docmgr doctor --ticket INFRA-006 --stale-after 30`.
- [x] Upload the final ticket bundle to reMarkable at `/ai/2026/06/10/INFRA-006/INFRA-006 docsctl rollout guide.pdf`.
