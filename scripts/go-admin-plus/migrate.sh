#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

profile=${1:-server-sqlite}
require_tool go
config_file=$(profile_config "$profile")
cd "$backend_root"
if test "$profile" = server-sqlite; then
  exec go run -tags sqlite3 . migrate --config "$config_file"
fi
exec go run . migrate --config "$config_file"
