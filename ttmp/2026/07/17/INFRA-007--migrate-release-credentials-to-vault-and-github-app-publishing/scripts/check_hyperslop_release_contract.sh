#!/usr/bin/env bash
# Verify the Hyperslop Vault-backed release contract without reading secrets.
# Usage: check_hyperslop_release_contract.sh INFRA_TOOLING_ROOT TERRAFORM_ROOT HYPERSLOP_ROOT

set -euo pipefail

if [[ "$#" -ne 3 ]]; then
  echo "usage: $0 INFRA_TOOLING_ROOT TERRAFORM_ROOT HYPERSLOP_ROOT" >&2
  exit 2
fi

infra_root="$1"
terraform_root="$2"
hyperslop_root="$3"
shared_workflow="${infra_root}/.github/workflows/publish-goreleaser-release.yml"
terraform_main="${terraform_root}/vault/github-actions/envs/k3s/main.tf"
hyperslop_workflow="${hyperslop_root}/.github/workflows/release.yaml"

for file in "${shared_workflow}" "${terraform_main}" "${hyperslop_workflow}"; do
  test -f "${file}" || { echo "required file is missing: ${file}" >&2; exit 1; }
done

require() {
  local description="$1" pattern="$2" file="$3"
  if ! rg -q --fixed-strings "${pattern}" "${file}"; then
    echo "contract violation: ${description}" >&2
    echo "  expected ${pattern} in ${file}" >&2
    exit 1
  fi
}

forbid() {
  local description="$1" pattern="$2" file="$3"
  if rg -q "${pattern}" "${file}"; then
    echo "contract violation: ${description}" >&2
    echo "  forbidden ${pattern} in ${file}" >&2
    exit 1
  fi
}

require "release trigger is limited to version tags" "      - 'v*'" "${hyperslop_workflow}"
require "Linux build uses the scoped builder role" "role: release-hyperslop-cli-builder" "${hyperslop_workflow}"
require "Darwin build uses the scoped builder role" "role: release-hyperslop-cli-builder" "${hyperslop_workflow}"
require "merge uses the scoped publisher role" "vault_role: release-hyperslop-cli-publisher" "${hyperslop_workflow}"
require "merge invokes the shared publisher" "uses: go-go-golems/infra-tooling/.github/workflows/publish-goreleaser-release.yml@main" "${hyperslop_workflow}"
require "merge chooses the approved Hyperslop profile" "credential_profile: hyperslop-homebrew-fury-gpg" "${hyperslop_workflow}"
require "build jobs request OIDC" "id-token: write" "${hyperslop_workflow}"
forbid "caller retains legacy release secrets" 'secrets\.(GORELEASER_KEY|GO_GO_GOLEMS_SIGN_KEY|GO_GO_GOLEMS_SIGN_PASSPHRASE|COSIGN_PWD|HOMEBREW_TAP_TOKEN|FURY_TOKEN)' "${hyperslop_workflow}"

require "shared workflow allows the Hyperslop profile" "hyperslop-homebrew-fury-gpg)" "${shared_workflow}"
require "shared workflow fixes the Hyperslop Homebrew target" "homebrew_owner=hyperslop-systems" "${shared_workflow}"
require "shared workflow fixes the Hyperslop Homebrew repository" "homebrew_repository=homebrew" "${shared_workflow}"
require "shared workflow reads scoped GPG credentials only for the GPG profile" "Read GPG signing credentials" "${shared_workflow}"

profile_map="$(sed -n '/release_credential_profiles = {/,/release_publishers = {/p' "${terraform_main}")"
for path in \
  'kv/data/ci/release/shared/goreleaser-pro' \
  'kv/data/ci/github/hyperslop-systems/homebrew-publisher-app' \
  'kv/data/ci/release/hyperslop-systems/fury' \
  'kv/data/ci/release/hyperslop-systems/gpg-signing'; do
  if [[ "${profile_map}" != *"${path}"* ]]; then
    echo "contract violation: Hyperslop profile is missing ${path}" >&2
    exit 1
  fi
done

require "Terraform records the immutable Hyperslop repository ID" 'repository_id     = "1314775506"' "${terraform_main}"
require "Terraform records the Hyperslop release workflow" 'hyperslop-systems/hyperslop-cli/.github/workflows/release.yaml@refs/tags/v*' "${terraform_main}"
require "roles bind each publisher's declared owner" 'repository_owner = each.value.repository_owner' "${terraform_main}"
require "publisher role binds the shared workflow" 'job_workflow_ref = var.release_publish_job_workflow_ref' "${terraform_main}"

echo "Hyperslop release contract: PASS"
