#!/bin/sh
set -eu
. "$(dirname "$0")/common.sh"

require_tool go
cd "$backend_root"
go test ./... -count=1
exec go test -tags sqlite ./... -count=1
