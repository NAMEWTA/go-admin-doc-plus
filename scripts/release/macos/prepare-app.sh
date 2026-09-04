#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 3 ]] || { echo "usage: prepare-app.sh <app> <version> <build-number>" >&2; exit 2; }
app=$1
version=$2
build_number=$3
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]
[[ "$build_number" =~ ^[1-9][0-9]*$ ]]
[[ -d "$app" && -f "$app/Contents/Info.plist" ]]

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
identity="$repository/release/macos/identity.json"
plist="$app/Contents/Info.plist"
product=$(jq -er '.productName' "$identity")
bundle=$(jq -er '.bundleIdentifier' "$identity")
minimum=$(jq -er '.minimumSystemVersion' "$identity")

set_plist() {
  local key=$1 type=$2 value=$3
  /usr/libexec/PlistBuddy -c "Set :$key $value" "$plist" 2>/dev/null || \
    /usr/libexec/PlistBuddy -c "Add :$key $type $value" "$plist"
}
set_plist CFBundleName string "$product"
set_plist CFBundleDisplayName string "$product"
set_plist CFBundleIdentifier string "$bundle"
set_plist CFBundleShortVersionString string "$version"
set_plist CFBundleVersion string "$build_number"
set_plist LSMinimumSystemVersion string "$minimum"
plutil -lint "$plist" >/dev/null

touch -t 202001010000 "$plist"
echo "GO_ADMIN_MACOS_APP_PREPARE_PASS"
