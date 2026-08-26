#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool pnpm
case ${1:-web} in
  web)
    cd "$frontend_root"
    exec pnpm build:prod
    ;;
  desktop)
    require_desktop_workspace
    cd "$frontend_root"
    exec pnpm --filter @go-admin/desktop tauri build
    ;;
  *)
    fail "unsupported frontend build target: ${1:-} (expected web or desktop)"
    ;;
esac
