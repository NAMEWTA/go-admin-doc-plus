#!/bin/sh
set -eu

if [ "$#" -ne 2 ]; then
  echo "usage: $0 IDENTITY_JSON APP_PATH" >&2
  exit 2
fi

identity_file=$1
source_app=$2
test -d "$source_app"
test "$(jq -er '.self_use_release_status' "$identity_file")" = approved

staging=$(mktemp -d "${TMPDIR:-/tmp}/go-admin-self-use.XXXXXX")
cleanup() {
  rm -rf "$staging"
}
trap cleanup EXIT HUP INT TERM

installed_app="$staging/Go Admin Plus.app"
ditto "$source_app" "$installed_app"
xattr -w com.apple.quarantine '0081;00000000;GoAdminPlus;self-use-verification' "$installed_app"
xattr -p com.apple.quarantine "$installed_app" >/dev/null
xattr -dr com.apple.quarantine "$installed_app"
if xattr -p com.apple.quarantine "$installed_app" >/dev/null 2>&1; then
  echo "scoped quarantine removal did not clear the app attribute" >&2
  exit 1
fi

"$(dirname "$0")/verify-app.sh" "$identity_file" "$installed_app" unsigned-self-use
set +e
spctl_output=$(spctl --assess --type execute -vv "$installed_app" 2>&1)
spctl_status=$?
set -e
printf '%s\n' "$spctl_output"
if [ "$spctl_status" -eq 0 ]; then
  echo "GO_ADMIN_MACOS_SPCTL_LOCAL_POLICY_ACCEPTED"
else
  echo "GO_ADMIN_MACOS_SPCTL_LOCAL_POLICY_REJECTED"
fi

echo "GO_ADMIN_MACOS_SELF_USE_INSTALL_PASS"
