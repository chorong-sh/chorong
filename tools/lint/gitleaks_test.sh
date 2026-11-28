#!/usr/bin/env bash
set -euo pipefail

GITLEAKS_BIN="$1"
CONFIG_FILE="$2"
GO_MOD_SYMLINK="$3"

REPO_ROOT="$(cd "$(dirname "$(readlink "$GO_MOD_SYMLINK")")" && pwd)"

if ! git -C "$REPO_ROOT" rev-parse --verify HEAD >/dev/null 2>&1; then
  # Before the initial commit exists, scan the working tree instead of git history.
  exec "$GITLEAKS_BIN" dir --no-banner --redact --config "$CONFIG_FILE" "$REPO_ROOT"
fi

exec "$GITLEAKS_BIN" git --no-banner --redact --config "$CONFIG_FILE" "$REPO_ROOT"
