#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF' >&2
Usage:
  bootstrap_federation_source_repo_from_terraform.sh <owner/repo> <PUBLIC_BASE_URL_VAR> [platform_version] [bucket_name]

Example:
  bootstrap_federation_source_repo_from_terraform.sh \
    go-go-golems/go-go-app-sqlite \
    SQLITE_FEDERATION_PUBLIC_BASE_URL \
    0.1.0-canary.5

This script reads the Hetzner object-storage credentials from the Terraform
environment at $TERRAFORM_ROOT and seeds the standard federation object-storage
GitHub secrets for the target repository.

It also sets:
  - the remote public base URL variable named by <PUBLIC_BASE_URL_VAR>
  - GO_GO_OS_PLATFORM_VERSION when [platform_version] is provided

It does not create GITOPS_PR_TOKEN or K3S_REPO_READ_TOKEN. Those still need to
be provisioned separately.
EOF
}

if [ "$#" -lt 2 ] || [ "$#" -gt 4 ]; then
  usage
  exit 1
fi

TARGET_REPO="$1"
PUBLIC_BASE_URL_VAR="$2"
PLATFORM_VERSION="${3:-}"
BUCKET_NAME="${4:-${FEDERATION_BUCKET_NAME:-scapegoat-federation-assets}}"
TERRAFORM_ROOT="${TERRAFORM_ROOT:-/home/manuel/code/wesen/terraform}"
AWS_PROFILE_VALUE="${AWS_PROFILE:-manuel}"

cd "$TERRAFORM_ROOT"

TARGET_REPO="$TARGET_REPO" \
PUBLIC_BASE_URL_VAR="$PUBLIC_BASE_URL_VAR" \
PLATFORM_VERSION="$PLATFORM_VERSION" \
BUCKET_NAME="$BUCKET_NAME" \
AWS_PROFILE="$AWS_PROFILE_VALUE" \
direnv exec . bash -lc '
set -euo pipefail

: "${TF_VAR_object_storage_server:?missing TF_VAR_object_storage_server}"
: "${TF_VAR_object_storage_region:?missing TF_VAR_object_storage_region}"
: "${TF_VAR_object_storage_access_key:?missing TF_VAR_object_storage_access_key}"
: "${TF_VAR_object_storage_secret_key:?missing TF_VAR_object_storage_secret_key}"

repo="${TARGET_REPO:?missing TARGET_REPO}"
public_base_url_var="${PUBLIC_BASE_URL_VAR:?missing PUBLIC_BASE_URL_VAR}"
platform_version="${PLATFORM_VERSION:-}"
bucket_name="${BUCKET_NAME:?missing BUCKET_NAME}"
endpoint="https://${TF_VAR_object_storage_server}"
region="${TF_VAR_object_storage_region}"
public_base_url="https://${bucket_name}.${TF_VAR_object_storage_server}"

printf "%s" "${TF_VAR_object_storage_access_key}" | gh secret set HETZNER_OBJECT_STORAGE_ACCESS_KEY_ID --repo "${repo}"
printf "%s" "${TF_VAR_object_storage_secret_key}" | gh secret set HETZNER_OBJECT_STORAGE_SECRET_ACCESS_KEY --repo "${repo}"
printf "%s" "${bucket_name}" | gh secret set HETZNER_OBJECT_STORAGE_BUCKET --repo "${repo}"
printf "%s" "${endpoint}" | gh secret set HETZNER_OBJECT_STORAGE_ENDPOINT --repo "${repo}"
printf "%s" "${region}" | gh secret set HETZNER_OBJECT_STORAGE_REGION --repo "${repo}"

gh variable set "${public_base_url_var}" --repo "${repo}" --body "${public_base_url}"

if [ -n "${platform_version}" ]; then
  gh variable set GO_GO_OS_PLATFORM_VERSION --repo "${repo}" --body "${platform_version}"
fi

echo "repo=${repo}"
echo "public_base_url_var=${public_base_url_var}"
echo "bucket_name=${bucket_name}"
echo "endpoint=${endpoint}"
echo "region=${region}"
echo "public_base_url=${public_base_url}"
if [ -n "${platform_version}" ]; then
  echo "platform_version=${platform_version}"
fi
'
