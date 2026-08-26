#!/bin/sh
set -eu

if [ "$#" -ne 5 ]; then
  echo "usage: $0 IDENTITY_JSON APP_PATH VERSION BUILD_NUMBER RELEASE_CLASS" >&2
  exit 2
fi

identity_file=$1
app_path=$2
version=$3
build_number=$4
release_class=$5

if ! printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'; then
  echo "version must contain exactly three decimal components" >&2
  exit 2
fi
case "$build_number" in
  ''|0|*[!0-9]*) echo "build number must be a positive integer" >&2; exit 2 ;;
esac
case "$release_class" in
  unsigned-self-use|signed-production) ;;
  *) echo "unsupported release class: $release_class" >&2; exit 2 ;;
esac

test -f "$identity_file"
test -d "$app_path"
test -f "$app_path/Contents/Info.plist"

identity_status=$(jq -er '.identity_status | select(. == "candidate" or . == "approved")' "$identity_file")
self_use_release_status=$(jq -er '.self_use_release_status | select(. == "approved")' "$identity_file")
product_name=$(jq -er '.product_name | select(length > 0)' "$identity_file")
bundle_identifier=$(jq -er '.bundle_identifier | select(test("^[A-Za-z0-9.-]+$"))' "$identity_file")
minimum_system_version=$(jq -er '.minimum_system_version | select(test("^[0-9]+(\\.[0-9]+)+$"))' "$identity_file")

if [ "$release_class" = signed-production ] && [ "$identity_status" != approved ]; then
  echo "signed production requires identity_status=approved" >&2
  exit 1
fi
if [ "$release_class" = unsigned-self-use ] && [ "$self_use_release_status" != approved ]; then
  echo "unsigned self-use release requires self_use_release_status=approved" >&2
  exit 1
fi

plist=$app_path/Contents/Info.plist
set_plist() {
  key=$1
  type=$2
  value=$3
  /usr/libexec/PlistBuddy -c "Set :$key $value" "$plist" 2>/dev/null || \
    /usr/libexec/PlistBuddy -c "Add :$key $type $value" "$plist"
}

set_plist CFBundleName string "$product_name"
set_plist CFBundleDisplayName string "$product_name"
set_plist CFBundleIdentifier string "$bundle_identifier"
set_plist CFBundleShortVersionString string "$version"
set_plist CFBundleVersion string "$build_number"
set_plist LSMinimumSystemVersion string "$minimum_system_version"
plutil -lint "$plist"

test "$(/usr/libexec/PlistBuddy -c 'Print :CFBundleIdentifier' "$plist")" = "$bundle_identifier"
test "$(/usr/libexec/PlistBuddy -c 'Print :CFBundleShortVersionString' "$plist")" = "$version"
test "$(/usr/libexec/PlistBuddy -c 'Print :CFBundleVersion' "$plist")" = "$build_number"
