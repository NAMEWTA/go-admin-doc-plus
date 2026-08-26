#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool pnpm
cd "$frontend_root"
exec pnpm --filter @go-admin/contracts generate
