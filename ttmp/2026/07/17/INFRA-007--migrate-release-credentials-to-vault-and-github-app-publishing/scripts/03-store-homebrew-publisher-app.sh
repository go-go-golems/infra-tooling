#!/usr/bin/env bash
# Store a newly generated Homebrew publisher App private key directly in Vault.
# The script prints only Vault metadata and key names; it never prints the key.

set -euo pipefail

if [[ "$#" -ne 2 ]]; then
  echo "usage: $0 APP_ID PRIVATE_KEY_PEM_FILE" >&2
  exit 2
fi

app_id="$1"
private_key_file="$2"
secret_path="kv/ci/github/homebrew-go-go-go/release-publisher-app"

if [[ ! "${app_id}" =~ ^[0-9]+$ ]]; then
  echo "APP_ID must be the numeric GitHub App ID, not the client ID" >&2
  exit 2
fi
if [[ ! -f "${private_key_file}" ]]; then
  echo "private-key file does not exist: ${private_key_file}" >&2
  exit 2
fi

openssl pkey -in "${private_key_file}" -noout
vault kv put "${secret_path}" \
  app_id="${app_id}" \
  private_key=@"${private_key_file}" \
  >/dev/null

vault kv get -format=json "${secret_path}" | jq -r \
  '{version: .data.metadata.version, keys: (.data.data | keys)}'
