#!/usr/bin/env bash
# Verify the cross-repository Sqleton release contract without reading Vault
# secrets or contacting GitHub. Run from any directory:
#
#   bash ttmp/.../scripts/check_sqleton_release_contract.sh \
#     /path/to/infra-tooling /path/to/terraform /path/to/sqleton

set -euo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: $0 INFRA_TOOLING_ROOT TERRAFORM_ROOT SQLETON_ROOT" >&2
  exit 2
fi

infra_root="$1"
terraform_root="$2"
sqleton_root="$3"

shared_workflow="${infra_root}/.github/workflows/publish-goreleaser-release.yml"
terraform_main="${terraform_root}/vault/github-actions/envs/k3s/main.tf"
sqleton_workflow="${sqleton_root}/.github/workflows/release.yml"
app_verifier_workflow="${sqleton_root}/.github/workflows/verify-homebrew-publisher-app.yml"

for file in "${shared_workflow}" "${terraform_main}" "${sqleton_workflow}" "${app_verifier_workflow}"; do
  if [[ ! -f "${file}" ]]; then
    echo "required file is missing: ${file}" >&2
    exit 1
  fi
done

require() {
  local description="$1"
  local pattern="$2"
  local file="$3"
  if ! rg -q --fixed-strings "${pattern}" "${file}"; then
    echo "contract violation: ${description}" >&2
    echo "  expected ${pattern} in ${file}" >&2
    exit 1
  fi
}

forbid() {
  local description="$1"
  local pattern="$2"
  local file="$3"
  if rg -q "${pattern}" "${file}"; then
    echo "contract violation: ${description}" >&2
    echo "  forbidden pattern ${pattern} in ${file}" >&2
    exit 1
  fi
}

# Caller workflow: build roles are deliberately distinct from publisher roles.
require "Sqleton only releases v-prefixed tags" '      - "v*"' "${sqleton_workflow}"
require "Linux build uses the builder role" 'role: release-sqleton-builder' "${sqleton_workflow}"
require "publisher call uses the publisher role" 'vault_role: release-sqleton-publisher' "${sqleton_workflow}"
require "Linux split limits GoReleaser to Linux" 'GGOOS: linux' "${sqleton_workflow}"
require "Darwin split limits GoReleaser to Darwin" 'GGOOS: darwin' "${sqleton_workflow}"
require "caller invokes the shared publisher" 'uses: go-go-golems/infra-tooling/.github/workflows/publish-goreleaser-release.yml@main' "${sqleton_workflow}"
forbid "caller retains a long-lived release secret" 'secrets\.(RELEASE_ACTION_PAT|GO_GO_GOLEMS_SIGN_KEY|GO_GO_GOLEMS_SIGN_PASSPHRASE|COSIGN_PWD|FURY_TOKEN)' "${sqleton_workflow}"

# Shared publisher: fixed paths and a fixed Homebrew target prevent workflow
# inputs from being turned into a generic Vault or GitHub-App proxy.
require "shared workflow validates the supported profile" 'homebrew-fury) ;;' "${shared_workflow}"
require "shared workflow fixes the Homebrew target" 'go-go-golems/homebrew-go-go-go' "${shared_workflow}"
require "shared workflow reads the GoReleaser license" 'kv/data/ci/release/shared/goreleaser-pro license_key | GORELEASER_KEY' "${shared_workflow}"
require "shared workflow mints an App token" 'uses: actions/create-github-app-token@v2' "${shared_workflow}"
require "shared workflow merges split artifacts" 'args: continue --merge --config=${{ inputs.goreleaser_config }}' "${shared_workflow}"

# Terraform: isolate the build policy text and verify it cannot read publisher
# credentials. The publisher policy must explicitly allow all three paths.
builder_policy="$(sed -n '/resource "vault_policy" "release_build"/,/resource "vault_jwt_auth_backend_role" "release_build"/p' "${terraform_main}")"
publisher_policy="$(sed -n '/resource "vault_policy" "release_publish"/,/resource "vault_jwt_auth_backend_role" "release_publish"/p' "${terraform_main}")"
profile_map="$(sed -n '/release_credential_profiles = {/,/release_publishers = {/p' "${terraform_main}")"

if [[ "${builder_policy}" != *'kv/data/ci/release/shared/goreleaser-pro'* ]]; then
  echo "contract violation: builder policy cannot read the GoReleaser license" >&2
  exit 1
fi
if [[ "${builder_policy}" == *'homebrew-go-go-go'* || "${builder_policy}" == *'fury-go-go-golems'* ]]; then
  echo "contract violation: builder policy can read publisher credentials" >&2
  exit 1
fi
if [[ "${publisher_policy}" != *'local.release_credential_profiles[each.value.profile].secret_paths'* ]]; then
  echo "contract violation: publisher policy does not derive paths from the allowlisted profile map" >&2
  exit 1
fi
for path in \
  'kv/data/ci/release/shared/goreleaser-pro' \
  'kv/data/ci/github/homebrew-go-go-go/release-publisher-app' \
  'kv/data/ci/release/shared/fury-go-go-golems'; do
  if [[ "${profile_map}" != *"${path}"* ]]; then
    echo "contract violation: publisher profile is missing ${path}" >&2
    exit 1
  fi
done

require "publisher role binds the immutable Sqleton repository id" 'repository_id    = each.value.repository_id' "${terraform_main}"
require "publisher role binds the reusable workflow" 'job_workflow_ref = var.release_publish_job_workflow_ref' "${terraform_main}"

# The App verifier may read the App key but not the release license or Fury
# token. It proves the App's selected-repository scope through a temporary
# branch create/delete cycle.
verifier_policy="$(sed -n '/resource "vault_policy" "release_app_verifier"/,/resource "vault_jwt_auth_backend_role" "release_app_verifier"/p' "${terraform_main}")"
if [[ "${verifier_policy}" != *'kv/data/ci/github/homebrew-go-go-go/release-publisher-app'* ]]; then
  echo "contract violation: App verifier cannot read the publisher App credential" >&2
  exit 1
fi
if [[ "${verifier_policy}" == *'goreleaser-pro'* || "${verifier_policy}" == *'fury-go-golems'* ]]; then
  echo "contract violation: App verifier can read unrelated release credentials" >&2
  exit 1
fi
require "App verifier uses the dedicated Vault role" 'role: release-sqleton-homebrew-app-verifier' "${app_verifier_workflow}"
require "App verifier scopes the App token to the Homebrew tap" 'repositories: homebrew-go-go-go' "${app_verifier_workflow}"
require "App verifier registers branch cleanup" 'trap cleanup EXIT' "${app_verifier_workflow}"

echo "Sqleton release contract: PASS"
