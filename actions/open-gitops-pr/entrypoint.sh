#!/usr/bin/env bash
set -euo pipefail

cmd=(python3 /action/src/gitops_pr_action/open_gitops_pr.py)
cmd+=(--config "${INPUT_CONFIG}")
cmd+=(--image "${INPUT_IMAGE}")

if [[ -n "${INPUT_TARGET:-}" ]]; then
  cmd+=(--target "${INPUT_TARGET}")
fi

if [[ "${INPUT_ALL_TARGETS:-true}" == "true" ]]; then
  cmd+=(--all-targets)
fi

if [[ "${INPUT_DRY_RUN:-false}" == "true" ]]; then
  cmd+=(--dry-run)
fi

if [[ "${INPUT_PUSH:-true}" == "true" ]]; then
  cmd+=(--push)
fi

if [[ "${INPUT_OPEN_PR:-true}" == "true" ]]; then
  cmd+=(--open-pr)
fi

if [[ -n "${INPUT_GITOPS_REPO_DIR:-}" ]]; then
  cmd+=(--gitops-repo-dir "${INPUT_GITOPS_REPO_DIR}")
fi

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  cmd+=(--github-output "${GITHUB_OUTPUT}")
fi

if [[ -n "${INPUT_GITHUB_TOKEN:-}" ]]; then
  export GH_TOKEN="${INPUT_GITHUB_TOKEN}"
fi

if [[ -n "${INPUT_GIT_AUTHOR_NAME:-}" ]]; then
  export GITOPS_PR_GIT_AUTHOR_NAME="${INPUT_GIT_AUTHOR_NAME}"
fi

if [[ -n "${INPUT_GIT_AUTHOR_EMAIL:-}" ]]; then
  export GITOPS_PR_GIT_AUTHOR_EMAIL="${INPUT_GIT_AUTHOR_EMAIL}"
fi

exec "${cmd[@]}"
