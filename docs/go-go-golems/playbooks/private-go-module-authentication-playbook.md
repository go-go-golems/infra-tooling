---
Title: Give a CI workflow read access to a private Go module
Slug: private-go-module-authentication-playbook
Short: How a GitHub Actions workflow authenticates to private github.com/hyperslop-systems Go modules — the shared-workflow profile, the local composite action, and which to choose.
Topics:
- github-actions
- github-apps
- go
- goprivate
- vault
- oidc
- ci-cd
- go-go-golems
Commands:
- gh
- vault
- go
- git
Flags:
- private_dependencies_profile
- GOPRIVATE
- vault_role
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

# Give a CI workflow read access to a private Go module

Use this playbook when a Go repository imports a module from a **private** GitHub
repository — today that means `github.com/hyperslop-systems/pbui` and
`github.com/hyperslop-systems/hyperslop-cli` — and its CI fails while resolving that
import.

The symptom is a `go mod download`, `go build` or `go test` failure naming the private
module. Two shapes are common: a `git ls-remote` failure carrying `remote: Repository not
found`, and a checksum-database rejection while verifying the module. Both mean the same
thing — the runner has no credential for that repository, and Go has not been told to skip
the public proxy and checksum database for it.

## What the credential actually is

Not a PAT, and not `GITHUB_TOKEN`. `GITHUB_TOKEN` is scoped to the repository running the
workflow and cannot read a different private repository, which is why this needs its own
mechanism.

The chain is the same one the GitOps PR automation uses:

```text
GitHub Actions OIDC token (workflow JWT)
  -> Vault auth/github-actions, role bound to this exact repository
  -> read app_id + private_key from kv/ci/github/hyperslop-systems/private-dependencies-app
  -> actions/create-github-app-token mints an installation token scoped to
     owner=hyperslop-systems, repositories=pbui,hyperslop-cli
  -> git config url.insteadOf rewrites https://github.com/ to embed that token
  -> GOPRIVATE + GONOSUMDB tell Go to bypass the proxy and sumdb for those modules
```

Every step matters, and the last one is the one people forget. The `insteadOf` rewrite alone
is not enough: without `GOPRIVATE`, Go still routes the module through `proxy.golang.org`,
which cannot see it, so the rewrite never comes into play.

`GOPRIVATE` is the one that does the work — `go help environment` defines it as the default
for both `GONOPROXY` and `GONOSUMDB`, so setting it alone would suffice. The workflow sets
`GONOSUMDB` explicitly as well, which is redundant but harmless and makes the intent
readable.

The token is minted per run and expires with it.

## Two implementations, both live

They do exactly the same four steps. The difference is where the configuration lives.

| | Shared-workflow profile | Local composite action |
|---|---|---|
| Where | `infra-tooling/.github/workflows/publish-ghcr-image.yml` | `<repo>/.github/actions/setup-private-go/action.yml` |
| Selected by | `private_dependencies_profile: <name>` input | `uses: ./.github/actions/setup-private-go` |
| Configuration | a `case` arm in the shared workflow, reviewed centrally | inline in the consuming repo |
| Applies to | only the shared publish workflow | any workflow in that repo |
| Changing the App or its repo list | one PR to `infra-tooling` | one PR per consuming repo |

`datalab` uses **both**, and that is not an accident to copy blindly: the shared workflow
covers `publish-image.yaml`, and the composite action covers `push.yml`, `lint.yml`,
`codeql-analysis.yml`, `dependency-scanning.yml` and `release.yaml` — every other workflow
that compiles Go. The shared workflow cannot help those, because they do not call it.

**Choose by how many workflows need it.** One workflow, and it is the shared publish
workflow: use the profile. More than one: add the composite action, and use the profile as
well if the repo also publishes images.

## Option A — the shared-workflow profile

Profiles are an allowlist, not a free-form input. The shared workflow rejects an unknown
name rather than silently skipping authentication:

```bash
*)
  echo "Unsupported private_dependencies_profile: ${PROFILE}" >&2
  exit 1
  ;;
```

So adding a repository means a PR to `infra-tooling` adding an arm to that `case`, next to
the existing one:

```bash
datalab-go)
  echo "enabled=true" >> "$GITHUB_OUTPUT"
  echo "vault_role=datalab-private-dependencies" >> "$GITHUB_OUTPUT"
  echo "secret_path=kv/data/ci/github/hyperslop-systems/private-dependencies-app" >> "$GITHUB_OUTPUT"
  echo "owner=hyperslop-systems" >> "$GITHUB_OUTPUT"
  echo "repositories=pbui,hyperslop-cli" >> "$GITHUB_OUTPUT"
  ;;
```

Only `vault_role` and `repositories` normally differ between arms. The App credential is
shared: one installed App identity, one KV path. What isolates each source repository is its
**Vault role**, whose bound claims name that repository alone.

The caller then passes one input:

```yaml
jobs:
  release:
    uses: go-go-golems/infra-tooling/.github/workflows/publish-ghcr-image.yml@main
    secrets: inherit
    with:
      private_dependencies_profile: <name>-go
      # ... the rest of the publish inputs
```

## Option B — the local composite action

Copy `datalab/.github/actions/setup-private-go/action.yml`, change the `role:` and, if the
module set differs, `repositories:`. Then in every workflow that compiles Go:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write        # required — no OIDC token without it
    steps:
      - uses: actions/checkout@v6
      - uses: ./.github/actions/setup-private-go
      - uses: actions/setup-go@v6
        with:
          go-version-file: go.mod
      - run: go test ./...
```

Order matters. `setup-private-go` must run **before** `actions/setup-go` if that step has
`cache: true`, because restoring the module cache resolves dependencies.

### Which jobs need it: more than the ones that obviously build

"Every workflow that compiles Go" undersells it. Anything that **typechecks** needs to
resolve the module, and most such tools do not look like builds. Onboarding turboproof left
four jobs red for this reason, and each reported it differently — only the first names the
cause:

| Job | What it says when the credential is missing |
|---|---|
| `go build` / `go test` | ``could not read Username for `https://github.com` `` |
| `golangci-lint` | `could not import <pkg> … (typecheck)`, **plus cascading errors that name innocent files** — an import reported as "imported and not used" while it is used a few lines down |
| `govulncheck` | `There are errors with the provided package patterns` |
| `gosec` | `package <pkg> has type errors, skipping SSA analysis, no ssa result` |
| CodeQL (Go) | fails during autobuild; datalab, agentlogic and turboproof all run the step in this job |
| `docker build` | needs the credential a different way — see the next section |

The golangci-lint case is the expensive one to debug, because the cascading errors are
plausible on their own. A reviewer chasing "unused import" will not find it, since the import
is genuinely used; the typecheck simply never got far enough to see the file properly.

### A missing credential can still go green

This is the part that turns a five-minute fix into an afternoon. A job with **no** credential
can pass, if a sibling job in the same workflow run has already warmed `setup-go`'s module
cache. The cache is keyed on `go.sum`, not on which job populated it, so a lint job with no
`id-token: write` can find the private module already downloaded by a test job that has it.

turboproof's `golangci-lint` passed on every pull request that way and then failed on the
first push to `main`, where no cache existed yet. Nothing about the workflow had changed.

The consequence for review: **a green lint job is not evidence that the job is configured
correctly.** Check the workflow file, not the run. The reliable test is a branch with a cold
cache, which in practice means the first push after the `go.sum` changes.

## The Docker build needs the credential separately

This is the step that catches people: the four steps above configure the **runner**, and a
`docker build` does not inherit them. BuildKit runs in its own context, so the runner's
global git configuration and `$GITHUB_ENV` are not visible inside it. Tests pass, then the
image build fails on the same module.

The shared workflow already handles the workflow half. When a profile is enabled it uses a
different build step that passes the minted token as a BuildKit secret:

```yaml
secrets: |
  github_token=${{ steps.private-dependency-token.outputs.token }}
```

The Dockerfile has to mount and use it. This is datalab's, and it is the pattern to copy:

```dockerfile
# syntax=docker/dockerfile:1
FROM golang:1.26.5-alpine AS build

WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN --mount=type=secret,id=github_token,required=true \
    PRIVATE_DEPENDENCY_TOKEN="$(cat /run/secrets/github_token)" \
    && git config --global \
        url."https://x-access-token:${PRIVATE_DEPENDENCY_TOKEN}@github.com/".insteadOf \
        "https://github.com/" \
    && GOPRIVATE='github.com/hyperslop-systems/*' \
       GONOSUMDB='github.com/hyperslop-systems/*' \
       go mod download \
    && git config --global --remove-section \
        "url.https://x-access-token:${PRIVATE_DEPENDENCY_TOKEN}@github.com/"
COPY . .
RUN CGO_ENABLED=0 GOWORK=off go build -o /out/<binary> ./cmd/<binary>
```

Four things that matter:

- **`# syntax=docker/dockerfile:1`** — `--mount=type=secret` needs it.
- **`required=true`** — fail loudly if the secret was not passed, rather than falling through
  to an anonymous fetch that produces a confusing `Repository not found` deep in the build.
- **The secret is mounted, never `COPY`ed or `ENV`-set.** A mounted secret is not written to
  any layer; a token in an `ENV` or an intermediate layer ships in the published image.
- **`git config --remove-section` in the same `RUN`.** The rewrite embeds the token in
  `/root/.gitconfig`, which *would* be committed to the layer. Removing it in the same
  instruction keeps it out. `go mod download` runs before the removal, and `go build` after
  it, using the module cache.

Local builds need the secret too:

```bash
gh auth token > /tmp/gh-token
DOCKER_BUILDKIT=1 docker build --secret id=github_token,src=/tmp/gh-token -t <image> .
rm /tmp/gh-token
```

## Vault objects

Both options need one Vault role per source repository. Follow the shape of the GitOps PR
roles.

Policy — `vault/policies/github-actions/<repo>-private-dependencies.hcl` in
`wesen/2026-03-27--hetzner-k3s`:

```hcl
# CI for <repo> may read the shared private-dependency App credential and nothing
# else. It cannot read application runtime secrets or GitOps PR credentials.
path "kv/data/ci/github/hyperslop-systems/private-dependencies-app" {
  capabilities = ["read"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}

path "auth/token/renew-self" {
  capabilities = ["update"]
}

path "auth/token/revoke-self" {
  capabilities = ["update"]
}
```

Role — `vault/roles/github-actions/<repo>-private-dependencies.json`:

```json
{
  "role_type": "jwt",
  "user_claim": "repository",
  "bound_audiences": ["https://vault.yolo.scapegoat.dev"],
  "bound_claims": {
    "repository_owner": "<source-owner>",
    "repository": "<source-owner>/<repo>"
  },
  "policies": ["gha-<repo>-private-dependencies"],
  "ttl": "10m",
  "max_ttl": "30m",
  "token_explicit_max_ttl": "30m"
}
```

`<source-owner>` is the owner of the **consuming** repository, which is not necessarily
`hyperslop-systems`. The OIDC `repository` and `repository_owner` claims identify the
repository whose workflow is running, while the App token's `owner:` identifies where the
private *dependency* lives. Those two differ whenever a `go-go-golems` or `wesen` repo
imports a `hyperslop-systems` module — `agentlogic` is `wesen/agentlogic` and consumes
`hyperslop-systems/pbui`.

Get it from the repository itself rather than assuming:

```bash
gh repo view <owner>/<repo> --json nameWithOwner --jq .nameWithOwner
```

A repository can exist under two owners at once — a fork and its upstream. Bind the one
whose Actions runs need the credential; for a pull request that is the repository the PR was
opened in.

> [!caution] This role lets pull-request code read the App private key
> The shared workflow reads `app_id` and `private_key` into the job environment, and *then*
> runs the repository's own tests:
>
> ```text
> 4  Read private dependency App credentials     <- private key enters the environment
> 5  Mint private dependency token
> 6  Configure private Go module authentication
> 7  Set up Go
> 8  Run tests                                   <- repository-controlled code
> ```
>
> Because the role below binds neither `ref` nor `event_name`, a pull request that changes a
> test can authenticate to this role and read `PRIVATE_DEPENDENCY_APP_PRIVATE_KEY` — the
> **long-lived App key**, not the scoped token. The `repositories:` input narrows only the
> token minted on the expected path; it does not constrain what the key itself can mint. A
> malicious branch could therefore mint tokens for every repository the App is installed on.
>
> This is a live weakness, not a hypothetical, and this playbook's own design enables it.
> Exposure equals the set of accounts with push access to the consuming repository — GitHub
> withholds `id-token` from fork pull requests, so it is not open to arbitrary outsiders. On
> a private repository with one or two trusted collaborators the practical risk is small; it
> grows the moment push access widens.
>
> Mitigations, in increasing order of cost:
>
> - **Give the App the narrowest installation possible** — `Contents: read` on exactly the
>   dependency repositories. This caps what a leaked key is worth.
> - **Gate PR builds behind a protected environment with required reviewers.** The
>   `environment` claim cannot be set by pull-request code, so binding it in the role and
>   requiring approval is the control that actually holds. It costs an approval per PR build.
> - **Do not give PR builds the credential at all** — run the private-module jobs only on
>   `push`, and accept that PRs cannot compile.
>
> Pinning `job_workflow_ref` is *not* an adequate substitute: on a pull request it carries
> `@refs/pull/N/merge`, so an exact pin breaks PR builds, and a glob still matches a pull
> request that edits the workflow file itself.

> [!important] Do not add `"ref"` or `"event_name"` to these bound claims
> The GitOps PR roles bind `ref: refs/heads/main` and `event_name: push`, because opening a
> deployment pull request is only ever legitimate from a trusted `main` push. A
> private-dependency role is different: it is used by `test` and `lint` on **pull requests**
> and branches. Copying the GitOps PR claims here makes every PR build fail at the Vault
> step with a claim mismatch, which reads like a misconfigured role rather than a
> deliberate restriction.

Apply with an operator token:

The script lives in the GitOps repository, not in `infra-tooling` where this playbook is
stored, so change into that checkout first:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
export VAULT_ADDR=https://vault.yolo.scapegoat.dev
vault login -method=oidc role=operators
bash scripts/bootstrap-vault-github-actions-oidc.sh
bash scripts/validate-vault-github-actions-oidc.sh
```

The validator knows about this role class: a `*-private-dependencies` role is checked for
repository binding and a policy, but is not required to pin `ref` or `event_name`.

The App must also be **installed** on the repositories listed in `repositories:`, with
`Contents: read`. A valid key for an App that is not installed on `pbui` mints a token that
authenticates and then cannot see the module.

## Validate

```bash
# The role exists and carries the right policy
vault read auth/github-actions/role/<repo>-private-dependencies

# The policy grants the App path and nothing else
vault policy read gha-<repo>-private-dependencies
```

Then push a branch and confirm in the run log, in order:

1. `Read GitHub App credentials from Vault` — succeeded, so OIDC and the policy are right.
2. `Mint private dependency token` — succeeded, so the App is installed on the target repos.
3. `go build` / `go test` resolves the private module.

## Failure modes

### `Repository not found` while resolving the module

The runner has no credential. Either the setup step did not run (check its `if:` and that
it is before `setup-go`), or `GOPRIVATE` is unset so Go went to the public proxy instead of
using the `insteadOf` rewrite.

### The checksum database rejects the module

`GOPRIVATE` does not match the module path, so `GONOSUMDB` does not inherit it. The
`insteadOf` rewrite is working; `sum.golang.org` is refusing a module it cannot fetch. Check
the glob: `github.com/hyperslop-systems/*` covers every repo in the org, but a narrower value
set by hand elsewhere in the workflow can shadow it.

### `permission denied` reading the Vault path

The policy is missing from Vault, or exists in git but was never applied. `vault policy read`
alone cannot tell those apart — both produce the same empty result. Check the file first:

```bash
cd /home/manuel/code/wesen/2026-03-27--hetzner-k3s
ls vault/policies/github-actions/<repo>-private-dependencies.hcl   # in git?
vault policy read gha-<repo>-private-dependencies                  # applied?
```

File absent → write it (§Vault objects). File present, policy absent or different → run the
bootstrap script. If both exist, diff them; a policy edited in git but never re-applied is
the case that looks correct in review and fails in CI.

### The workflow gets no Vault token at all

Either the job is missing `id-token: write`, or the role's bound claims do not match. If the
role was copied from a GitOps PR role, check for `ref` and `event_name` claims — see the
warning above.

### `Unsupported private_dependencies_profile`

The profile name is not in the shared workflow's `case`. It is an allowlist; add an arm in
`infra-tooling` rather than passing an arbitrary string.

> [!warning] Re-running the failed run will not pick up the new arm
> A workflow that calls a reusable workflow by branch (`…/publish-ghcr-image.yml@main`)
> resolves that ref **when the run starts**. Re-running an old run reuses the version resolved
> then, so a run that failed before your `infra-tooling` merge keeps failing afterwards with
> the identical message, however many times it is re-run.
>
> This reads exactly like "the merge did not work". It is not: only a **new** run — a fresh
> push, a new pull request event, or a `workflow_dispatch` — sees the new arm. Confirm by
> comparing the run's start time against the merge time before debugging anything else.

### The role does not exist, or names a repository that no longer pushes

Two variants, with two different Vault messages:

```text
{"errors":["role \"<repo>-private-dependencies\" could not be found"]}
{"errors":["error validating claims: claim \"repository\" does not match any associated bound claim values"]}
```

The first means the role was never created — its declaration file may exist in git while the
bootstrap script was never run, or neither exists. The second usually means the repository was
**renamed or moved between owners** after the role was written: `agentlogic` bound
`wesen/agentlogic` and kept failing after moving to `hyperslop-systems`, exactly as `datalab`
had after `go-go-golems/go-go-datadrop`. A rename leaves a role bound to a name that no longer
pushes, and nothing fails until the next build.

Check what is live rather than what is in git:

```bash
vault read -format=json auth/github-actions/role/<repo>-private-dependencies |
  jq '.data.bound_claims'
```

## Known gap

`datalab-private-dependencies` and the policy granting
`kv/ci/github/hyperslop-systems/private-dependencies-app` exist **only inside Vault**. They
are declared in neither `wesen/2026-03-27--hetzner-k3s/vault/` nor
`terraform/vault/github-actions/`, so they cannot be reviewed, rebuilt, or audited from
code, and a Vault restore would not recreate them.

Any new role added by following this playbook should be committed to
`wesen/2026-03-27--hetzner-k3s/vault/` as described above, and `datalab`'s should be
backfilled to match.

**Partially closed, 2026-07-31.** `turboproof-private-dependencies` and
`agentlogic-private-dependencies` are now declared in
`wesen/2026-03-27--hetzner-k3s/vault/{policies,roles}/github-actions/`, along with the two
matching `-gitops-pr` roles. `datalab`'s remain undeclared.

The gap is not theoretical. turboproof's role had **no file and no Vault object**, so the
repository never had a green build from the day it was created; the failure was a Vault
`role could not be found` that named the role but not the fact that nobody had ever made it.
Because a role that exists only in Vault is invisible to review, the absence of one is
invisible too.

## Related

- [github-app-gitops-pr-migration-playbook.md](./github-app-gitops-pr-migration-playbook.md) —
  the sibling credential, for opening GitOps pull requests. Same Vault auth path, same App
  pattern, different scope and different bound claims.
- [../../platform/source-repo-to-gitops-pr.md](../../platform/source-repo-to-gitops-pr.md) —
  the caller contract for the shared publish workflow. It now carries the
  `private_dependencies_profile` input and points here for the rest.
- `wesen/2026-03-27--hetzner-k3s/docs/github-actions-vault-oidc-playbook.md` — the
  GitHub Actions to Vault OIDC path this builds on.
