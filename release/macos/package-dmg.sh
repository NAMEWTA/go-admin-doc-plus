#!/bin/sh
set -eu

if [ "$#" -ne 3 ]; then
  echo "usage: $0 APP_PATH OUTPUT_DMG VOLUME_NAME" >&2
  exit 2
fi

app_path=$1
output_dmg=$2
volume_name=$3
test -d "$app_path"

staging=$(mktemp -d "${TMPDIR:-/tmp}/go-admin-dmg.XXXXXX")
cleanup() {
  rm -rf "$staging"
}
trap cleanup EXIT HUP INT TERM

mkdir -p "$(dirname "$output_dmg")"
ditto "$app_path" "$staging/$(basename "$app_path")"
ln -s /Applications "$staging/Applications"
hdiutil create -quiet -ov -format UDZO -fs HFS+ \
  -volname "$volume_name" -srcfolder "$staging" "$output_dmg"
hdiutil verify "$output_dmg"
