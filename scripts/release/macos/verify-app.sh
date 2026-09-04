#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 2 ]] || { echo "usage: verify-app.sh <app> <development|signed-production>" >&2; exit 2; }
app=$1
release_class=$2
[[ "$release_class" == development || "$release_class" == signed-production ]]
[[ -d "$app" ]]
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
identity="$repository/release/macos/identity.json"
plist="$app/Contents/Info.plist"
executable_name=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$plist")
host="$app/Contents/MacOS/$executable_name"
sidecar="$app/Contents/MacOS/go-admin-sidecar"

[[ -x "$host" && -x "$sidecar" ]]
test "$(/usr/libexec/PlistBuddy -c 'Print :CFBundleIdentifier' "$plist")" = "$(jq -er '.bundleIdentifier' "$identity")"
test "$(/usr/libexec/PlistBuddy -c 'Print :LSMinimumSystemVersion' "$plist")" = "$(jq -er '.minimumSystemVersion' "$identity")"
for binary in "$host" "$sidecar"; do
  test "$(lipo -archs "$binary")" = "x86_64 arm64"
done
node "$repository/go-admin-plus-ui/apps/admin-desktop/scripts/verify-production.mjs" --files "$host" "$sidecar"

if [[ "$release_class" == signed-production ]]; then
  codesign --verify --deep --strict --verbose=2 "$app"
  signature=$(codesign -dvvv "$app" 2>&1)
  grep -q 'Authority=Developer ID Application:' <<<"$signature"
  grep -Eq 'flags=.*runtime' <<<"$signature"
  ! codesign -d --entitlements :- "$app" 2>/dev/null | grep -q 'com.apple.security.get-task-allow'
else
  ! codesign -dvv "$app" 2>&1 | grep -q 'Authority=Developer ID Application:'
fi
echo "GO_ADMIN_MACOS_APP_VERIFY_PASS"
