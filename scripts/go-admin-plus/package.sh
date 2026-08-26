#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

target=${1:-}
profile=${2:-server-sqlite}

case $target in
  server|server-sqlite|server-postgres)
    require_tool go
    test "$target" = server || profile=$target
    validate_profile "$profile"
    os=$(go env GOOS)
    arch=$(go env GOARCH)
    mkdir -p "$artifacts_root/packages"
    if test "$profile" = server-sqlite; then
      exec sh -c 'cd "$1" && go build -tags sqlite3 -trimpath -o "$2/go-admin-$3-$4" .' sh \
        "$backend_root" "$artifacts_root/packages" "$os" "$arch"
    fi
    exec sh -c 'cd "$1" && go build -trimpath -o "$2/go-admin-$3-$4" .' sh \
      "$backend_root" "$artifacts_root/packages" "$os" "$arch"
    ;;
  web|desktop)
    exec "$repo_root/scripts/go-admin-plus-ui/package.sh" "$target"
    ;;
  *)
    fail "unsupported package target: $target (expected server, web, or desktop)"
    ;;
esac
