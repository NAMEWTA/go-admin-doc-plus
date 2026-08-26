#!/bin/sh

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
frontend_root=$repo_root/go-admin-ui-plus
artifacts_root=${GO_ADMIN_ARTIFACTS_DIR:-$repo_root/.artifacts}

fail() {
  printf '%s\n' "go-admin-plus-ui task: $*" >&2
  exit 1
}

require_tool() {
  command -v "$1" >/dev/null 2>&1 || fail "required tool is not installed: $1"
}

require_desktop_workspace() {
  test -f "$frontend_root/apps/admin-desktop/src-tauri/tauri.conf.json" ||
    fail 'Tauri 2 desktop workspace is not available; T-16 owns this target'
}
