#!/bin/sh
set -eu

if [ "$#" -ne 3 ]; then
  echo "usage: $0 IDENTITY_JSON APP_PATH RELEASE_CLASS" >&2
  exit 2
fi

identity_file=$1
app_path=$2
release_class=$3
plist=$app_path/Contents/Info.plist
test -f "$plist"

bundle_identifier=$(jq -er '.bundle_identifier' "$identity_file")
minimum_system_version=$(jq -er '.minimum_system_version' "$identity_file")
executable_name=$(/usr/libexec/PlistBuddy -c 'Print :CFBundleExecutable' "$plist")
executable=$app_path/Contents/MacOS/$executable_name

test -x "$executable"
test "$(/usr/libexec/PlistBuddy -c 'Print :CFBundleIdentifier' "$plist")" = "$bundle_identifier"
test "$(/usr/libexec/PlistBuddy -c 'Print :LSMinimumSystemVersion' "$plist")" = "$minimum_system_version"
test "$(lipo -archs "$executable")" = arm64
file "$executable" | grep -q 'arm64'

case "$release_class" in
  unsigned-self-use)
    codesign --verify --deep --strict "$app_path"
    signature=$(codesign -dvv "$app_path" 2>&1 || true)
    echo "$signature" | grep -q 'Signature=adhoc'
    echo "$signature" | grep -q 'TeamIdentifier=not set'
    if echo "$signature" | grep -q '^Authority='; then
      echo "self-use app unexpectedly contains a signing authority" >&2
      exit 1
    fi
    ;;
  signed-production)
    codesign --verify --deep --strict --verbose=2 "$app_path"
    signature=$(codesign -dvvv "$app_path" 2>&1)
    echo "$signature" | grep -q 'Authority=Developer ID Application:'
    echo "$signature" | grep -q 'flags=.*runtime'
    entitlements=$(codesign -d --entitlements :- "$app_path" 2>/dev/null || true)
    if echo "$entitlements" | grep -q 'com.apple.security.get-task-allow'; then
      echo "production signature contains forbidden get-task-allow entitlement" >&2
      exit 1
    fi
    ;;
  *) echo "unsupported release class: $release_class" >&2; exit 2 ;;
esac

echo "GO_ADMIN_MACOS_APP_VERIFY_PASS"
