#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

target=${1:-all}
profile=${2:-server-sqlite}

build_server() {
  require_tool go
  validate_profile "$1"
  mkdir -p "$artifacts_root/build"
  if test "$1" = server-sqlite; then
    (cd "$backend_root" && go build -tags sqlite3 -o "$artifacts_root/build/go-admin" .)
  else
    (cd "$backend_root" && go build -o "$artifacts_root/build/go-admin" .)
  fi
}

case $target in
  all)
    build_server "$profile"
    "$repo_root/scripts/go-admin-plus-ui/build.sh" web
    ;;
  server|server-sqlite|server-postgres)
    test "$target" = server || profile=$target
    build_server "$profile"
    ;;
  web|desktop)
    exec "$repo_root/scripts/go-admin-plus-ui/build.sh" "$target"
    ;;
  *)
    fail "unsupported build target: $target (expected all, server, web, or desktop)"
    ;;
esac
