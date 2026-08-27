#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool pnpm
case ${1:-} in
  web)
    cd "$frontend_root"
    exec pnpm dev
    ;;
  desktop)
    require_desktop_workspace
    cd "$frontend_root"
    node "$repo_root/release/shared/sidecar/build.mjs" --host
    exec pnpm --filter @go-admin/admin-desktop tauri dev
    ;;
  *)
    fail "unsupported frontend development target: ${1:-} (expected web or desktop)"
    ;;
esac
