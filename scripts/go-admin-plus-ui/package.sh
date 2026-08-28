#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

target=${1:-}
case $target in
  web)
    require_tool pnpm
    cd "$frontend_root"
    pnpm build:prod
    mkdir -p "$artifacts_root/packages"
    tar -C "$frontend_root" -czf "$artifacts_root/packages/go-admin-web.tar.gz" dist
    ;;
  desktop)
    require_tool pnpm
    require_desktop_workspace
    cd "$frontend_root"
    node "$repo_root/release/shared/sidecar/build.mjs" --host
    exec pnpm --filter @go-admin-plus/admin-desktop tauri build
    ;;
  *)
    fail "unsupported frontend package target: $target (expected web or desktop)"
    ;;
esac
