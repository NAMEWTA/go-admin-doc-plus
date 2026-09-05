#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 1 ]] || { echo "usage: verify-dmg.sh <dmg>" >&2; exit 2; }
dmg=$1
[[ -f "$dmg" ]]
hdiutil verify "$dmg" >/dev/null
mountpoint=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/go-admin-macos-mount.XXXXXX")
mounted=false
cleanup() {
  if [[ "$mounted" == true ]]; then hdiutil detach -quiet "$mountpoint" || true; fi
  rmdir "$mountpoint" 2>/dev/null || true
}
trap cleanup EXIT
hdiutil attach -noverify -readonly -nobrowse -mountpoint "$mountpoint" "$dmg"
mounted=true
app=$(find "$mountpoint" -maxdepth 1 -type d -name '*.app' -print -quit)
[[ -n "$app" ]]
"$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)/verify-app.sh" "$app" private-release
echo "GO_ADMIN_MACOS_DMG_VERIFY_PASS"
