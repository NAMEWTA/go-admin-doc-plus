#!/usr/bin/env bash
set -euo pipefail

[[ $# -eq 3 ]] || { echo "usage: notarize.sh <app|dmg> <artifact> <output-json>" >&2; exit 2; }
kind=$1
artifact=$2
output=$3
[[ "$kind" == app || "$kind" == dmg ]]
[[ -e "$artifact" && ! -e "$output" ]]
for key in APPLE_NOTARY_KEY_PATH APPLE_NOTARY_KEY_ID APPLE_NOTARY_ISSUER_ID; do
  [[ -n "${!key:-}" ]] || { echo "notary credential is required" >&2; exit 1; }
done

submission=$artifact
temporary=
if [[ "$kind" == app ]]; then
  temporary=$(mktemp "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/go-admin-app.XXXXXX.zip")
  ditto -c -k --keepParent "$artifact" "$temporary"
  submission=$temporary
fi
trap 'if [[ -n "$temporary" ]]; then rm -f -- "$temporary"; fi' EXIT
mkdir -p -- "$(dirname -- "$output")"
xcrun notarytool submit "$submission" \
  --key "$APPLE_NOTARY_KEY_PATH" \
  --key-id "$APPLE_NOTARY_KEY_ID" \
  --issuer "$APPLE_NOTARY_ISSUER_ID" \
  --wait --output-format json > "$output"
jq -e '.status == "Accepted" and (.id | type == "string" and length > 0)' "$output" >/dev/null
xcrun stapler staple "$artifact"
xcrun stapler validate "$artifact"
echo "GO_ADMIN_MACOS_NOTARY_PASS"
