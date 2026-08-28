#!/bin/sh

script_dir=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/../.." && pwd)
frontend_root=$repo_root/go-admin-plus-ui

fail() {
  printf '%s\n' "go-admin-plus-ui task: $*" >&2
  exit 1
}

require_tool() {
  command -v "$1" >/dev/null 2>&1 || fail "required tool is not installed: $1"
}

require_pnpm() {
  command -v pnpm >/dev/null 2>&1 && return 0
  command -v corepack >/dev/null 2>&1 || fail 'required tool is not installed: pnpm or Corepack'
  pnpm_shim_root=$artifacts_root/tool-shims/corepack
  mkdir -p "$pnpm_shim_root"
  corepack enable pnpm --install-directory "$pnpm_shim_root" >/dev/null
  PATH=$pnpm_shim_root${PATH:+:$PATH}
  export PATH
  command -v pnpm >/dev/null 2>&1 || fail 'Corepack did not provide the pnpm shim'
}

run_pnpm() {
  require_pnpm
  pnpm "$@"
}

exec_pnpm() {
  require_pnpm
  exec pnpm "$@"
}

resolve_repo_path() {
  case $1 in
    /*|[A-Za-z]:[\\/]*) printf '%s\n' "$1" ;;
    *) printf '%s/%s\n' "$repo_root" "$1" ;;
  esac
}

artifacts_root=$(resolve_repo_path "${GO_ADMIN_ARTIFACTS_DIR:-.artifacts}")

require_desktop_workspace() {
  test -f "$frontend_root/apps/admin-desktop/src-tauri/tauri.conf.json" ||
    fail 'Tauri 2 desktop workspace is not available'
}
