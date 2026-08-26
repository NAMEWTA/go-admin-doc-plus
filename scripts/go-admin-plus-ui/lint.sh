#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool pnpm
cd "$frontend_root"
pnpm lint
exec pnpm type-check
