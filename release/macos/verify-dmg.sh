#!/bin/sh
set -eu

if [ "$#" -ne 3 ]; then
  echo "usage: $0 IDENTITY_JSON DMG_PATH RELEASE_CLASS" >&2
  exit 2
fi

identity_file=$1
dmg_path=$2
release_class=$3
test -f "$dmg_path"
hdiutil verify "$dmg_path"

mountpoint=$(mktemp -d "${TMPDIR:-/tmp}/go-admin-mount.XXXXXX")
mounted=false
cleanup() {
  if [ "$mounted" = true ]; then
    hdiutil detach -quiet "$mountpoint" || true
  fi
  rmdir "$mountpoint" 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM

hdiutil attach -quiet -readonly -nobrowse -mountpoint "$mountpoint" "$dmg_path"
mounted=true
app_path=$(find "$mountpoint" -maxdepth 1 -type d -name '*.app' -print -quit)
test -n "$app_path"
"$(dirname "$0")/verify-app.sh" "$identity_file" "$app_path" "$release_class"

if [ "$release_class" = signed-production ]; then
  xcrun stapler validate "$dmg_path"
  spctl --assess --type open --context context:primary-signature -vv "$dmg_path"
  spctl --assess --type execute -vv "$app_path"
fi

echo "GO_ADMIN_MACOS_DMG_VERIFY_PASS"
