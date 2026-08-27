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

generator="$app/Contents/Resources/generator"
for marker in repository/.git repository/scripts/contracts/cli.mjs repository/go-admin-plus/go.mod repository/go-admin-plus-ui/pnpm-workspace.yaml pnpm-store go-mod; do
  test -e "$generator/$marker"
done
test -z "$(git -C "$generator/repository" remote)"
for layout in darwin-amd64 darwin-arm64; do
  test -x "$generator/toolchains/$layout/go/bin/go"
  test -x "$generator/toolchains/$layout/node/bin/node"
  test -x "$generator/toolchains/$layout/bin/pnpm"
done
modules="$generator/repository/go-admin-plus-ui/node_modules/.pnpm"
test -d "$modules"
for pattern in '*android*' '*freebsd*' '*linux*' '*openbsd*' '*openharmony*' '*sunos*' '*win32*'; do
  ! compgen -G "$modules/$pattern" >/dev/null
done
compgen -G "$modules/@rolldown+binding-darwin-arm64@*" >/dev/null
compgen -G "$modules/@rolldown+binding-darwin-x64@*" >/dev/null
test "$(lipo -archs "$generator/toolchains/darwin-amd64/go/bin/go")" = x86_64
test "$(lipo -archs "$generator/toolchains/darwin-arm64/go/bin/go")" = arm64
test "$(lipo -archs "$generator/toolchains/darwin-amd64/node/bin/node")" = x86_64
test "$(lipo -archs "$generator/toolchains/darwin-arm64/node/bin/node")" = arm64

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
