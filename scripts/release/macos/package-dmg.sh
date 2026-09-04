#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 2 ]] || { echo "usage: package-dmg.sh <app> <output-dmg>" >&2; exit 2; }
app=$1
output=$2
[[ -d "$app" && ! -e "$output" ]]
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
volume=$(jq -er '.productName' "$repository/release/macos/identity.json")
staging=$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/go-admin-macos-dmg.XXXXXX")
trap 'rm -rf -- "$staging"' EXIT
ditto "$app" "$staging/$(basename -- "$app")"
ln -s /Applications "$staging/Applications"
mkdir -p -- "$(dirname -- "$output")"
hdiutil create -quiet -format UDZO -fs HFS+ -volname "$volume" -srcfolder "$staging" "$output"
hdiutil verify "$output" >/dev/null
echo "GO_ADMIN_MACOS_DMG_PACKAGE_PASS"
