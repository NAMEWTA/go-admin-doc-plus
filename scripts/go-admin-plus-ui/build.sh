#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool pnpm
case ${1:-all} in
  all)
    cd "$frontend_root"
    exec pnpm build:prod
    ;;
  web)
    cd "$frontend_root"
    exec pnpm --filter @go-admin-plus/admin-web build
    ;;
  desktop)
    require_desktop_workspace
    cd "$frontend_root"
    exec pnpm --filter @go-admin-plus/admin-desktop build
    ;;
  *)
    fail "unsupported frontend build target: ${1:-} (expected all, web, or desktop)"
    ;;
esac
