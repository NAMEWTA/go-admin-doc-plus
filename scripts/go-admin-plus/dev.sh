#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

target=${1:-}
profile=${2:-server-sqlite}

case $target in
  server|server-sqlite|server-postgres)
    require_tool go
    test "$target" = server || profile=$target
    config_file=$(profile_config "$profile")
    cd "$backend_root"
    if test "$profile" = server-sqlite; then
      exec go run -tags sqlite3 . server --config "$config_file"
    fi
    exec go run . server --config "$config_file"
    ;;
  web)
    exec "$repo_root/scripts/go-admin-plus-ui/dev.sh" web
    ;;
  desktop)
    exec "$repo_root/scripts/go-admin-plus-ui/dev.sh" desktop
    ;;
  *)
    fail "unsupported development target: $target (expected server, web, or desktop)"
    ;;
esac
