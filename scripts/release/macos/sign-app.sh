#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 1 ]] || { echo "usage: sign-app.sh <app>" >&2; exit 2; }
app=$1
[[ -d "$app" ]]
[[ -n "${APPLE_SIGNING_IDENTITY:-}" ]] || { echo "Developer ID identity is required" >&2; exit 1; }
script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
repository=$(cd -- "$script_dir/../../.." && pwd -P)
entitlements="$repository/release/macos/entitlements.plist"
plutil -lint "$entitlements" >/dev/null

chmod -R u+w "$app/Contents/Resources/generator"
while IFS= read -r -d '' file_path; do
  if file -b "$file_path" | grep -q 'Mach-O'; then
    codesign --force --options runtime --timestamp --sign "$APPLE_SIGNING_IDENTITY" "$file_path"
  fi
done < <(find "$app/Contents" -type f -print0)
chmod -R a-w "$app/Contents/Resources/generator"
codesign --force --options runtime --timestamp --entitlements "$entitlements" \
  --sign "$APPLE_SIGNING_IDENTITY" "$app"
codesign --verify --deep --strict --verbose=2 "$app"
echo "GO_ADMIN_MACOS_APP_SIGN_PASS"
