#!/usr/bin/env bash
set -euo pipefail

DOCSYNC_BIN="$1"
GO_MOD_SYMLINK="$2"

REPO_ROOT="$(cd "$(dirname "$(readlink "$GO_MOD_SYMLINK")")" && pwd)"
exec "$DOCSYNC_BIN" validate --config "$REPO_ROOT/docsync.yml"
